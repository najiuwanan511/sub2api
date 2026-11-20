import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import WalletResetCardsView from '../WalletResetCardsView.vue'

const { previewResetCard, applyResetCard, getResetCredits } = vi.hoisted(() => ({
  previewResetCard: vi.fn(), applyResetCard: vi.fn(), getResetCredits: vi.fn(),
}))
vi.mock('@/api/redeem', () => ({ redeemAPI: { previewResetCard, applyResetCard, getResetCredits } }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ user: { balance: 10 } }) }))
vi.mock('@/utils/format', () => ({ formatDateTime: (value: string) => value }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('@/components/layout/AppLayout.vue', () => ({ default: { template: '<main><slot /></main>' } }))

const credit = {
  redeem_code_id: 1, group_id: 4, group_name: 'Wallet', amount_usd: 3,
  remaining_usd: 2, expires_at: '2026-11-27T00:00:00Z', status: 'active',
}

function mountPage() {
  return mount(WalletResetCardsView, { global: { stubs: { RouterLink: true } } })
}

describe('Wallet reset cards', () => {
  it('previews a code, confirms redemption, and reloads server totals and usage', async () => {
    getResetCredits.mockResolvedValueOnce({ active_usd: 0, credits: [] })
      .mockResolvedValueOnce({ active_usd: 4, credits: [credit] })
    previewResetCard.mockResolvedValueOnce({ group_name: 'Wallet', percent: 40, credit_usd: 3, eligible_usd: 10, validity_days: 7 })
    applyResetCard.mockResolvedValueOnce(credit)
    const wrapper = mountPage()
    await flushPromises()
    expect(wrapper.text()).toContain('walletReset.empty')
    await wrapper.get('input').setValue('RESET-CODE')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(previewResetCard).toHaveBeenCalledWith('RESET-CODE')
    expect(applyResetCard).not.toHaveBeenCalled()
    await wrapper.get('button.btn-primary').trigger('click')
    await flushPromises()
    expect(applyResetCard).toHaveBeenCalledWith('RESET-CODE')
    expect(wrapper.text()).toContain('walletReset.applied')
    expect(wrapper.text()).toContain('$4.00')
    expect(wrapper.get('progress').attributes('value')).toBe('1')
    expect(wrapper.get('progress').attributes('max')).toBe('3')
    wrapper.unmount()
  })

  it('does not allow a preview with no eligible spend to be redeemed', async () => {
    applyResetCard.mockClear()
    getResetCredits.mockResolvedValueOnce({ active_usd: 0, credits: [] })
    previewResetCard.mockResolvedValueOnce({ group_name: 'Wallet', percent: 40, credit_usd: 0, eligible_usd: 0, validity_days: 7 })
    const wrapper = mountPage()
    await flushPromises()
    await wrapper.get('input').setValue('NO-SPEND')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(wrapper.get('button.btn-primary').attributes('disabled')).toBeDefined()
    expect(applyResetCard).not.toHaveBeenCalled()
    wrapper.unmount()
  })
})
