import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

export function useGatewayAdminLayout() {
  const route = useRoute()
  const auth = useAuthStore()
  return computed(() => auth.isAdmin && (
    route.path === '/admin' || route.path.startsWith('/admin/') || route.path === '/keys'
  ))
}
