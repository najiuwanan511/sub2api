// Accept the upstream site or one of its OpenAI API URLs, including a reverse-proxy prefix.
// This is input normalization only; the server still validates all outbound destinations.
export function normalizeSub2APIUpstreamUrl(raw: string): string | null {
  const text = raw.trim()
  if (!/^https?:\/\/[^/]/i.test(text) || /[\s\\?#]/.test(text)) return null
  try {
    const url = new URL(text)
    if (!url.hostname || url.username || url.password) return null
    let path = url.pathname.replace(/\/+$/, '')
    path = path.replace(/\/v1\/(?:models|responses(?:\/compact)?|chat\/completions)$/, '/v1')
    if (!path.endsWith('/v1')) path += '/v1'
    return `${url.origin}${path}`
  } catch {
    return null
  }
}
