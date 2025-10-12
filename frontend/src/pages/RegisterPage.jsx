import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { useToast } from '../contexts/ToastContext'
import { ChevronLeftIcon } from '@heroicons/react/24/outline'
import SkillChipInput from '../components/SkillChipInput'

export default function RegisterPage() {
  const [formData, setFormData] = useState({
    email: '',
    password: '',
    confirmPassword: '',
    name: '',
    phone: '',
    locationAddress: '',
    selectedSkills: [], // New chip-based skills
    availability: {
      weekdays: false,
      weekends: false,
      mornings: false,
      afternoons: false,
      evenings: false,
      flexible: false
    },
    skillsVisible: true,
    consentGiven: false
  })
  const [loading, setLoading] = useState(false)
  const [errors, setErrors] = useState({})

  const { register } = useAuth()
  const { showToast } = useToast()
  const navigate = useNavigate()

  const handleInputChange = (e) => {
    const { name, value, type, checked } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }))
    
    // Clear error when user starts typing
    if (errors[name]) {
      setErrors(prev => ({ ...prev, [name]: '' }))
    }
  }

  // Removed handleSkillToggle - now using textarea for skills

  const handleAvailabilityChange = (field) => {
    setFormData(prev => ({
      ...prev,
      availability: {
        ...prev.availability,
        [field]: !prev.availability[field]
      }
    }))
  }

  const validateForm = () => {
    const newErrors = {}

    if (!formData.email) {
      newErrors.email = 'Email is required'
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = 'Email is invalid'
    }

    if (!formData.password) {
      newErrors.password = 'Password is required'
    } else if (formData.password.length < 8) {
      newErrors.password = 'Password must be at least 8 characters'
    }

    if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match'
    }

    if (!formData.name.trim()) {
      newErrors.name = 'Name is required'
    }

    // Skills are now optional - no validation needed

    if (!formData.consentGiven) {
      newErrors.consentGiven = 'You must agree to the terms and conditions'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    
    if (!validateForm()) {
      return
    }

    setLoading(true)
    try {
      const response = await register({
        email: formData.email,
        password: formData.password,
        name: formData.name,
        phone: formData.phone,
        location_address: formData.locationAddress,
        selected_skills: formData.selectedSkills, // New chip-based skills
        availability: formData.availability,
        skills_visible: formData.skillsVisible,
        consent_given: formData.consentGiven
      })
      
      // Check if email verification is required (based on backend message)
      const requiresVerification = response.message?.includes('email for verification')
      
      if (requiresVerification) {
        showToast('Registration successful! Please check your email for verification.', 'success')
        navigate('/verify-email', { state: { email: formData.email } })
      } else {
        showToast('Registration successful! You can now login.', 'success')
        navigate('/login', { state: { email: formData.email } })
      }
    } catch (error) {
      showToast(error.message || 'Registration failed. Please try again.', 'error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <Link 
            to="/" 
            className="inline-flex items-center text-primary-600 hover:text-primary-500 mb-4"
          >
            <ChevronLeftIcon className="h-4 w-4 mr-1" />
            Back to Home
          </Link>
          
          <h2 className="mt-6 text-center text-3xl font-extrabold text-secondary-900">
            Join CivicWeave
          </h2>
          <p className="mt-2 text-center text-sm text-secondary-600">
            Start making a difference in your community
          </p>
        </div>

        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          <div className="space-y-4">
            {/* Email */}
            <div>
              <label htmlFor="email" className="form-label">
                Email Address *
              </label>
              <input
                id="email"
                name="email"
                type="email"
                autoComplete="email"
                required
                className={`input-field ${errors.email ? 'border-red-500' : ''}`}
                placeholder="Enter your email"
                value={formData.email}
                onChange={handleInputChange}
              />
              {errors.email && <p className="mt-1 text-sm text-red-600">{errors.email}</p>}
            </div>

            {/* Password */}
            <div>
              <label htmlFor="password" className="form-label">
                Password *
              </label>
              <input
                id="password"
                name="password"
                type="password"
                autoComplete="new-password"
                required
                className={`input-field ${errors.password ? 'border-red-500' : ''}`}
                placeholder="Create a password (min 8 characters)"
                value={formData.password}
                onChange={handleInputChange}
              />
              {errors.password && <p className="mt-1 text-sm text-red-600">{errors.password}</p>}
            </div>

            {/* Confirm Password */}
            <div>
              <label htmlFor="confirmPassword" className="form-label">
                Confirm Password *
              </label>
              <input
                id="confirmPassword"
                name="confirmPassword"
                type="password"
                autoComplete="new-password"
                required
                className={`input-field ${errors.confirmPassword ? 'border-red-500' : ''}`}
                placeholder="Confirm your password"
                value={formData.confirmPassword}
                onChange={handleInputChange}
              />
              {errors.confirmPassword && <p className="mt-1 text-sm text-red-600">{errors.confirmPassword}</p>}
            </div>

            {/* Name */}
            <div>
              <label htmlFor="name" className="form-label">
                Full Name *
              </label>
              <input
                id="name"
                name="name"
                type="text"
                autoComplete="name"
                required
                className={`input-field ${errors.name ? 'border-red-500' : ''}`}
                placeholder="Enter your full name"
                value={formData.name}
                onChange={handleInputChange}
              />
              {errors.name && <p className="mt-1 text-sm text-red-600">{errors.name}</p>}
            </div>

            {/* Phone */}
            <div>
              <label htmlFor="phone" className="form-label">
                Phone Number
              </label>
              <input
                id="phone"
                name="phone"
                type="tel"
                autoComplete="tel"
                className="input-field"
                placeholder="Enter your phone number"
                value={formData.phone}
                onChange={handleInputChange}
              />
            </div>

            {/* Location */}
            <div>
              <label htmlFor="locationAddress" className="form-label">
                Location
              </label>
              <input
                id="locationAddress"
                name="locationAddress"
                type="text"
                className="input-field"
                placeholder="City, State/Province (e.g., Toronto, ON)"
                value={formData.locationAddress}
                onChange={handleInputChange}
              />
              <p className="mt-1 text-xs text-secondary-500">
                This helps us match you with nearby opportunities
              </p>
            </div>

            {/* Skills Selection */}
            <div>
              <label className="form-label">
                Skills & Experience (Optional)
              </label>
              <div className="text-sm text-secondary-600 mb-3">
                Select your skills to help us match you with relevant opportunities. You can add more later in your profile.
              </div>
              <SkillChipInput
                selectedSkills={formData.selectedSkills}
                onChange={(skills) => setFormData(prev => ({ ...prev, selectedSkills: skills }))}
                placeholder="Type skills or select from suggestions"
                maxSkills={15}
              />
            </div>

            {/* Skills Visibility */}
            <div>
              <label className="flex items-start space-x-2">
                <input
                  type="checkbox"
                  checked={formData.skillsVisible}
                  onChange={handleInputChange}
                  name="skillsVisible"
                  className="mt-1 rounded border-secondary-300 text-primary-600 focus:ring-primary-500"
                />
                <span className="text-sm text-secondary-700">
                  Make my skills visible to initiative organizers for better matching
                  <span className="block text-xs text-secondary-500 mt-1">
                    You can change this later in your profile
                  </span>
                </span>
              </label>
            </div>

            {/* Availability */}
            <div>
              <label className="form-label">
                Availability
              </label>
              <div className="grid grid-cols-2 gap-2">
                {Object.entries(formData.availability).map(([key, value]) => (
                  <label key={key} className="flex items-center space-x-2 text-sm">
                    <input
                      type="checkbox"
                      checked={value}
                      onChange={() => handleAvailabilityChange(key)}
                      className="rounded border-secondary-300 text-primary-600 focus:ring-primary-500"
                    />
                    <span className="capitalize">{key}</span>
                  </label>
                ))}
              </div>
            </div>

            {/* Consent */}
            <div>
              <label className="flex items-start space-x-2">
                <input
                  type="checkbox"
                  checked={formData.consentGiven}
                  onChange={handleInputChange}
                  name="consentGiven"
                  className="mt-1 rounded border-secondary-300 text-primary-600 focus:ring-primary-500"
                />
                <span className="text-sm text-secondary-700">
                  I agree to the{' '}
                  <a href="#" className="text-primary-600 hover:text-primary-500">
                    Terms of Service
                  </a>{' '}
                  and{' '}
                  <a href="#" className="text-primary-600 hover:text-primary-500">
                    Privacy Policy
                  </a>
                  . I consent to receiving emails about volunteer opportunities and platform updates.
                </span>
              </label>
              {errors.consentGiven && <p className="mt-1 text-sm text-red-600">{errors.consentGiven}</p>}
            </div>
          </div>

          <div>
            <button
              type="submit"
              disabled={loading}
              className="group relative w-full flex justify-center py-3 px-4 border border-transparent text-sm font-medium rounded-lg text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? 'Creating Account...' : 'Create Account'}
            </button>
          </div>

          <div className="text-center">
            <p className="text-sm text-secondary-600">
              Already have an account?{' '}
              <Link to="/login" className="font-medium text-primary-600 hover:text-primary-500">
                Sign in
              </Link>
            </p>
          </div>
        </form>
      </div>
    </div>
  )
}
