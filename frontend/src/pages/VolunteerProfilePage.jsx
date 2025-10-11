import { useState, useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { useToast } from '../contexts/ToastContext'
import SkillChipInput from '../components/SkillChipInput'
import ProfileCompletionModal, { useProfileCompletionModal } from '../components/ProfileCompletionModal'
import api from '../services/api'
import { 
  UserIcon, 
  EyeIcon, 
  EyeSlashIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon,
  PencilIcon
} from '@heroicons/react/24/outline'

export default function VolunteerProfilePage() {
  const [selectedSkills, setSelectedSkills] = useState([])
  const [skillsVisible, setSkillsVisible] = useState(true)
  const [loading, setLoading] = useState(true)
  const [updatingVisibility, setUpdatingVisibility] = useState(false)
  const [updatingSkills, setUpdatingSkills] = useState(false)
  const [profileCompletion, setProfileCompletion] = useState(0)
  const [editingSkills, setEditingSkills] = useState(false)
  
  const { user } = useAuth()
  const { showToast } = useToast()
  const { showModal, completionPercentage, onClose } = useProfileCompletionModal()

  useEffect(() => {
    fetchSkills()
    fetchSkillsVisibility()
    fetchProfileCompletion()
  }, [])

  const fetchSkills = async () => {
    try {
      const response = await api.get('/volunteers/me/skills')
      const skills = response.data.skills || []
      setSelectedSkills(skills.map(skill => skill.skill_name))
    } catch (error) {
      showToast('Failed to load skills', 'error')
    }
  }

  const fetchProfileCompletion = async () => {
    try {
      const response = await api.get('/volunteers/me/profile-completion')
      setProfileCompletion(response.data.completion_percentage || 0)
    } catch (error) {
      console.error('Failed to load profile completion:', error)
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

  const handleSkillsUpdate = async (newSkills) => {
    setUpdatingSkills(true)
    try {
      await api.put('/volunteers/me/skills', {
        skill_names: newSkills
      })
      setSelectedSkills(newSkills)
      await fetchProfileCompletion() // Refresh completion percentage
      showToast('Skills updated successfully!', 'success')
      setEditingSkills(false)
    } catch (error) {
      showToast('Failed to update skills', 'error')
    } finally {
      setUpdatingSkills(false)
    }
  }

  const handleSkillClaimAdded = (newClaim) => {
    // Legacy support - not used in new system
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
    <>
      <ProfileCompletionModal 
        isOpen={showModal}
        onClose={onClose}
        completionPercentage={completionPercentage}
      />
      <div className="min-h-screen bg-secondary-50 py-8">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center space-x-3">
              <UserIcon className="h-8 w-8 text-primary-600" />
              <div>
                <h1 className="text-3xl font-bold text-secondary-900">My Profile</h1>
                <p className="text-secondary-600">
                  Manage your skills and preferences to get better volunteer opportunities
                </p>
              </div>
            </div>
            <div className="text-right">
              <div className={`text-2xl font-bold ${
                profileCompletion >= 100 ? 'text-green-600' :
                profileCompletion >= 70 ? 'text-blue-600' :
                profileCompletion >= 40 ? 'text-yellow-600' : 'text-red-600'
              }`}>
                {profileCompletion}%
              </div>
              <div className="text-sm text-secondary-500">Profile Complete</div>
            </div>
          </div>
          
          {/* Progress Bar */}
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div 
              className={`h-2 rounded-full transition-all duration-500 ${
                profileCompletion >= 100 ? 'bg-green-500' :
                profileCompletion >= 70 ? 'bg-blue-500' :
                profileCompletion >= 40 ? 'bg-yellow-500' : 'bg-red-500'
              }`}
              style={{ width: `${Math.max(profileCompletion, 5)}%` }}
            />
          </div>
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
                <div className="flex items-center justify-between">
                  <div>
                    <h2 className="text-lg font-semibold text-secondary-900">
                      Skills & Experience
                    </h2>
                    <p className="text-sm text-secondary-600 mt-1">
                      Select your skills to get matched with relevant opportunities
                    </p>
                  </div>
                  {!editingSkills && (
                    <button
                      onClick={() => setEditingSkills(true)}
                      className="flex items-center space-x-2 px-3 py-2 text-sm bg-primary-600 text-white rounded-lg hover:bg-primary-700"
                    >
                      <PencilIcon className="h-4 w-4" />
                      <span>Edit Skills</span>
                    </button>
                  )}
                </div>
              </div>
              <div className="px-6 py-6">
                {editingSkills ? (
                  <div className="space-y-4">
                    <SkillChipInput
                      selectedSkills={selectedSkills}
                      onChange={setSelectedSkills}
                      placeholder="Type skills or select from suggestions"
                      maxSkills={20}
                    />
                    <div className="flex justify-end space-x-3">
                      <button
                        onClick={() => setEditingSkills(false)}
                        className="px-4 py-2 text-sm text-secondary-700 bg-secondary-100 rounded-lg hover:bg-secondary-200"
                      >
                        Cancel
                      </button>
                      <button
                        onClick={() => handleSkillsUpdate(selectedSkills)}
                        disabled={updatingSkills}
                        className="px-4 py-2 text-sm bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50"
                      >
                        {updatingSkills ? 'Saving...' : 'Save Changes'}
                      </button>
                    </div>
                  </div>
                ) : (
                  <div>
                    {selectedSkills.length > 0 ? (
                      <div className="flex flex-wrap gap-2">
                        {selectedSkills.map((skill, index) => (
                          <span
                            key={`${skill}-${index}`}
                            className="inline-flex items-center px-3 py-1 bg-primary-100 text-primary-800 rounded-full text-sm font-medium"
                          >
                            {skill}
                          </span>
                        ))}
                      </div>
                    ) : (
                      <div className="text-center py-8 text-secondary-500">
                        <p>No skills added yet.</p>
                        <p className="text-sm">Click "Edit Skills" to add your skills.</p>
                      </div>
                    )}
                  </div>
                )}
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
    </>
  )
}
