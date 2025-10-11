import { createContext, useContext, useState, useEffect } from 'react'
import api from '../services/api'

const AuthContext = createContext()

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)
  const [token, setToken] = useState(localStorage.getItem('token'))

  useEffect(() => {
    if (token) {
      api.defaults.headers.common['Authorization'] = `Bearer ${token}`
      // Verify token and get user profile
      fetchUserProfile()
    } else {
      setLoading(false)
    }
  }, [token])

  const fetchUserProfile = async () => {
    try {
      const response = await api.get('/me')
      setUser(response.data)
    } catch (error) {
      // Token is invalid, clear it
      localStorage.removeItem('token')
      setToken(null)
      delete api.defaults.headers.common['Authorization']
    } finally {
      setLoading(false)
    }
  }

  const register = async (userData) => {
    try {
      const response = await api.post('/auth/register', userData)
      return response.data
    } catch (error) {
      throw new Error(error.response?.data?.error || 'Registration failed')
    }
  }

  const login = async (email, password) => {
    console.log('🔐 AUTH_CONTEXT: Starting login process')
    console.log('📧 AUTH_CONTEXT: Email:', email)
    console.log('🔑 AUTH_CONTEXT: Password length:', password?.length || 0)
    
    try {
      console.log('🚀 AUTH_CONTEXT: Making API request to /auth/login')
      const response = await api.post('/auth/login', { email, password })
      console.log('✅ AUTH_CONTEXT: API response received:', response.status, response.statusText)
      console.log('📦 AUTH_CONTEXT: Response data:', response.data)
      
      const { token: newToken, user: userData } = response.data
      console.log('🎫 AUTH_CONTEXT: Token received, length:', newToken?.length || 0)
      console.log('👤 AUTH_CONTEXT: User data received:', userData)
      
      console.log('💾 AUTH_CONTEXT: Storing token in localStorage')
      localStorage.setItem('token', newToken)
      setToken(newToken)
      setUser(userData)
      
      console.log('🔧 AUTH_CONTEXT: Setting Authorization header')
      api.defaults.headers.common['Authorization'] = `Bearer ${newToken}`
      
      console.log('🎉 AUTH_CONTEXT: Login process completed successfully')
      return response.data
    } catch (error) {
      console.error('❌ AUTH_CONTEXT: Login failed:', error)
      console.error('❌ AUTH_CONTEXT: Error response:', error.response)
      console.error('❌ AUTH_CONTEXT: Error message:', error.message)
      console.error('❌ AUTH_CONTEXT: Error data:', error.response?.data)
      
      const errorMessage = error.response?.data?.error || 'Login failed'
      console.error('❌ AUTH_CONTEXT: Final error message:', errorMessage)
      throw new Error(errorMessage)
    }
  }

  const logout = () => {
    localStorage.removeItem('token')
    setToken(null)
    setUser(null)
    delete api.defaults.headers.common['Authorization']
  }

  const verifyEmail = async (token) => {
    try {
      const response = await api.post('/auth/verify-email', { token })
      return response.data
    } catch (error) {
      throw new Error(error.response?.data?.error || 'Email verification failed')
    }
  }

  // Helper functions for role checking
  const hasRole = (roleName) => {
    if (!user?.roles) return false
    return user.roles.includes(roleName)
  }

  const hasAnyRole = (...roleNames) => {
    if (!user?.roles) return false
    return roleNames.some(role => user.roles.includes(role))
  }

  const hasAllRoles = (...roleNames) => {
    if (!user?.roles) return false
    return roleNames.every(role => user.roles.includes(role))
  }

  const value = {
    user,
    token,
    loading,
    register,
    login,
    logout,
    verifyEmail,
    isAuthenticated: !!token && !!user,
    // Legacy role checks (for backward compatibility)
    isAdmin: user?.role === 'admin' || hasRole('admin'),
    isVolunteer: user?.role === 'volunteer' || hasRole('volunteer'),
    // New multi-role support
    hasRole,
    hasAnyRole,
    hasAllRoles,
    // Specific role checks
    isTeamLead: hasRole('team_lead'),
    isCampaignManager: hasRole('campaign_manager')
  }

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  )
}
