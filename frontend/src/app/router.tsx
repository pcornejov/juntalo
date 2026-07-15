import { createBrowserRouter } from 'react-router-dom'
import { ComponentsShowcasePage } from './ComponentsShowcasePage'
import { DashboardLayout } from './DashboardLayout'
import { RequireAdmin, RequireAuth } from './guard'
import { HowItWorksPage } from './HowItWorksPage'
import { LandingPage } from './LandingPage'
import { NotFoundPage } from './NotFoundPage'
import { PrivacyPage } from './PrivacyPage'
import { TermsPage } from './TermsPage'
import { RedirectToPublicCampaign } from './RedirectToPublicCampaign'
import { ForgotPasswordPage } from '../features/auth/pages/ForgotPasswordPage'
import { LoginPage } from '../features/auth/pages/LoginPage'
import { RegisterPage } from '../features/auth/pages/RegisterPage'
import { ResetPasswordPage } from '../features/auth/pages/ResetPasswordPage'
import { VerifyEmailPage } from '../features/auth/pages/VerifyEmailPage'
import { CampaignListPage } from '../features/campaigns/pages/CampaignListPage'
import { CampaignFormPage } from '../features/campaigns/pages/CampaignFormPage'
import { CampaignDetailPage } from '../features/campaigns/pages/CampaignDetailPage'
import { ExploreCampaignsPage } from '../features/public-campaign/pages/ExploreCampaignsPage'
import { OrgProfilePage } from '../features/public-campaign/pages/OrgProfilePage'
import { PublicCampaignPage } from '../features/public-campaign/pages/PublicCampaignPage'
import { BlogIndexPage } from '../features/blog/pages/BlogIndexPage'
import { BlogArticlePage } from '../features/blog/pages/BlogArticlePage'
import { BackofficePage } from '../features/admin/pages/BackofficePage'
import { PayoutSettingsPage } from '../features/organizations/pages/PayoutSettingsPage'

export const router = createBrowserRouter([
  { path: '/', element: <LandingPage /> },
  { path: '/login', element: <LoginPage /> },
  { path: '/register', element: <RegisterPage /> },
  { path: '/forgot-password', element: <ForgotPasswordPage /> },
  { path: '/reset-password', element: <ResetPasswordPage /> },
  { path: '/verify-email', element: <VerifyEmailPage /> },
  { path: '/design', element: <ComponentsShowcasePage /> },
  { path: '/privacidad', element: <PrivacyPage /> },
  { path: '/bases', element: <TermsPage /> },
  { path: '/como-funciona', element: <HowItWorksPage /> },
  { path: '/explorar', element: <ExploreCampaignsPage /> },
  { path: '/org/:slug', element: <OrgProfilePage /> },
  { path: '/consejos', element: <BlogIndexPage /> },
  { path: '/consejos/:slug', element: <BlogArticlePage /> },
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
          { path: 'payout', element: <PayoutSettingsPage /> },
          {
            element: <RequireAdmin />,
            children: [{ path: 'backoffice', element: <BackofficePage /> }],
          },
        ],
      },
    ],
  },
])
