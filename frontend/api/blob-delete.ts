import { del } from '@vercel/blob'
import { createHmac, timingSafeEqual } from 'node:crypto'

function base64UrlToBuffer(value: string) {
  const normalized = value.replace(/-/g, '+').replace(/_/g, '/')
  const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=')
  return Buffer.from(padded, 'base64')
}

function verifyAdmin(token: string) {
  try {
    const [header, payload, signature] = token.split('.')
    if (!header || !payload || !signature) return false

    const secret = process.env.EDUPULSE_JWT_SECRET || process.env.JWT_SECRET || 'dev-secret-change-me'
    const expected = createHmac('sha256', secret).update(`${header}.${payload}`).digest()
    const actual = base64UrlToBuffer(signature)
    if (expected.length !== actual.length || !timingSafeEqual(expected, actual)) return false

    const claims = JSON.parse(base64UrlToBuffer(payload).toString('utf8')) as { exp?: number; role?: string }
    return Boolean(claims.exp && claims.exp * 1000 > Date.now() && claims.role === 'admin')
  } catch {
    return false
  }
}

function bearerToken(request: { headers: Record<string, string | string[] | undefined> }) {
  const rawHeader = request.headers.authorization || request.headers.Authorization
  const header = Array.isArray(rawHeader) ? rawHeader[0] : rawHeader || ''
  if (!header.toLowerCase().startsWith('bearer ')) return ''
  return header.slice(7).trim()
}

function isEduPulseBlobURL(value: string) {
  try {
    const url = new URL(value)
    return url.hostname.endsWith('.vercel-storage.com') &&
      (url.pathname.includes('/edupulse/videos/') || url.pathname.includes('/edupulse/materials/'))
  } catch {
    return false
  }
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
    if (request.method !== 'DELETE') {
      return sendJson(response, 405, { error: 'method not allowed' })
    }
    if (!process.env.BLOB_READ_WRITE_TOKEN) {
      return sendJson(response, 500, { error: 'BLOB_READ_WRITE_TOKEN is not configured' })
    }
    if (!verifyAdmin(bearerToken(request))) {
      return sendJson(response, 403, { error: 'admin only' })
    }

    const body = (await readJsonBody(request)) as { url?: string }
    if (!body.url || !isEduPulseBlobURL(body.url)) {
      return sendJson(response, 400, { error: 'invalid blob url' })
    }

    await del(body.url)
    return sendJson(response, 200, { status: 'deleted' })
  } catch (error) {
    return sendJson(response, 400, { error: error instanceof Error ? error.message : 'delete failed' })
  }
}
