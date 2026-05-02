import { handleUpload, type HandleUploadBody } from '@vercel/blob/client'
import { createHmac, timingSafeEqual } from 'node:crypto'

const uploadRoles = new Set(['admin', 'manager', 'teacher'])

function base64UrlToBuffer(value: string) {
  const normalized = value.replace(/-/g, '+').replace(/_/g, '/')
  const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=')
  return Buffer.from(padded, 'base64')
}

function verifyJwt(token: string) {
  const [header, payload, signature] = token.split('.')
  if (!header || !payload || !signature) return null

  const secret = process.env.EDUPULSE_JWT_SECRET || process.env.JWT_SECRET || 'dev-secret-change-me'
  const expected = createHmac('sha256', secret).update(`${header}.${payload}`).digest()
  const actual = base64UrlToBuffer(signature)
  if (expected.length !== actual.length || !timingSafeEqual(expected, actual)) return null

  const claims = JSON.parse(base64UrlToBuffer(payload).toString('utf8')) as { exp?: number; role?: string; uid?: number }
  if (!claims.exp || claims.exp * 1000 < Date.now()) return null
  if (!claims.role || !uploadRoles.has(claims.role)) return null
  return claims
}

function bearerToken(request: Request) {
  const header = request.headers.get('authorization') || ''
  if (!header.toLowerCase().startsWith('bearer ')) return ''
  return header.slice(7).trim()
}

export default async function handler(request: Request) {
  if (request.method !== 'POST') {
    return Response.json({ error: 'method not allowed' }, { status: 405 })
  }

  const body = (await request.json()) as HandleUploadBody
  const isTokenRequest = body.type === 'blob.generate-client-token'
  const claims = isTokenRequest ? verifyJwt(bearerToken(request)) : null
  if (isTokenRequest && !claims) {
    return Response.json({ error: 'unauthorized' }, { status: 401 })
  }

  try {
    const response = await handleUpload({
      body,
      request,
      onBeforeGenerateToken: async (pathname) => {
        const isVideo = pathname.startsWith('edupulse/videos/')
        return {
          allowedContentTypes: isVideo
            ? ['video/mp4', 'video/quicktime', 'video/x-matroska']
            : [
                'application/pdf',
                'image/*',
                'text/*',
                'application/msword',
                'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
                'application/vnd.openxmlformats-officedocument.presentationml.presentation',
                'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
              ],
          maximumSizeInBytes: isVideo ? 500 * 1024 * 1024 : 50 * 1024 * 1024,
          addRandomSuffix: true,
          tokenPayload: JSON.stringify({ uid: claims?.uid, role: claims?.role }),
        }
      },
      onUploadCompleted: async () => {},
    })
    return Response.json(response)
  } catch (error) {
    return Response.json({ error: error instanceof Error ? error.message : 'upload failed' }, { status: 400 })
  }
}
