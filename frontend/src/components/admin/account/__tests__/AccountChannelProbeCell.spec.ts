import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { AccountListItem } from '@/types'
import AccountChannelProbeCell from '../AccountChannelProbeCell.vue'

const { probeChannel } = vi.hoisted(() => ({ probeChannel: vi.fn() }))
vi.mock('@/api/admin', () => ({ adminAPI: { accounts: { probeChannel } } }))
vi.mock('@/utils/format', () => ({ formatDateTime: (stamp: string) => stamp }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string, params?: { ms?: number }) => params?.ms === undefined ? key : `${key} ${params.ms} ms` }) }))

const account = { id: 7, type: 'apikey', credentials: { base_url: 'https://upstream.example.com/v1' } } as AccountListItem
const checked_at = '2026-11-25T05:00:00Z'

describe('AccountChannelProbeCell', () => {
  beforeEach(() => probeChannel.mockReset())

  it('does not run automatically and treats a 401 as reachable without claiming authentication', async () => {
    probeChannel.mockResolvedValueOnce({ status: 'reachable', latency_ms: 123, http_status: 401, checked_at })
    const wrapper = mount(AccountChannelProbeCell, { props: { account } })
    expect(probeChannel).not.toHaveBeenCalled()
    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(probeChannel).toHaveBeenCalledTimes(1)
    expect(probeChannel).toHaveBeenCalledWith(7)
    expect(wrapper.text()).toContain('reachable 123 ms')
    expect(wrapper.text()).toContain('HTTP 401')
    await wrapper.setProps({ account: { ...account, name: 'Updated row' } })
    expect(wrapper.text()).toContain('HTTP 401')
  })

  it('shows timeout and resets when the endpoint changes', async () => {
    probeChannel.mockResolvedValueOnce({ status: 'timeout', checked_at })
    const wrapper = mount(AccountChannelProbeCell, { props: { account } })
    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('channelProbe.timeout')
    expect(wrapper.text()).not.toContain('HTTP')
    await wrapper.setProps({ account: { ...account, credentials: { base_url: 'https://other.example.com' } } })
    expect(wrapper.text()).toContain('channelProbe.run')
    expect(wrapper.text()).not.toContain('channelProbe.timeout')
  })

  it('does not display a probe for accounts without a configured upstream', () => {
    const wrapper = mount(AccountChannelProbeCell, { props: { account: { ...account, credentials: {} } } })
    expect(wrapper.find('button').exists()).toBe(false)
  })
})
