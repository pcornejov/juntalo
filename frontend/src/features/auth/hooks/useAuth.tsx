import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import * as authApi from '../api'
import { setAccessToken } from '../../../shared/api/client'

interface AuthState {
  user: authApi.User | null
  organization: authApi.Organization | null
  isLoading: boolean
}

interface AuthContextValue extends AuthState {
  login: (email: string, password: string) => Promise<void>
  register: (email: string, fullName: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({ user: null, organization: null, isLoading: true })

  // Al montar, intenta un refresh silencioso: si hay cookie de sesión válida,
  // la app arranca ya logueada sin que el usuario haga nada (Hito 1).
  useEffect(() => {
    let cancelled = false
    async function bootstrap() {
      try {
        const { access_token } = await authApi.refresh()
        setAccessToken(access_token)
        const me = await authApi.me()
        if (!cancelled) {
          setState({ user: me.user, organization: me.organization, isLoading: false })
        }
      } catch {
        setAccessToken(null)
        if (!cancelled) {
          setState({ user: null, organization: null, isLoading: false })
        }
      }
    }
    bootstrap()
    return () => {
      cancelled = true
    }
  }, [])

  async function login(email: string, password: string) {
    const res = await authApi.login({ email, password })
    setAccessToken(res.access_token)
    setState({ user: res.user, organization: res.organization, isLoading: false })
  }

  async function register(email: string, fullName: string, password: string) {
    const res = await authApi.register({ email, full_name: fullName, password })
    setAccessToken(res.access_token)
    setState({ user: res.user, organization: res.organization, isLoading: false })
  }

  async function logout() {
    await authApi.logout().catch(() => undefined)
    setAccessToken(null)
    setState({ user: null, organization: null, isLoading: false })
  }

  return (
    <AuthContext.Provider value={{ ...state, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error('useAuth debe usarse dentro de <AuthProvider>')
  }
  return ctx
}
