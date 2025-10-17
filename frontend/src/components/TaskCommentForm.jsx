import { useState } from 'react'
import { addTaskComment } from '../services/api'

export default function TaskCommentForm({ taskId, onCommentAdded }) {
  const [commentText, setCommentText] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (!commentText.trim()) return

    try {
      setIsSubmitting(true)
      await addTaskComment(taskId, commentText.trim())
      setCommentText('')
      onCommentAdded?.()
    } catch (error) {
      console.error('Error adding comment:', error)
      alert('Failed to add comment')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label htmlFor="comment" className="block text-sm font-medium text-secondary-900 mb-2">
          Add a comment
        </label>
        <textarea
          id="comment"
          value={commentText}
          onChange={(e) => setCommentText(e.target.value)}
          placeholder="Share your progress, ask questions, or provide updates..."
          rows={3}
          className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
          required
        />
      </div>
      
      <div className="flex justify-end">
        <button
          type="submit"
          disabled={!commentText.trim() || isSubmitting}
          className="btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isSubmitting ? 'Adding...' : 'Add Comment'}
        </button>
      </div>
    </form>
  )
}
