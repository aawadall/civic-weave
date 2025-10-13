import { Link } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import MessagesIcon from './MessagesIcon'

export default function Header() {
  const { user, logout, isAuthenticated, hasAnyRole, hasRole } = useAuth()

  return (
    <header className="bg-white shadow-sm border-b border-secondary-200">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          <div className="flex items-center">
            <Link to="/" className="flex items-center space-x-2">
              <div className="w-8 h-8 bg-primary-600 rounded-lg flex items-center justify-center">
                <span className="text-white font-bold text-sm">CW</span>
              </div>
              <span className="text-xl font-bold text-secondary-900">CivicWeave</span>
            </Link>
          </div>

          <nav className="hidden md:flex space-x-8">
            <Link 
              to="/" 
              className="text-secondary-600 hover:text-secondary-900 px-3 py-2 text-sm font-medium"
            >
              Home
            </Link>
            {isAuthenticated && (
              <Link 
                to="/dashboard" 
                className="text-secondary-600 hover:text-secondary-900 px-3 py-2 text-sm font-medium"
              >
                Dashboard
              </Link>
            )}
            {/* Projects - visible to all authenticated users */}
            {isAuthenticated && (
              <Link 
                to="/projects" 
                className="text-secondary-600 hover:text-secondary-900 px-3 py-2 text-sm font-medium"
              >
                Projects
              </Link>
            )}
            {/* Volunteers - visible to team leads, campaign managers, and admins */}
            {hasAnyRole('team_lead', 'campaign_manager', 'admin') && (
              <Link 
                to="/volunteers" 
                className="text-secondary-600 hover:text-secondary-900 px-3 py-2 text-sm font-medium"
              >
                Volunteers
              </Link>
            )}
            {/* Campaigns - visible to campaign managers and admins */}
            {hasAnyRole('campaign_manager', 'admin') && (
              <Link 
                to="/campaigns" 
                className="text-secondary-600 hover:text-secondary-900 px-3 py-2 text-sm font-medium"
              >
                Campaigns
              </Link>
            )}
            {/* Admin - visible to admins only */}
            {hasRole('admin') && (
              <Link 
                to="/admin" 
                className="text-secondary-600 hover:text-secondary-900 px-3 py-2 text-sm font-medium"
              >
                Admin
              </Link>
            )}
          </nav>

          <div className="flex items-center space-x-4">
            {isAuthenticated ? (
              <div className="flex items-center space-x-4">
                <MessagesIcon />
                <div className="flex flex-col">
                  <span className="text-sm text-secondary-700">
                    Welcome, {user?.name || user?.email}
                  </span>
                  {user?.roles && user.roles.length > 0 && (
                    <span className="text-xs text-secondary-500">
                      {user.roles.join(', ')}
                    </span>
                  )}
                </div>
                <button
                  onClick={logout}
                  className="text-secondary-600 hover:text-secondary-900 text-sm font-medium"
                >
                  Logout
                </button>
              </div>
            ) : (
              <div className="flex items-center space-x-4">
                <Link
                  to="/login"
                  className="text-secondary-600 hover:text-secondary-900 text-sm font-medium"
                >
                  Sign In
                </Link>
                <Link
                  to="/register"
                  className="btn-primary text-sm"
                >
                  Join Now
                </Link>
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  )
}
