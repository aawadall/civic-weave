import { useState } from 'react'
import { useToast } from '../contexts/ToastContext'
import api from '../services/api'
import { TrashIcon, PlusIcon } from '@heroicons/react/24/outline'

export default function SkillClaimInput({ onClaimAdded, onClaimDeleted, existingClaims = [] }) {
  const [newClaim, setNewClaim] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isDeleting, setIsDeleting] = useState({})
  const { showToast } = useToast()

  const handleSubmit = async (e) => {
    e.preventDefault()
    
    if (!newClaim.trim()) {
      showToast('Please enter a skill description', 'error')
      return
    }

    if (newClaim.trim().length < 10) {
      showToast('Skill description must be at least 10 characters', 'error')
      return
    }

    if (newClaim.trim().length > 500) {
      showToast('Skill description must be less than 500 characters', 'error')
      return
    }

    setIsSubmitting(true)
    try {
      const response = await api.post('/volunteers/me/skills', {
        claim_text: newClaim.trim()
      })

      showToast('Skill claim added successfully!', 'success')
      setNewClaim('')
      
      if (onClaimAdded) {
        onClaimAdded(response.data.claim)
      }
    } catch (error) {
      const message = error.response?.data?.error || 'Failed to add skill claim'
      showToast(message, 'error')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleDelete = async (claimId) => {
    if (!window.confirm('Are you sure you want to remove this skill claim?')) {
      return
    }

    setIsDeleting(prev => ({ ...prev, [claimId]: true }))
    try {
      await api.delete(`/volunteers/me/skills/${claimId}`)
      
      showToast('Skill claim removed successfully!', 'success')
      
      if (onClaimDeleted) {
        onClaimDeleted(claimId)
      }
    } catch (error) {
      const message = error.response?.data?.error || 'Failed to remove skill claim'
      showToast(message, 'error')
    } finally {
      setIsDeleting(prev => ({ ...prev, [claimId]: false }))
    }
  }

  const getWeightColor = (weight) => {
    if (weight >= 0.8) return 'text-green-600 bg-green-50'
    if (weight >= 0.6) return 'text-blue-600 bg-blue-50'
    if (weight >= 0.4) return 'text-yellow-600 bg-yellow-50'
    return 'text-red-600 bg-red-50'
  }

  const getWeightLabel = (weight) => {
    if (weight >= 0.8) return 'Excellent'
    if (weight >= 0.6) return 'Good'
    if (weight >= 0.4) return 'Fair'
    return 'Needs Review'
  }

  return (
    <div className="space-y-6">
      {/* Existing Claims */}
      {existingClaims.length > 0 && (
        <div>
          <h3 className="text-lg font-semibold text-secondary-900 mb-4">Your Skill Claims</h3>
          <div className="space-y-3">
            {existingClaims.map((claim) => (
              <div key={claim.id} className="border border-secondary-200 rounded-lg p-4">
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <p className="text-secondary-900 mb-2">{claim.claim_text}</p>
                    <div className="flex items-center space-x-4 text-sm">
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${getWeightColor(claim.weight.weight)}`}>
                        {getWeightLabel(claim.weight.weight)} ({Math.round(claim.weight.weight * 100)}%)
                      </span>
                      <span className="text-secondary-500">
                        Added {new Date(claim.created_at).toLocaleDateString()}
                      </span>
                    </div>
                  </div>
                  <button
                    onClick={() => handleDelete(claim.id)}
                    disabled={isDeleting[claim.id]}
                    className="ml-4 p-1 text-red-500 hover:text-red-700 hover:bg-red-50 rounded disabled:opacity-50"
                    title="Remove this skill claim"
                  >
                    <TrashIcon className="h-4 w-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Add New Claim */}
      <div>
        <h3 className="text-lg font-semibold text-secondary-900 mb-4">
          {existingClaims.length === 0 ? 'Add Your First Skill Claim' : 'Add Another Skill'}
        </h3>
        
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="claim_text" className="block text-sm font-medium text-secondary-700 mb-2">
              Describe your skills and experience
            </label>
            <textarea
              id="claim_text"
              value={newClaim}
              onChange={(e) => setNewClaim(e.target.value)}
              placeholder="e.g., I have 3 years of experience in community event planning, social media marketing, and volunteer coordination. I've organized several successful fundraising events and managed teams of 10+ volunteers."
              className="w-full px-3 py-2 border border-secondary-300 rounded-lg focus:ring-primary-500 focus:border-primary-500 resize-none"
              rows={4}
              maxLength={500}
              disabled={isSubmitting}
            />
            <div className="mt-1 flex justify-between text-sm text-secondary-500">
              <span>Be specific about your experience, skills, and achievements</span>
              <span>{newClaim.length}/500</span>
            </div>
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-blue-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-blue-800">Tips for better skill matching:</h3>
                <div className="mt-2 text-sm text-blue-700">
                  <ul className="list-disc list-inside space-y-1">
                    <li>Include specific technologies, tools, or methodologies you know</li>
                    <li>Mention years of experience or project outcomes</li>
                    <li>Describe leadership roles or team management experience</li>
                    <li>Add any certifications or specialized training</li>
                  </ul>
                </div>
              </div>
            </div>
          </div>

          <button
            type="submit"
            disabled={isSubmitting || newClaim.trim().length < 10}
            className="w-full flex items-center justify-center px-4 py-2 border border-transparent text-sm font-medium rounded-lg text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isSubmitting ? (
              <>
                <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Adding Skill...
              </>
            ) : (
              <>
                <PlusIcon className="h-5 w-5 mr-2" />
                Add Skill Claim
              </>
            )}
          </button>
        </form>
      </div>
    </div>
  )
}
