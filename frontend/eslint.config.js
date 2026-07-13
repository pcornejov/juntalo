import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import boundaries from 'eslint-plugin-boundaries'
import tseslint from 'typescript-eslint'

export default tseslint.config(
  { ignores: ['dist'] },
  {
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      ecmaVersion: 2022,
      globals: globals.browser,
    },
    plugins: {
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
      boundaries,
    },
    settings: {
      'boundaries/elements': [
        { type: 'shared', pattern: 'src/shared/*' },
        { type: 'features', pattern: 'src/features/*' },
        { type: 'app', pattern: 'src/app/*' },
      ],
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
      // Etapa 5 §4: shared no importa features; features no se importan entre sí; app puede todo.
      'boundaries/dependencies': [
        'error',
        {
          default: 'disallow',
          policies: [
            { from: ['shared'], allow: ['shared'] },
            { from: ['features'], allow: ['shared'] },
            { from: ['app'], allow: ['shared', 'features', 'app'] },
          ],
        },
      ],
    },
  },
)
