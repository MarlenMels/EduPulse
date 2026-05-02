import { del } from '@vercel/blob'
import { createHmac, timingSafeEqual } from 'node:crypto'

function base64UrlToBuffer(value: string) {
  const normalized = value.replace(/-/g, '+').replace(/_/g, '/')
  const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=')
  return Buffer.from(padded, 'base64')
}

function verifyAdmin(token: string) {
  const [header, payload, signature] = token.split('.')
  if (!header || !payload || !signature) return false

  const secret = process.env.EDUPULSE_JWT_SECRET || process.env.JWT_SECRET || 'dev-secret-change-me'
  const expected = createHmac('sha256', secret).update(`${header}.${payload}`).digest()
  const actual = base64UrlToBuffer(signature)
  if (expected.length !== actual.length || !timingSafeEqual(expected, actual)) return false

  const claims = JSON.parse(base64UrlToBuffer(payload).toString('utf8')) as { exp?: number; role?: string }
  return Boolean(claims.exp && claims.exp * 1000 > Date.now() && claims.role === 'admin')
}

function bearerToken(request: Request) {
  const header = request.headers.get('authorization') || ''
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

export default async function handler(request: Request) {
  if (request.method !== 'DELETE') {
    return Response.json({ error: 'method not allowed' }, { status: 405 })
  }
  if (!verifyAdmin(bearerToken(request))) {
    return Response.json({ error: 'admin only' }, { status: 403 })
  }

  const body = (await request.json()) as { url?: string }
  if (!body.url || !isEduPulseBlobURL(body.url)) {
    return Response.json({ error: 'invalid blob url' }, { status: 400 })
  }

  await del(body.url)
  return Response.json({ status: 'deleted' })
}
