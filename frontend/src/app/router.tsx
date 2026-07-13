import { createBrowserRouter } from 'react-router-dom'
import { ComponentsShowcasePage } from './ComponentsShowcasePage'

export const router = createBrowserRouter([
  {
    path: '/',
    element: <ComponentsShowcasePage />,
  },
  // Rutas de features (auth, campaigns, public-campaign) se montan en los hitos siguientes.
])
