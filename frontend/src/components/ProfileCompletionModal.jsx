import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { XMarkIcon, CheckCircleIcon, ExclamationTriangleIcon } from '@heroicons/react/24/outline'
import api from '../services/api'

export default function ProfileCompletionModal({ isOpen, onClose, completionPercentage }) {
  const [showModal, setShowModal] = useState(isOpen)

  useEffect(() => {
    setShowModal(isOpen)
  }, [isOpen])

  const handleClose = () => {
    setShowModal(false)
    onClose()
  }

  const handleRemindLater = () => {
    // Store timestamp in localStorage to prevent showing again for 3 days
    const now = new Date()
    const threeDaysLater = new Date(now.getTime() + (3 * 24 * 60 * 60 * 1000))
    localStorage.setItem('profileCompletionReminder', threeDaysLater.toISOString())
    handleClose()
  }

  const handleCompleteNow = () => {
    handleClose()
    // The Link component will handle navigation
  }

  const getCompletionColor = (percentage) => {
    if (percentage >= 100) return 'text-green-600'
    if (percentage >= 70) return 'text-blue-600'
    if (percentage >= 40) return 'text-yellow-600'
    return 'text-red-600'
  }

  const getCompletionMessage = (percentage) => {
    if (percentage >= 100) return 'Your profile is complete!'
    if (percentage >= 70) return 'Almost there! Just a few more details.'
    if (percentage >= 40) return 'Good start! Let\'s complete your profile.'
    return 'Let\'s get your profile set up!'
  }

  const getMissingSections = (percentage) => {
    const missing = []
    if (percentage < 100) {
      if (percentage < 40) missing.push('Skills')
      if (percentage < 70) missing.push('Location')
      if (percentage < 100) missing.push('Availability')
    }
    return missing
  }

  if (!showModal) return null

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex min-h-full items-center justify-center p-4">
        {/* Backdrop */}
        <div 
          className="fixed inset-0 bg-black bg-opacity-50 transition-opacity"
          onClick={handleClose}
        />

        {/* Modal */}
        <div className="relative transform overflow-hidden rounded-lg bg-white shadow-xl transition-all sm:w-full sm:max-w-lg">
          {/* Header */}
          <div className="bg-white px-4 pb-4 pt-5 sm:p-6 sm:pb-4">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center space-x-3">
                <div className={`flex-shrink-0 ${getCompletionColor(completionPercentage)}`}>
                  {completionPercentage >= 100 ? (
                    <CheckCircleIcon className="h-6 w-6" />
                  ) : (
                    <ExclamationTriangleIcon className="h-6 w-6" />
                  )}
                </div>
                <h3 className="text-lg font-medium text-gray-900">
                  Profile Completion
                </h3>
              </div>
              <button
                onClick={handleClose}
                className="rounded-md bg-white text-gray-400 hover:text-gray-500 focus:outline-none"
              >
                <XMarkIcon className="h-6 w-6" />
              </button>
            </div>

            {/* Progress Bar */}
            <div className="mb-6">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-medium text-gray-700">
                  {getCompletionMessage(completionPercentage)}
                </span>
                <span className={`text-sm font-bold ${getCompletionColor(completionPercentage)}`}>
                  {completionPercentage}%
                </span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-3">
                <div 
                  className={`h-3 rounded-full transition-all duration-500 ${
                    completionPercentage >= 100 ? 'bg-green-500' :
                    completionPercentage >= 70 ? 'bg-blue-500' :
                    completionPercentage >= 40 ? 'bg-yellow-500' : 'bg-red-500'
                  }`}
                  style={{ width: `${Math.max(completionPercentage, 5)}%` }}
                />
              </div>
            </div>

            {/* Missing Sections */}
            {completionPercentage < 100 && (
              <div className="mb-6">
                <h4 className="text-sm font-medium text-gray-700 mb-3">
                  Complete these sections to improve your profile:
                </h4>
                <div className="space-y-2">
                  {getMissingSections(completionPercentage).map((section, index) => (
                    <div key={index} className="flex items-center space-x-2 text-sm text-gray-600">
                      <div className="w-2 h-2 bg-gray-300 rounded-full" />
                      <span>{section}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Benefits */}
            <div className="mb-6 bg-blue-50 border border-blue-200 rounded-lg p-4">
              <h4 className="text-sm font-medium text-blue-900 mb-2">
                Why complete your profile?
              </h4>
              <ul className="text-sm text-blue-800 space-y-1">
                <li>• Get matched with relevant volunteer opportunities</li>
                <li>• Receive personalized recommendations</li>
                <li>• Help organizers understand your skills</li>
                <li>• Build your volunteer reputation</li>
              </ul>
            </div>
          </div>

          {/* Footer */}
          <div className="bg-gray-50 px-4 py-3 sm:flex sm:flex-row-reverse sm:px-6">
            {completionPercentage < 100 ? (
              <>
                <Link
                  to="/profile"
                  onClick={handleCompleteNow}
                  className="inline-flex w-full justify-center rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600 sm:ml-3 sm:w-auto"
                >
                  Complete Profile
                </Link>
                <button
                  type="button"
                  onClick={handleRemindLater}
                  className="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                >
                  Remind Me Later
                </button>
              </>
            ) : (
              <button
                type="button"
                onClick={handleClose}
                className="inline-flex w-full justify-center rounded-md bg-primary-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-primary-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600 sm:w-auto"
              >
                Great Job!
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

// Hook to check if profile completion modal should be shown
export function useProfileCompletionModal() {
  const [showModal, setShowModal] = useState(false)
  const [completionPercentage, setCompletionPercentage] = useState(0)

  useEffect(() => {
    checkShouldShowModal()
  }, [])

  const checkShouldShowModal = async () => {
    try {
      // Check if user has dismissed the modal recently
      const lastReminder = localStorage.getItem('profileCompletionReminder')
      if (lastReminder) {
        const reminderDate = new Date(lastReminder)
        const now = new Date()
        if (now < reminderDate) {
          return // Don't show modal yet
        }
      }

      // Get profile completion percentage
      const response = await api.get('/volunteers/me/profile-completion')
      const percentage = response.data.completion_percentage
      setCompletionPercentage(percentage)

      // Show modal if profile is incomplete
      if (percentage < 100) {
        setShowModal(true)
      }
    } catch (error) {
      console.error('Failed to check profile completion:', error)
    }
  }

  return {
    showModal,
    completionPercentage,
    onClose: () => setShowModal(false)
  }
}
