import { ApiError } from './client'

// El backend ya manda mensajes en español (Etapa 4 §1); este mapa solo pisa
// los casos donde conviene un texto más orientado a la UI que el genérico del API.
const codeMessages: Record<string, string> = {
  invalid_credentials: 'Email o contraseña incorrectos',
  email_already_registered: 'Ya existe una cuenta con este email',
  weak_password: 'La contraseña debe tener al menos 8 caracteres',
  validation_failed: 'Revisa los datos ingresados',
  session_expired: 'Tu sesión expiró, inicia sesión nuevamente',
  unauthorized: 'Debes iniciar sesión para continuar',
}

export function errorMessage(err: unknown): string {
  if (err instanceof ApiError) {
    return codeMessages[err.code] ?? err.message
  }
  return 'Ocurrió un error inesperado'
}
