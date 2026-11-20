import { defineStore } from 'pinia'
import { shallowRef } from 'vue'

export type GatewaySearchTarget = '/admin/accounts' | '/admin/users' | '/keys'

export const useGatewaySearchStore = defineStore('gatewaySearch', () => {
  const pending = shallowRef<{ target: GatewaySearchTarget; term: string } | null>(null)
  function submit(target: GatewaySearchTarget, term: string) {
    pending.value = { target, term }
  }
  function consume(target: GatewaySearchTarget): string | undefined {
    if (pending.value?.target !== target) return
    const term = pending.value.term
    pending.value = null
    return term
  }
  return { pending, submit, consume }
})
