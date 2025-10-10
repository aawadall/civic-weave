import { Link } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'

export default function Header() {
  const { user, logout, isAuthenticated } = useAuth()

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
            {user?.role === 'admin' && (
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
                <span className="text-sm text-secondary-700">
                  Welcome, {user?.name || user?.email}
                </span>
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
