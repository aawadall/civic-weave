import { useState, useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { useToast } from '../contexts/ToastContext'
import SkillClaimInput from '../components/SkillClaimInput'
import api from '../services/api'
import { 
  UserIcon, 
  EyeIcon, 
  EyeSlashIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon
} from '@heroicons/react/24/outline'

export default function VolunteerProfilePage() {
  const [skillClaims, setSkillClaims] = useState([])
  const [skillsVisible, setSkillsVisible] = useState(true)
  const [loading, setLoading] = useState(true)
  const [updatingVisibility, setUpdatingVisibility] = useState(false)
  const [showSkillsPrompt, setShowSkillsPrompt] = useState(false)
  
  const { user } = useAuth()
  const { showToast } = useToast()

  useEffect(() => {
    fetchSkillClaims()
    fetchSkillsVisibility()
    
    // Show prompt if user has no skills and this is their first visit
    const lastPromptDate = localStorage.getItem('lastSkillsPrompt')
    const today = new Date().toDateString()
    if (!lastPromptDate || lastPromptDate !== today) {
      setShowSkillsPrompt(true)
      localStorage.setItem('lastSkillsPrompt', today)
    }
  }, [])

  const fetchSkillClaims = async () => {
    try {
      const response = await api.get('/volunteers/me/skills')
      setSkillClaims(response.data.claims || [])
    } catch (error) {
      showToast('Failed to load skill claims', 'error')
    }
  }

  const fetchSkillsVisibility = async () => {
    try {
      const response = await api.get('/volunteers/me/skills-visibility')
      setSkillsVisible(response.data.visible)
    } catch (error) {
      showToast('Failed to load skills visibility setting', 'error')
    } finally {
      setLoading(false)
    }
  }

  const handleSkillClaimAdded = (newClaim) => {
    setSkillClaims(prev => [newClaim, ...prev])
    setShowSkillsPrompt(false) // Hide prompt when user adds a skill
  }

  const handleSkillClaimDeleted = (claimId) => {
    setSkillClaims(prev => prev.filter(claim => claim.id !== claimId))
  }

  const handleVisibilityToggle = async () => {
    setUpdatingVisibility(true)
    try {
      const newVisibility = !skillsVisible
      await api.put('/volunteers/me/skills-visibility', {
        visible: newVisibility
      })
      setSkillsVisible(newVisibility)
      showToast(
        newVisibility 
          ? 'Your skills are now visible to organizers' 
          : 'Your skills are now hidden from organizers',
        'success'
      )
    } catch (error) {
      const message = error.response?.data?.error || 'Failed to update visibility setting'
      showToast(message, 'error')
    } finally {
      setUpdatingVisibility(false)
    }
  }

  const dismissSkillsPrompt = () => {
    setShowSkillsPrompt(false)
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-secondary-50 py-8">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center space-x-3 mb-2">
            <UserIcon className="h-8 w-8 text-primary-600" />
            <h1 className="text-3xl font-bold text-secondary-900">My Profile</h1>
          </div>
          <p className="text-secondary-600">
            Manage your skills and preferences to get better volunteer opportunities
          </p>
        </div>

        {/* Skills Update Prompt */}
        {showSkillsPrompt && (
          <div className="mb-6 bg-blue-50 border border-blue-200 rounded-lg p-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <ExclamationTriangleIcon className="h-5 w-5 text-blue-400" />
              </div>
              <div className="ml-3 flex-1">
                <h3 className="text-sm font-medium text-blue-800">
                  Are your skills up to date?
                </h3>
                <div className="mt-2 text-sm text-blue-700">
                  <p>
                    Help us match you with the right opportunities by keeping your skills current. 
                    Add or update your skills to improve your matching results.
                  </p>
                </div>
                <div className="mt-3">
                  <button
                    onClick={dismissSkillsPrompt}
                    className="text-sm bg-blue-100 text-blue-800 px-3 py-1 rounded hover:bg-blue-200"
                  >
                    Dismiss
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Content */}
          <div className="lg:col-span-2 space-y-6">
            {/* Skills Management */}
            <div className="bg-white shadow rounded-lg">
              <div className="px-6 py-4 border-b border-secondary-200">
                <h2 className="text-lg font-semibold text-secondary-900">
                  Skills & Experience
                </h2>
                <p className="text-sm text-secondary-600 mt-1">
                  Describe your skills and experience to get matched with relevant opportunities
                </p>
              </div>
              <div className="px-6 py-6">
                <SkillClaimInput
                  existingClaims={skillClaims}
                  onClaimAdded={handleSkillClaimAdded}
                  onClaimDeleted={handleSkillClaimDeleted}
                />
              </div>
            </div>
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Skills Visibility */}
            <div className="bg-white shadow rounded-lg">
              <div className="px-6 py-4 border-b border-secondary-200">
                <h3 className="text-lg font-semibold text-secondary-900">
                  Skills Visibility
                </h3>
              </div>
              <div className="px-6 py-4">
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <p className="text-sm font-medium text-secondary-900">
                      Show my skills to organizers
                    </p>
                    <p className="text-sm text-secondary-500 mt-1">
                      {skillsVisible 
                        ? 'Your skills are visible and will be used for matching'
                        : 'Your skills are hidden from organizers'
                      }
                    </p>
                  </div>
                  <button
                    onClick={handleVisibilityToggle}
                    disabled={updatingVisibility}
                    className={`ml-4 relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 ${
                      skillsVisible ? 'bg-primary-600' : 'bg-secondary-200'
                    } disabled:opacity-50`}
                  >
                    <span
                      className={`pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out ${
                        skillsVisible ? 'translate-x-5' : 'translate-x-0'
                      }`}
                    />
                  </button>
                </div>
                <div className="mt-4 flex items-center space-x-2 text-sm">
                  {skillsVisible ? (
                    <>
                      <EyeIcon className="h-4 w-4 text-green-500" />
                      <span className="text-green-700">Visible</span>
                    </>
                  ) : (
                    <>
                      <EyeSlashIcon className="h-4 w-4 text-gray-500" />
                      <span className="text-gray-700">Hidden</span>
                    </>
                  )}
                </div>
              </div>
            </div>

            {/* Profile Stats */}
            <div className="bg-white shadow rounded-lg">
              <div className="px-6 py-4 border-b border-secondary-200">
                <h3 className="text-lg font-semibold text-secondary-900">
                  Profile Stats
                </h3>
              </div>
              <div className="px-6 py-4 space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-secondary-600">Skill Claims</span>
                  <span className="text-sm font-medium text-secondary-900">
                    {skillClaims.length}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-secondary-600">Average Weight</span>
                  <span className="text-sm font-medium text-secondary-900">
                    {skillClaims.length > 0 
                      ? Math.round(
                          (skillClaims.reduce((sum, claim) => sum + claim.weight.weight, 0) / skillClaims.length) * 100
                        ) + '%'
                      : 'N/A'
                    }
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-secondary-600">Member Since</span>
                  <span className="text-sm font-medium text-secondary-900">
                    {user?.created_at ? new Date(user.created_at).toLocaleDateString() : 'N/A'}
                  </span>
                </div>
              </div>
            </div>

            {/* Quick Actions */}
            <div className="bg-white shadow rounded-lg">
              <div className="px-6 py-4 border-b border-secondary-200">
                <h3 className="text-lg font-semibold text-secondary-900">
                  Quick Actions
                </h3>
              </div>
              <div className="px-6 py-4 space-y-3">
                <button className="w-full text-left px-3 py-2 text-sm text-secondary-700 hover:bg-secondary-50 rounded-lg">
                  View My Matches
                </button>
                <button className="w-full text-left px-3 py-2 text-sm text-secondary-700 hover:bg-secondary-50 rounded-lg">
                  Update Availability
                </button>
                <button className="w-full text-left px-3 py-2 text-sm text-secondary-700 hover:bg-secondary-50 rounded-lg">
                  Account Settings
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
