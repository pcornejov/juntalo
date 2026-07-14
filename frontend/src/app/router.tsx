import { createBrowserRouter } from 'react-router-dom'
import { ComponentsShowcasePage } from './ComponentsShowcasePage'
import { DashboardLayout } from './DashboardLayout'
import { RequireAuth } from './guard'
import { LandingPage } from './LandingPage'
import { NotFoundPage } from './NotFoundPage'
import { PrivacyPage } from './PrivacyPage'
import { RedirectToPublicCampaign } from './RedirectToPublicCampaign'
import { ForgotPasswordPage } from '../features/auth/pages/ForgotPasswordPage'
import { LoginPage } from '../features/auth/pages/LoginPage'
import { RegisterPage } from '../features/auth/pages/RegisterPage'
import { ResetPasswordPage } from '../features/auth/pages/ResetPasswordPage'
import { VerifyEmailPage } from '../features/auth/pages/VerifyEmailPage'
import { CampaignListPage } from '../features/campaigns/pages/CampaignListPage'
import { CampaignFormPage } from '../features/campaigns/pages/CampaignFormPage'
import { CampaignDetailPage } from '../features/campaigns/pages/CampaignDetailPage'
import { PublicCampaignPage } from '../features/public-campaign/pages/PublicCampaignPage'

export const router = createBrowserRouter([
  { path: '/', element: <LandingPage /> },
  { path: '/login', element: <LoginPage /> },
  { path: '/register', element: <RegisterPage /> },
  { path: '/forgot-password', element: <ForgotPasswordPage /> },
  { path: '/reset-password', element: <ResetPasswordPage /> },
  { path: '/verify-email', element: <VerifyEmailPage /> },
  { path: '/design', element: <ComponentsShowcasePage /> },
  { path: '/privacidad', element: <PrivacyPage /> },
  { path: '/public/:slug', element: <PublicCampaignPage /> },
  { path: '/c/:slug', element: <RedirectToPublicCampaign /> },
  { path: '*', element: <NotFoundPage /> },
  {
    element: <RequireAuth />,
    children: [
      {
        path: '/dashboard',
        element: <DashboardLayout />,
        children: [
          { index: true, element: <CampaignListPage /> },
          { path: 'campaigns/new', element: <CampaignFormPage /> },
          { path: 'campaigns/:id', element: <CampaignDetailPage /> },
        ],
      },
    ],
  },
])
