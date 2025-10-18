import { useState } from 'react'
import { 
  getAvailableTransitions, 
  getStatusLabel, 
  getStatusColor, 
  getTransitionRequirements,
  getValidationChecklist,
  getWorkflowStages,
  PROJECT_STATUSES 
} from '../utils/projectLifecycle'
import api from '../services/api'

export default function ProjectStatusTransition({ 
  project, 
  onStatusChange, 
  showWorkflow = false,
  compact = false 
}) {
  const [isTransitioning, setIsTransitioning] = useState(false)
  const [showConfirmModal, setShowConfirmModal] = useState(false)
  const [selectedStatus, setSelectedStatus] = useState(null)
  const [error, setError] = useState(null)

  // Don't render if no project data
  if (!project) {
    console.log('ProjectStatusTransition: No project data')
    return null
  }

  const currentStatus = project?.project_status || project?.status
  console.log('ProjectStatusTransition: Current status:', currentStatus, 'Project:', project)
  const availableTransitions = getAvailableTransitions(currentStatus, project || {})
  const workflowStages = getWorkflowStages()

  const handleStatusSelect = (newStatus) => {
    setSelectedStatus(newStatus)
    setError(null)
    
    // Check if transition requires confirmation
    const requirements = getTransitionRequirements(currentStatus, newStatus)
    if (requirements.message) {
      setShowConfirmModal(true)
    } else {
      handleTransition(newStatus)
    }
  }

  const handleTransition = async (newStatus) => {
    if (!newStatus || !project?.id) return

    setIsTransitioning(true)
    setError(null)

    try {
      const response = await api.put(`/projects/${project.id}/status`, {
        status: newStatus
      })
      
      console.log('✅ ProjectStatusTransition: Status updated successfully:', response.data)
      setShowConfirmModal(false)
      setSelectedStatus(null)
      
      if (onStatusChange) {
        onStatusChange(response.data.project || { ...project, project_status: newStatus })
      }
    } catch (err) {
      console.error('Error transitioning project status:', err)
      const errorData = err.response?.data
      
      // Handle structured error responses
      if (errorData?.code === 'INVALID_TRANSITION') {
        setError({
          title: 'Cannot Transition Project Status',
          message: errorData.error,
          requirements: errorData.details?.requirements || [],
          type: 'validation'
        })
      } else if (errorData?.code === 'INSUFFICIENT_PERMISSIONS') {
        setError({
          title: 'Permission Denied',
          message: 'You do not have permission to change this project\'s status',
          type: 'permission'
        })
      } else {
        setError({
          title: 'Transition Failed',
          message: errorData?.error || 'Failed to update project status',
          type: 'error'
        })
      }
    } finally {
      setIsTransitioning(false)
    }
  }

  const handleConfirmTransition = () => {
    handleTransition(selectedStatus)
  }

  const handleCancelTransition = () => {
    setShowConfirmModal(false)
    setSelectedStatus(null)
    setError(null)
  }

  if (compact) {
    return (
      <div className="flex items-center gap-2">
        <span className={`px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(currentStatus)}`}>
          {getStatusLabel(currentStatus)}
        </span>
        {availableTransitions.length > 0 && (
          <select
            value=""
            onChange={(e) => handleStatusSelect(e.target.value)}
            disabled={isTransitioning}
            className="text-xs border border-gray-300 rounded px-2 py-1 bg-white"
          >
            <option value="">Change Status</option>
            {availableTransitions.map(status => (
              <option key={status} value={status}>
                {getStatusLabel(status)}
              </option>
            ))}
          </select>
        )}
        {error && (
          <span className="text-xs text-red-600">{error}</span>
        )}
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Current Status */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-sm font-medium text-gray-900">Project Status</h3>
          <p className="text-sm text-gray-500">Current stage in the project lifecycle</p>
        </div>
        <span className={`px-3 py-1 text-sm font-medium rounded-full ${getStatusColor(currentStatus)}`}>
          {getStatusLabel(currentStatus)}
        </span>
      </div>

      {/* Workflow Visualization */}
      {showWorkflow && (
        <div className="space-y-2">
          <h4 className="text-sm font-medium text-gray-900">Project Workflow</h4>
          <div className="flex items-center space-x-2">
            {workflowStages.map((stage, index) => (
              <div key={stage.status} className="flex items-center">
                <div className={`px-3 py-1 text-xs font-medium rounded-full ${
                  stage.status === currentStatus 
                    ? stage.color 
                    : 'bg-gray-100 text-gray-500'
                }`}>
                  {stage.label}
                </div>
                {index < workflowStages.length - 1 && (
                  <div className="w-4 h-0.5 bg-gray-300 mx-2" />
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Status Transition */}
      {availableTransitions.length > 0 && (
        <div className="space-y-2">
          <h4 className="text-sm font-medium text-gray-900">Change Status</h4>
          <div className="flex flex-wrap gap-2">
            {availableTransitions.map(status => (
              <button
                key={status}
                onClick={() => handleStatusSelect(status)}
                disabled={isTransitioning}
                className={`px-3 py-2 text-sm font-medium rounded-md border transition-colors ${
                  isTransitioning
                    ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                    : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
                }`}
              >
                {isTransitioning ? 'Updating...' : `Move to ${getStatusLabel(status)}`}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Validation Checklist */}
      {currentStatus === PROJECT_STATUSES.DRAFT && (
        <div className="space-y-2">
          <h4 className="text-sm font-medium text-gray-900">Requirements to Move Forward</h4>
          <div className="space-y-1">
            {getValidationChecklist(project).map(item => (
              <div key={item.key} className="flex items-center text-sm">
                <span className={`w-4 h-4 rounded-full mr-2 flex items-center justify-center ${
                  item.valid ? 'bg-green-100 text-green-600' : 'bg-red-100 text-red-600'
                }`}>
                  {item.valid ? '✓' : '✗'}
                </span>
                <span className={item.valid ? 'text-green-700' : 'text-red-700'}>
                  {item.label}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Error Display */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">
                {typeof error === 'string' ? 'Error' : error.title}
              </h3>
              <div className="mt-2 text-sm text-red-700">
                <p>{typeof error === 'string' ? error : error.message}</p>
                {error.requirements && error.requirements.length > 0 && (
                  <ul className="mt-2 list-disc list-inside">
                    {error.requirements.map((req, index) => (
                      <li key={index}>{req}</li>
                    ))}
                  </ul>
                )}
              </div>
              <div className="mt-3">
                <button
                  onClick={() => setError(null)}
                  className="text-sm font-medium text-red-800 hover:text-red-600"
                >
                  Dismiss
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Confirmation Modal */}
      {showConfirmModal && selectedStatus && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-medium text-gray-900 mb-4">
              Confirm Status Change
            </h3>
            <p className="text-sm text-gray-600 mb-4">
              Are you sure you want to change this project from{' '}
              <span className="font-medium">{getStatusLabel(currentStatus)}</span> to{' '}
              <span className="font-medium">{getStatusLabel(selectedStatus)}</span>?
            </p>
            <div className="flex justify-end space-x-3">
              <button
                onClick={handleCancelTransition}
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200"
              >
                Cancel
              </button>
              <button
                onClick={handleConfirmTransition}
                disabled={isTransitioning}
                className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
              >
                {isTransitioning ? 'Updating...' : 'Confirm'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
