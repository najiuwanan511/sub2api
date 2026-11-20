import { describe, expect, it } from 'vitest'
import { normalizeSub2APIUpstreamUrl } from '../sub2apiUpstream'

describe('Sub2API upstream URL', () => {
  it.each([
    ['https://upstream.example', 'https://upstream.example/v1'],
    [' https://upstream.example/v1/ ', 'https://upstream.example/v1'],
    ['https://upstream.example/api/v1/responses/compact/', 'https://upstream.example/api/v1'],
    ['https://upstream.example/api/v1/chat/completions', 'https://upstream.example/api/v1'],
    ['https://upstream.example:8443/v1/models', 'https://upstream.example:8443/v1'],
  ])('normalizes %s', (input, expected) => {
    expect(normalizeSub2APIUpstreamUrl(input)).toBe(expected)
  })

  it.each(['', 'upstream.example', 'ftp://upstream.example', 'https:///upstream.example',
    'https://user@upstream.example', 'https://upstream.example/v1?', 'https://upstream.example/#section',
  ])('rejects %s', input => {
    expect(normalizeSub2APIUpstreamUrl(input)).toBeNull()
  })
})
