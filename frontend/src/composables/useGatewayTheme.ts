import { ref } from 'vue'

export function useGatewayTheme() {
  const isDark = ref(document.documentElement.classList.contains('dark'))
  function toggleTheme() {
    isDark.value = !document.documentElement.classList.contains('dark')
    document.documentElement.classList.toggle('dark', isDark.value)
    localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
  }
  return { isDark, toggleTheme }
}
