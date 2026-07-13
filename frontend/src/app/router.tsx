import { createBrowserRouter, Navigate } from 'react-router-dom'
import { ComponentsShowcasePage } from './ComponentsShowcasePage'
import { DashboardLayout } from './DashboardLayout'
import { RequireAuth } from './guard'
import { LoginPage } from '../features/auth/pages/LoginPage'
import { RegisterPage } from '../features/auth/pages/RegisterPage'
import { CampaignListPage } from '../features/campaigns/pages/CampaignListPage'

export const router = createBrowserRouter([
  { path: '/', element: <Navigate to="/dashboard" replace /> },
  { path: '/login', element: <LoginPage /> },
  { path: '/register', element: <RegisterPage /> },
  { path: '/design', element: <ComponentsShowcasePage /> },
  {
    element: <RequireAuth />,
    children: [
      {
        path: '/dashboard',
        element: <DashboardLayout />,
        children: [{ index: true, element: <CampaignListPage /> }],
      },
    ],
  },
])
