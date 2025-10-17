import { useState } from 'react'
import { markTaskBlocked, requestTaskTakeover, markTaskDone } from '../services/api'

export default function TaskStatusActions({ task, onStatusChanged }) {
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [showBlockedModal, setShowBlockedModal] = useState(false)
  const [showTakeoverModal, setShowTakeoverModal] = useState(false)
  const [showDoneModal, setShowDoneModal] = useState(false)
  const [reason, setReason] = useState('')
  const [completionNote, setCompletionNote] = useState('')

  const handleMarkBlocked = async () => {
    try {
      setIsSubmitting(true)
      await markTaskBlocked(task.id, reason)
      setShowBlockedModal(false)
      setReason('')
      onStatusChanged?.()
    } catch (error) {
      console.error('Error marking task as blocked:', error)
      alert('Failed to mark task as blocked')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleRequestTakeover = async () => {
    try {
      setIsSubmitting(true)
      await requestTaskTakeover(task.id, reason)
      setShowTakeoverModal(false)
      setReason('')
      onStatusChanged?.()
    } catch (error) {
      console.error('Error requesting takeover:', error)
      alert('Failed to request takeover')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleMarkDone = async () => {
    try {
      setIsSubmitting(true)
      await markTaskDone(task.id, completionNote)
      setShowDoneModal(false)
      setCompletionNote('')
      onStatusChanged?.()
    } catch (error) {
      console.error('Error marking task as done:', error)
      alert('Failed to mark task as done')
    } finally {
      setIsSubmitting(false)
    }
  }

  // Don't show actions if task is already done
  if (task.status === 'done') {
    return (
      <div className="text-center py-4">
        <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-green-100 text-green-800">
          ✅ Task Completed
        </span>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      <h3 className="text-lg font-semibold text-secondary-900">Task Actions</h3>
      
      <div className="grid grid-cols-1 gap-3">
        <button
          onClick={() => setShowDoneModal(true)}
          className="w-full px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
        >
          ✅ Mark as Done
        </button>
        
        <button
          onClick={() => setShowBlockedModal(true)}
          className="w-full px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
        >
          🚫 Mark as Blocked
        </button>
        
        <button
          onClick={() => setShowTakeoverModal(true)}
          className="w-full px-4 py-2 bg-orange-600 text-white rounded-lg hover:bg-orange-700 transition-colors"
        >
          🔄 Request Takeover
        </button>
      </div>

      {/* Mark as Done Modal */}
      {showDoneModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg max-w-md w-full p-6">
            <h3 className="text-lg font-bold text-secondary-900 mb-4">Mark Task as Done</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-secondary-900 mb-2">
                  Completion Note (optional)
                </label>
                <textarea
                  value={completionNote}
                  onChange={(e) => setCompletionNote(e.target.value)}
                  placeholder="What was accomplished? Any final notes?"
                  rows={3}
                  className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500"
                />
              </div>
              <div className="flex justify-end gap-3">
                <button
                  onClick={() => setShowDoneModal(false)}
                  className="px-4 py-2 border border-secondary-300 text-secondary-700 rounded-lg hover:bg-secondary-50"
                >
                  Cancel
                </button>
                <button
                  onClick={handleMarkDone}
                  disabled={isSubmitting}
                  className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50"
                >
                  {isSubmitting ? 'Marking...' : 'Mark as Done'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Mark as Blocked Modal */}
      {showBlockedModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg max-w-md w-full p-6">
            <h3 className="text-lg font-bold text-secondary-900 mb-4">Mark Task as Blocked</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-secondary-900 mb-2">
                  Reason for blocking
                </label>
                <textarea
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  placeholder="What's blocking this task? What help do you need?"
                  rows={3}
                  className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500"
                />
              </div>
              <div className="flex justify-end gap-3">
                <button
                  onClick={() => setShowBlockedModal(false)}
                  className="px-4 py-2 border border-secondary-300 text-secondary-700 rounded-lg hover:bg-secondary-50"
                >
                  Cancel
                </button>
                <button
                  onClick={handleMarkBlocked}
                  disabled={isSubmitting}
                  className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 disabled:opacity-50"
                >
                  {isSubmitting ? 'Marking...' : 'Mark as Blocked'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Request Takeover Modal */}
      {showTakeoverModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg max-w-md w-full p-6">
            <h3 className="text-lg font-bold text-secondary-900 mb-4">Request Task Takeover</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-secondary-900 mb-2">
                  Reason for takeover request
                </label>
                <textarea
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  placeholder="Why do you need someone else to take over this task?"
                  rows={3}
                  className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500"
                />
              </div>
              <div className="flex justify-end gap-3">
                <button
                  onClick={() => setShowTakeoverModal(false)}
                  className="px-4 py-2 border border-secondary-300 text-secondary-700 rounded-lg hover:bg-secondary-50"
                >
                  Cancel
                </button>
                <button
                  onClick={handleRequestTakeover}
                  disabled={isSubmitting}
                  className="px-4 py-2 bg-orange-600 text-white rounded-lg hover:bg-orange-700 disabled:opacity-50"
                >
                  {isSubmitting ? 'Requesting...' : 'Request Takeover'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
