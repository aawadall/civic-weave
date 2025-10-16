import { Navigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'

export default function ProtectedRoute({ 
  children, 
  requiredRole, 
  requiredRoles, 
  requireAllRoles = false 
}) {
  const { isAuthenticated, user, loading, hasRole, hasAnyRole, hasAllRoles } = useAuth()

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  // Check role requirements
  let hasRequiredRole = true

  if (requiredRole) {
    // Single role check
    hasRequiredRole = hasRole(requiredRole)
  } else if (requiredRoles && requiredRoles.length > 0) {
    // Multi-role support
    if (requireAllRoles) {
      hasRequiredRole = hasAllRoles(...requiredRoles)
    } else {
      hasRequiredRole = hasAnyRole(...requiredRoles)
    }
  }

  if (!hasRequiredRole) {
    return <Navigate to="/dashboard" replace />
  }

  return children
}
