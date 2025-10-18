import { useState, useEffect } from 'react'
import { getTaskComments } from '../services/api'
import TaskCommentForm from './TaskCommentForm'
import TaskTimeLogForm from './TaskTimeLogForm'
import TaskStatusActions from './TaskStatusActions'
import TaskTimeSummary from './TaskTimeSummary'

export default function TaskDetailModal({ task, isOpen, onClose, onTaskUpdated }) {
  const [comments, setComments] = useState([])
  const [loading, setLoading] = useState(false)
  const [activeTab, setActiveTab] = useState('overview')

  useEffect(() => {
    if (isOpen && task) {
      fetchComments()
    }
  }, [isOpen, task])

  const fetchComments = async () => {
    try {
      setLoading(true)
      const response = await getTaskComments(task.id)
      setComments(response.data.comments || [])
    } catch (error) {
      console.error('Error fetching comments:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleCommentAdded = () => {
    fetchComments()
  }

  const handleTimeLogged = () => {
    // Refresh the time summary
    onTaskUpdated?.()
  }

  const handleStatusChanged = () => {
    onTaskUpdated?.()
  }

  if (!isOpen || !task) return null

  const getStatusBadge = (status) => {
    const statusConfig = {
      todo: { bg: 'bg-gray-100', text: 'text-gray-800', label: 'To Do' },
      in_progress: { bg: 'bg-blue-100', text: 'text-blue-800', label: 'In Progress' },
      done: { bg: 'bg-green-100', text: 'text-green-800', label: 'Done' },
      blocked: { bg: 'bg-red-100', text: 'text-red-800', label: 'Blocked' },
      takeover_requested: { bg: 'bg-orange-100', text: 'text-orange-800', label: 'Takeover Requested' }
    }
    
    const config = statusConfig[status] || statusConfig.todo
    return (
      <span className={`px-3 py-1 text-sm font-medium rounded-full ${config.bg} ${config.text}`}>
        {config.label}
      </span>
    )
  }

  const getPriorityBadge = (priority) => {
    const priorityConfig = {
      low: { bg: 'bg-green-100', text: 'text-green-800', label: 'Low' },
      medium: { bg: 'bg-yellow-100', text: 'text-yellow-800', label: 'Medium' },
      high: { bg: 'bg-red-100', text: 'text-red-800', label: 'High' }
    }
    
    const config = priorityConfig[priority] || priorityConfig.medium
    return (
      <span className={`px-2 py-1 text-xs font-medium rounded-full ${config.bg} ${config.text}`}>
        {config.label}
      </span>
    )
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg max-w-4xl w-full max-h-[90vh] overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-secondary-200">
          <div className="flex-1">
            <h2 className="text-2xl font-bold text-secondary-900">{task.title}</h2>
            <div className="flex items-center gap-3 mt-2">
              {getStatusBadge(task.status)}
              {getPriorityBadge(task.priority)}
              {task.due_date && (
                <span className="text-sm text-secondary-600">
                  Due: {new Date(task.due_date).toLocaleDateString()}
                </span>
              )}
            </div>
          </div>
          <button
            onClick={onClose}
            className="text-secondary-400 hover:text-secondary-600 text-2xl"
          >
            ×
          </button>
        </div>

        {/* Content */}
        <div className="flex h-[calc(90vh-120px)]">
          {/* Main Content */}
          <div className="flex-1 p-6 overflow-y-auto">
            {/* Tabs */}
            <div className="border-b border-secondary-200 mb-6">
              <nav className="flex space-x-8">
                {['overview', 'comments', 'time', 'actions'].map((tab) => (
                  <button
                    key={tab}
                    onClick={() => setActiveTab(tab)}
                    className={`py-2 px-1 border-b-2 font-medium text-sm ${
                      activeTab === tab
                        ? 'border-primary-500 text-primary-600'
                        : 'border-transparent text-secondary-500 hover:text-secondary-700 hover:border-secondary-300'
                    }`}
                  >
                    {tab.charAt(0).toUpperCase() + tab.slice(1)}
                  </button>
                ))}
              </nav>
            </div>

            {/* Tab Content */}
            {activeTab === 'overview' && (
              <div className="space-y-6">
                <div>
                  <h3 className="text-lg font-semibold text-secondary-900 mb-3">Description</h3>
                  <p className="text-secondary-700 whitespace-pre-wrap">
                    {task.description || 'No description provided'}
                  </p>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div>
                    <h3 className="text-lg font-semibold text-secondary-900 mb-3">Task Details</h3>
                    <dl className="space-y-2">
                      <div>
                        <dt className="text-sm font-medium text-secondary-500">Created</dt>
                        <dd className="text-sm text-secondary-900">
                          {new Date(task.created_at).toLocaleDateString()}
                        </dd>
                      </div>
                      {task.assignee_id && (
                        <div>
                          <dt className="text-sm font-medium text-secondary-500">Assignee</dt>
                          <dd className="text-sm text-secondary-900">
                            {task.assignee_name || 'Unknown'}
                            {task.assignee_email && (
                              <span className="text-secondary-500 ml-1">({task.assignee_email})</span>
                            )}
                          </dd>
                        </div>
                      )}
                      {task.project_title && (
                        <div>
                          <dt className="text-sm font-medium text-secondary-500">Project</dt>
                          <dd className="text-sm text-secondary-900">
                            {task.project_title}
                            {task.project_status && (
                              <span className={`ml-2 px-2 py-0.5 rounded text-xs ${
                                task.project_status === 'active' ? 'bg-green-100 text-green-700' :
                                task.project_status === 'recruiting' ? 'bg-blue-100 text-blue-700' :
                                task.project_status === 'completed' ? 'bg-gray-100 text-gray-700' :
                                'bg-yellow-100 text-yellow-700'
                              }`}>
                                {task.project_status}
                              </span>
                            )}
                          </dd>
                        </div>
                      )}
                      {task.labels && task.labels.length > 0 && (
                        <div>
                          <dt className="text-sm font-medium text-secondary-500">Labels</dt>
                          <dd className="flex flex-wrap gap-1 mt-1">
                            {task.labels.map((label, index) => (
                              <span
                                key={index}
                                className="px-2 py-1 bg-secondary-100 text-secondary-700 text-xs rounded-full"
                              >
                                {label}
                              </span>
                            ))}
                          </dd>
                        </div>
                      )}
                    </dl>
                  </div>

                  <div>
                    <TaskTimeSummary taskId={task.id} />
                  </div>
                </div>

                {/* Timeline Section */}
                {(task.started_at || task.blocked_at || task.completed_at || task.takeover_requested_at) && (
                  <div>
                    <h3 className="text-lg font-semibold text-secondary-900 mb-3">Timeline</h3>
                    <div className="space-y-3">
                      {task.started_at && (
                        <div className="flex items-center space-x-3 p-3 bg-blue-50 rounded-lg">
                          <div className="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
                            <span className="text-blue-600 text-sm">🚀</span>
                          </div>
                          <div>
                            <p className="text-sm font-medium text-blue-900">Task Started</p>
                            <p className="text-xs text-blue-700">{new Date(task.started_at).toLocaleString()}</p>
                          </div>
                        </div>
                      )}
                      {task.blocked_at && (
                        <div className="flex items-center space-x-3 p-3 bg-red-50 rounded-lg">
                          <div className="w-8 h-8 bg-red-100 rounded-full flex items-center justify-center">
                            <span className="text-red-600 text-sm">🚫</span>
                          </div>
                          <div>
                            <p className="text-sm font-medium text-red-900">Task Blocked</p>
                            <p className="text-xs text-red-700">
                              {new Date(task.blocked_at).toLocaleString()}
                              {task.blocked_reason && ` - ${task.blocked_reason}`}
                            </p>
                          </div>
                        </div>
                      )}
                      {task.completed_at && (
                        <div className="flex items-center space-x-3 p-3 bg-green-50 rounded-lg">
                          <div className="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center">
                            <span className="text-green-600 text-sm">✅</span>
                          </div>
                          <div>
                            <p className="text-sm font-medium text-green-900">Task Completed</p>
                            <p className="text-xs text-green-700">
                              {new Date(task.completed_at).toLocaleString()}
                              {task.completion_note && ` - ${task.completion_note}`}
                            </p>
                          </div>
                        </div>
                      )}
                      {task.takeover_requested_at && (
                        <div className="flex items-center space-x-3 p-3 bg-orange-50 rounded-lg">
                          <div className="w-8 h-8 bg-orange-100 rounded-full flex items-center justify-center">
                            <span className="text-orange-600 text-sm">🔄</span>
                          </div>
                          <div>
                            <p className="text-sm font-medium text-orange-900">Takeover Requested</p>
                            <p className="text-xs text-orange-700">
                              {new Date(task.takeover_requested_at).toLocaleString()}
                              {task.takeover_reason && ` - ${task.takeover_reason}`}
                            </p>
                          </div>
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </div>
            )}

            {activeTab === 'comments' && (
              <div className="space-y-6">
                <div>
                  <h3 className="text-lg font-semibold text-secondary-900 mb-4">Comments & Updates</h3>
                  <TaskCommentForm taskId={task.id} onCommentAdded={handleCommentAdded} />
                </div>

                <div className="space-y-4">
                  {loading ? (
                    <div className="flex items-center justify-center py-8">
                      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
                    </div>
                  ) : comments.length === 0 ? (
                    <div className="text-center py-8 text-secondary-500">
                      <p>No comments yet</p>
                      <p className="text-sm">Be the first to add a comment!</p>
                    </div>
                  ) : (
                    comments.map((comment) => (
                      <div key={comment.id} className="bg-secondary-50 rounded-lg p-4">
                        <div className="flex items-center justify-between mb-2">
                          <div className="flex items-center space-x-2">
                            <span className="font-medium text-secondary-900">{comment.user_name}</span>
                            <span className="text-sm text-secondary-500">
                              {new Date(comment.created_at).toLocaleString()}
                            </span>
                          </div>
                        </div>
                        <p className="text-secondary-700 whitespace-pre-wrap">{comment.comment_text}</p>
                      </div>
                    ))
                  )}
                </div>
              </div>
            )}

            {activeTab === 'time' && (
              <div className="space-y-6">
                <div>
                  <h3 className="text-lg font-semibold text-secondary-900 mb-4">Log Time</h3>
                  <TaskTimeLogForm taskId={task.id} onTimeLogged={handleTimeLogged} />
                </div>

                <div>
                  <TaskTimeSummary taskId={task.id} />
                </div>
              </div>
            )}

            {activeTab === 'actions' && (
              <div>
                <TaskStatusActions task={task} onStatusChanged={handleStatusChanged} />
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
