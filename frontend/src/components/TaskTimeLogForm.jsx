import { useState } from 'react'
import { logTaskTime } from '../services/api'

export default function TaskTimeLogForm({ taskId, onTimeLogged }) {
  const [hours, setHours] = useState('')
  const [logDate, setLogDate] = useState(new Date().toISOString().split('T')[0])
  const [description, setDescription] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (!hours || parseFloat(hours) <= 0) return

    try {
      setIsSubmitting(true)
      await logTaskTime(taskId, {
        hours: parseFloat(hours),
        log_date: logDate,
        description: description.trim()
      })
      setHours('')
      setDescription('')
      setLogDate(new Date().toISOString().split('T')[0])
      onTimeLogged?.()
    } catch (error) {
      console.error('Error logging time:', error)
      alert('Failed to log time')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label htmlFor="hours" className="block text-sm font-medium text-secondary-900 mb-2">
            Hours *
          </label>
          <input
            id="hours"
            type="number"
            step="0.25"
            min="0.1"
            value={hours}
            onChange={(e) => setHours(e.target.value)}
            placeholder="2.5"
            className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
            required
          />
        </div>
        
        <div>
          <label htmlFor="logDate" className="block text-sm font-medium text-secondary-900 mb-2">
            Date *
          </label>
          <input
            id="logDate"
            type="date"
            value={logDate}
            onChange={(e) => setLogDate(e.target.value)}
            className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
            required
          />
        </div>
      </div>
      
      <div>
        <label htmlFor="description" className="block text-sm font-medium text-secondary-900 mb-2">
          Description
        </label>
        <textarea
          id="description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="What did you work on? (optional)"
          rows={2}
          className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
        />
      </div>
      
      <div className="flex justify-end">
        <button
          type="submit"
          disabled={!hours || parseFloat(hours) <= 0 || isSubmitting}
          className="btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isSubmitting ? 'Logging...' : 'Log Time'}
        </button>
      </div>
    </form>
  )
}
