import { handleUpload, type HandleUploadBody } from '@vercel/blob/client'
import { createHmac, timingSafeEqual } from 'node:crypto'

const uploadRoles = new Set(['admin', 'manager', 'teacher'])

function base64UrlToBuffer(value: string) {
  const normalized = value.replace(/-/g, '+').replace(/_/g, '/')
  const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=')
  return Buffer.from(padded, 'base64')
}

function verifyJwt(token: string) {
  try {
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
  } catch {
    return null
  }
}

function bearerToken(request: { headers: Record<string, string | string[] | undefined> }) {
  const rawHeader = request.headers.authorization || request.headers.Authorization
  const header = Array.isArray(rawHeader) ? rawHeader[0] : rawHeader || ''
  if (!header.toLowerCase().startsWith('bearer ')) return ''
  return header.slice(7).trim()
}

function sendJson(response: any, status: number, data: unknown) {
  response.status(status).json(data)
}

async function readJsonBody(request: any) {
  if (request.body && typeof request.body === 'object') return request.body
  if (typeof request.body === 'string') return JSON.parse(request.body)

  const chunks: Buffer[] = []
  for await (const chunk of request) {
    chunks.push(Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk))
  }
  const rawBody = Buffer.concat(chunks).toString('utf8')
  return rawBody ? JSON.parse(rawBody) : {}
}

export default async function handler(request: any, response: any) {
  try {
    if (request.method !== 'POST') {
      return sendJson(response, 405, { error: 'method not allowed' })
    }
    if (!process.env.BLOB_READ_WRITE_TOKEN) {
      return sendJson(response, 500, { error: 'BLOB_READ_WRITE_TOKEN is not configured' })
    }

    const body = (await readJsonBody(request)) as HandleUploadBody
    const isTokenRequest = body.type === 'blob.generate-client-token'
    const claims = verifyJwt(bearerToken(request))
    
    // Always verify token for security
    if (!claims) {
      return sendJson(response, 401, { error: 'unauthorized' })
    }

    const uploadResponse = await handleUpload({
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
    return sendJson(response, 200, uploadResponse)
  } catch (error) {
    const message = error instanceof Error ? error.message : 'upload failed'
    return sendJson(response, 400, { error: message })
  }
}
