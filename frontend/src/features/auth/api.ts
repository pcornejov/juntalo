import { apiClient } from '../../shared/api/client'

export interface User {
  id: string
  email: string
  full_name: string
  email_verified: boolean
  is_admin: boolean
}

export interface Organization {
  id: string
  name: string
  slug: string
  kind: string
}

export interface AuthResponse {
  user: User
  organization: Organization
  access_token: string
}

export interface MeResponse {
  user: User
  organization: Organization
}

export function register(input: {
  email: string
  full_name: string
  password: string
  captcha_token?: string
}) {
  return apiClient.post<AuthResponse>('/auth/register', input, { skipAuth: true })
}

export function login(input: { email: string; password: string }) {
  return apiClient.post<AuthResponse>('/auth/login', input, { skipAuth: true })
}

export function refresh() {
  return apiClient.post<{ access_token: string }>('/auth/refresh', undefined, { skipAuth: true })
}

export function logout() {
  return apiClient.post<void>('/auth/logout')
}

export function me() {
  return apiClient.get<MeResponse>('/auth/me')
}

export function forgotPassword(email: string) {
  return apiClient.post<{ message: string; reset_token?: string }>(
    '/auth/forgot-password',
    { email },
    { skipAuth: true },
  )
}

export function resetPassword(token: string, newPassword: string) {
  return apiClient.post<void>(
    '/auth/reset-password',
    { token, new_password: newPassword },
    { skipAuth: true },
  )
}

export function verifyEmail(token: string) {
  return apiClient.post<void>('/auth/verify-email', { token }, { skipAuth: true })
}

export function resendVerification() {
  return apiClient.post<{ message: string }>('/auth/resend-verification')
}
