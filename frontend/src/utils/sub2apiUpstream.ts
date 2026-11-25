export type Sub2APIUpstreamPlatform = 'openai' | 'anthropic' | 'grok'

// Accept the upstream site or a common API URL, including a reverse-proxy prefix.
// This is input normalization only; the server still validates all outbound destinations.
export function normalizeSub2APIUpstreamUrl(raw: string, platform: Sub2APIUpstreamPlatform = 'openai'): string | null {
  const text = raw.trim()
  if (!/^https?:\/\/[^/]/i.test(text) || /[\s\\?#]/.test(text)) return null
  try {
    const url = new URL(text)
    if (!url.hostname || url.username || url.password) return null
    let path = url.pathname.replace(/\/+$/, '')
    path = path.replace(/\/v1\/(?:models|messages(?:\/count_tokens)?|responses(?:\/compact)?|chat\/completions)$/, '/v1')
    // Claude's backend appends /v1/messages; OpenAI and Grok use a /v1 base.
    if (platform === 'anthropic') path = path.replace(/\/v1$/, '')
    else if (!path.endsWith('/v1')) path += '/v1'
    return `${url.origin}${path}`
  } catch {
    return null
  }
}
