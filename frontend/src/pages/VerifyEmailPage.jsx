import { useState, useEffect } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { useToast } from '../contexts/ToastContext'

export default function VerifyEmailPage() {
  const [token, setToken] = useState('')
  const [loading, setLoading] = useState(false)
  const [email, setEmail] = useState('')

  const { verifyEmail } = useAuth()
  const { showToast } = useToast()
  const location = useLocation()
  const navigate = useNavigate()

  useEffect(() => {
    // Get email from location state or URL params
    const emailFromState = location.state?.email
    const urlParams = new URLSearchParams(location.search)
    const emailFromParams = urlParams.get('email')
    const tokenFromParams = urlParams.get('token')

    if (emailFromState) setEmail(emailFromState)
    if (emailFromParams) setEmail(emailFromParams)
    if (tokenFromParams) setToken(tokenFromParams)
  }, [location])

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (!token.trim()) {
      showToast('Please enter a verification token', 'error')
      return
    }

    setLoading(true)
    try {
      await verifyEmail(token)
      showToast('Email verified successfully!', 'success')
      navigate('/login')
    } catch (error) {
      showToast(error.message, 'error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div className="text-center">
          <div className="mx-auto h-12 w-12 bg-primary-100 rounded-full flex items-center justify-center">
            <span className="text-2xl">📧</span>
          </div>
          <h2 className="mt-6 text-3xl font-extrabold text-secondary-900">
            Verify Your Email
          </h2>
          <p className="mt-2 text-sm text-secondary-600">
            {email ? `We've sent a verification link to ${email}` : 'Check your email for a verification link'}
          </p>
        </div>

        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          <div>
            <label htmlFor="token" className="form-label">
              Verification Token
            </label>
            <input
              id="token"
              name="token"
              type="text"
              required
              className="input-field"
              placeholder="Enter verification token from email"
              value={token}
              onChange={(e) => setToken(e.target.value)}
            />
            <p className="mt-1 text-xs text-secondary-500">
              Paste the verification token from your email, or click the link in your email.
            </p>
          </div>

          <div>
            <button
              type="submit"
              disabled={loading}
              className="group relative w-full flex justify-center py-3 px-4 border border-transparent text-sm font-medium rounded-lg text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? 'Verifying...' : 'Verify Email'}
            </button>
          </div>

          <div className="text-center">
            <p className="text-sm text-secondary-600">
              Didn't receive an email?{' '}
              <button
                type="button"
                className="font-medium text-primary-600 hover:text-primary-500"
                onClick={() => showToast('Resend functionality coming soon!', 'info')}
              >
                Resend verification email
              </button>
            </p>
          </div>
        </form>
      </div>
    </div>
  )
}
