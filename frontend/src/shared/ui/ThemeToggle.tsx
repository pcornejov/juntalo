import { useState } from 'react'
import { Moon, Sun } from 'lucide-react'
import { applyTheme, getStoredTheme } from '../lib/theme'

export function ThemeToggle() {
  const [theme, setTheme] = useState<'light' | 'dark'>(getStoredTheme)

  function toggle() {
    const next = theme === 'dark' ? 'light' : 'dark'
    setTheme(next)
    applyTheme(next)
  }

  return (
    <button
      type="button"
      onClick={toggle}
      aria-label={theme === 'dark' ? 'Cambiar a modo claro' : 'Cambiar a modo oscuro'}
      className="flex h-8 w-8 items-center justify-center rounded-full text-text-secondary hover:bg-bg-subtle hover:text-text-primary"
    >
      {theme === 'dark' ? <Sun className="h-4 w-4" strokeWidth={1.75} /> : <Moon className="h-4 w-4" strokeWidth={1.75} />}
    </button>
  )
}
