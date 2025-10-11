import { Routes, Route } from 'react-router-dom'
import { AuthProvider } from './contexts/AuthContext'
import { ToastProvider } from './contexts/ToastContext'
import Layout from './components/Layout'
import HomePage from './pages/HomePage'
import RegisterPage from './pages/RegisterPage'
import LoginPage from './pages/LoginPage'
import VolunteerDashboard from './pages/VolunteerDashboard'
import AdminDashboard from './pages/AdminDashboard'
import VerifyEmailPage from './pages/VerifyEmailPage'
import CreateInitiativePage from './pages/admin/CreateInitiativePage'
import InitiativesListPage from './pages/admin/InitiativesListPage'
import ApplicationsPage from './pages/admin/ApplicationsPage'
import VolunteerPortal from './pages/VolunteerPortal'
import VolunteerProfilePage from './pages/VolunteerProfilePage'
import InitiativeDetailPage from './pages/InitiativeDetailPage'
import SkillManagementPage from './pages/admin/SkillManagementPage'
import AdminProfilePage from './pages/admin/AdminProfilePage'
// Project pages
import ProjectsListPage from './pages/projects/ProjectsListPage'
import ProjectDetailPage from './pages/projects/ProjectDetailPage'
import CreateProjectPage from './pages/projects/CreateProjectPage'
// Volunteer pages
import VolunteersPoolPage from './pages/volunteers/VolunteersPoolPage'
import VolunteerScorecardPage from './pages/volunteers/VolunteerScorecardPage'
// Campaign pages
import CampaignsListPage from './pages/campaigns/CampaignsListPage'
import CreateCampaignPage from './pages/campaigns/CreateCampaignPage'
// Admin pages
import UserManagementPage from './pages/admin/UserManagementPage'
import RoleManagementPage from './pages/admin/RoleManagementPage'
import ProtectedRoute from './components/ProtectedRoute'
import Toast from './components/Toast'

function App() {
  return (
    <ToastProvider>
      <AuthProvider>
        <div className="min-h-screen bg-secondary-50">
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route index element={<HomePage />} />
              <Route path="register" element={<RegisterPage />} />
              <Route path="login" element={<LoginPage />} />
              <Route path="verify-email" element={<VerifyEmailPage />} />
              <Route 
                path="dashboard" 
                element={
                  <ProtectedRoute>
                    <VolunteerDashboard />
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="admin" 
                element={
                  <ProtectedRoute requiredRole="admin">
                    <AdminDashboard />
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="admin/initiatives" 
                element={
                  <ProtectedRoute requiredRole="admin">
                    <InitiativesListPage />
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="admin/initiatives/create" 
                element={
                  <ProtectedRoute requiredRole="admin">
                    <CreateInitiativePage />
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="admin/applications" 
                element={
                  <ProtectedRoute requiredRole="admin">
                    <ApplicationsPage />
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="admin/skills" 
                element={
                  <ProtectedRoute requiredRole="admin">
                    <SkillManagementPage />
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="admin/profile" 
                element={
                  <ProtectedRoute requiredRole="admin">
                    <AdminProfilePage />
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="volunteer" 
                element={
                  <ProtectedRoute requiredRole="volunteer">
                    <VolunteerPortal />
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="profile" 
                element={
                  <ProtectedRoute requiredRole="volunteer">
                    <VolunteerProfilePage />
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="initiatives/:id" 
                element={
                  <ProtectedRoute requiredRole="volunteer">
                    <InitiativeDetailPage />
                  </ProtectedRoute>
                } 
              />
              {/* Project routes */}
              <Route 
                path="projects" 
                element={
                  <ProtectedRoute>
                    <ProjectsListPage />
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="projects/create" 
                element={
                  <ProtectedRoute requiredRoles={['team_lead', 'admin']}>
                    <CreateProjectPage />
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="projects/:id" 
                element={
                  <ProtectedRoute>
                    <ProjectDetailPage />
                  </ProtectedRoute>
                } 
              />
              {/* Volunteer routes */}
              <Route 
                path="volunteers" 
                element={
                  <ProtectedRoute requiredRoles={['team_lead', 'campaign_manager', 'admin']}>
                    <VolunteersPoolPage />
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="volunteers/:id/scorecard" 
                element={
                  <ProtectedRoute requiredRoles={['team_lead', 'admin']}>
                    <VolunteerScorecardPage />
                  </ProtectedRoute>
                } 
              />
              {/* Campaign routes */}
              <Route 
                path="campaigns" 
                element={
                  <ProtectedRoute requiredRoles={['campaign_manager', 'admin']}>
                    <CampaignsListPage />
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="campaigns/create" 
                element={
                  <ProtectedRoute requiredRoles={['campaign_manager', 'admin']}>
                    <CreateCampaignPage />
                  </ProtectedRoute>
                } 
              />
              {/* Admin routes */}
              <Route 
                path="admin/users" 
                element={
                  <ProtectedRoute requiredRole="admin">
                    <UserManagementPage />
                  </ProtectedRoute>
                } 
              />
              <Route 
                path="admin/roles" 
                element={
                  <ProtectedRoute requiredRole="admin">
                    <RoleManagementPage />
                  </ProtectedRoute>
                } 
              />
            </Route>
          </Routes>
          <Toast />
        </div>
      </AuthProvider>
    </ToastProvider>
  )
}

export default App
