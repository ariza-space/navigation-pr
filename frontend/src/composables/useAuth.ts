import { ref } from 'vue'

import { APIError, getSession, login as loginRequest, logout as logoutRequest, updateAccount } from '@/lib/api'
import type { AccountInput, LoginInput, UserSession } from '@/types/api'

export function useAuth() {
  const user = ref<UserSession | null>(null)
  const loginOpen = ref(false)
  const loginError = ref('')

  function setAnonymous() {
    user.value = null
    loginOpen.value = false
  }

  async function refreshSession() {
    try {
      user.value = await getSession()
    } catch {
      user.value = null
    }
  }

  function requireLogin() {
    if (user.value) return true
    loginOpen.value = true
    return false
  }

  async function login(input: LoginInput) {
    loginError.value = ''
    try {
      user.value = await loginRequest(input)
      loginOpen.value = false
    } catch (error) {
      loginError.value = error instanceof Error ? error.message : '登录失败'
      throw error
    }
  }

  async function saveAccount(input: AccountInput) {
    user.value = await updateAccount(input)
  }

  async function logout() {
    await logoutRequest().catch(() => null)
    setAnonymous()
  }

  function handleAuthError(error: unknown) {
    if (error instanceof APIError && error.status === 401) {
      user.value = null
      loginOpen.value = true
      return true
    }
    return false
  }

  return {
    user,
    loginOpen,
    loginError,
    setAnonymous,
    refreshSession,
    requireLogin,
    login,
    saveAccount,
    logout,
    handleAuthError,
  }
}
