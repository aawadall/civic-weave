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
            </Route>
          </Routes>
          <Toast />
        </div>
      </AuthProvider>
    </ToastProvider>
  )
}

export default App
