import { useState, useEffect } from 'react'
import { getTaskTimeLogs } from '../services/api'

export default function TaskTimeSummary({ taskId }) {
  const [timeLogs, setTimeLogs] = useState([])
  const [loading, setLoading] = useState(true)
  const [totalHours, setTotalHours] = useState(0)

  useEffect(() => {
    fetchTimeLogs()
  }, [taskId])

  const fetchTimeLogs = async () => {
    try {
      setLoading(true)
      const response = await getTaskTimeLogs(taskId)
      setTimeLogs(response.data.time_logs || [])
      
      // Calculate total hours
      const total = response.data.time_logs?.reduce((sum, log) => sum + log.hours, 0) || 0
      setTotalHours(total)
    } catch (error) {
      console.error('Error fetching time logs:', error)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-4">
        <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold text-secondary-900">Time Logged</h3>
        <div className="text-2xl font-bold text-primary-600">
          {totalHours.toFixed(1)}h
        </div>
      </div>

      {timeLogs.length === 0 ? (
        <div className="text-center py-8 text-secondary-500">
          <p>No time logged yet</p>
          <p className="text-sm">Start logging time to track your progress</p>
        </div>
      ) : (
        <div className="space-y-3">
          {timeLogs.map((log) => (
            <div key={log.id} className="bg-secondary-50 rounded-lg p-4">
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center space-x-2">
                  <span className="font-medium text-secondary-900">{log.volunteer_name}</span>
                  <span className="text-sm text-secondary-500">
                    {new Date(log.log_date).toLocaleDateString()}
                  </span>
                </div>
                <span className="font-semibold text-primary-600">
                  {log.hours}h
                </span>
              </div>
              {log.description && (
                <p className="text-sm text-secondary-600">{log.description}</p>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
