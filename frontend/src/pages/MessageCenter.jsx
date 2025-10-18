import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { getInbox } from '../services/api'
import api from '../services/api'

export default function MessageCenter() {
  const { user } = useAuth()
  const [activeTab, setActiveTab] = useState('projects')
  const [unreadCounts, setUnreadCounts] = useState([])
  const [projects, setProjects] = useState({})
  const [directMessages, setDirectMessages] = useState([])
  const [directLoading, setDirectLoading] = useState(false)
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()

  useEffect(() => {
    fetchUnreadCounts()
    fetchDirectMessages()
  }, [])

  useEffect(() => {
    if (activeTab === 'direct' && directMessages.length === 0 && !directLoading) {
      fetchDirectMessages()
    }
  }, [activeTab])

  const fetchUnreadCounts = async () => {
    try {
      setLoading(true)
      const response = await api.get('/messages/unread-counts')
      const counts = response.data.unread_counts || []
      setUnreadCounts(counts)

      // Fetch project details for each project with unread messages
      const projectPromises = counts.map(c => 
        api.get(`/projects/${c.project_id}`).catch(err => null)
      )
      const projectResponses = await Promise.all(projectPromises)
      
      const projectsMap = {}
      projectResponses.forEach((res, idx) => {
        if (res?.data) {
          projectsMap[counts[idx].project_id] = res.data
        }
      })
      setProjects(projectsMap)
    } catch (error) {
      console.error('Error fetching unread counts:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchDirectMessages = async () => {
    try {
      setDirectLoading(true)
      const response = await getInbox({ limit: 50, offset: 0 })
      setDirectMessages(response.data.messages || [])
    } catch (error) {
      console.error('Error fetching direct messages:', error)
    } finally {
      setDirectLoading(false)
    }
  }

  const formatTime = (dateString) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now - date
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMs / 3600000)
    const diffDays = Math.floor(diffMs / 86400000)

    if (diffMins < 1) return 'Just now'
    if (diffMins < 60) return `${diffMins}m ago`
    if (diffHours < 24) return `${diffHours}h ago`
    if (diffDays < 7) return `${diffDays}d ago`
    return date.toLocaleDateString()
  }

  const getMessageSnippet = (message) => {
    const text = message.subject || message.message_text || ''
    return text.length > 50 ? text.substring(0, 50) + '...' : text
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-secondary-50">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-secondary-50 py-8">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-secondary-900">Message Center</h1>
          <p className="text-secondary-600 mt-2">
            View all your messages in one place
          </p>
        </div>

        {/* Tab Navigation */}
        <div className="mb-6">
          <div className="border-b border-secondary-200">
            <nav className="-mb-px flex space-x-8">
              <button
                onClick={() => setActiveTab('projects')}
                className={`py-2 px-1 border-b-2 font-medium text-sm ${
                  activeTab === 'projects'
                    ? 'border-primary-500 text-primary-600'
                    : 'border-transparent text-secondary-500 hover:text-secondary-700 hover:border-secondary-300'
                }`}
              >
                Project Messages
                {unreadCounts.length > 0 && (
                  <span className="ml-2 inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-primary-100 text-primary-800">
                    {unreadCounts.reduce((sum, count) => sum + count.count, 0)}
                  </span>
                )}
              </button>
              <button
                onClick={() => setActiveTab('direct')}
                className={`py-2 px-1 border-b-2 font-medium text-sm ${
                  activeTab === 'direct'
                    ? 'border-primary-500 text-primary-600'
                    : 'border-transparent text-secondary-500 hover:text-secondary-700 hover:border-secondary-300'
                }`}
              >
                Direct Messages
                {directMessages.filter(msg => !msg.is_read).length > 0 && (
                  <span className="ml-2 inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-primary-100 text-primary-800">
                    {directMessages.filter(msg => !msg.is_read).length}
                  </span>
                )}
              </button>
            </nav>
          </div>
        </div>

        {/* Tab Content */}
        {activeTab === 'projects' ? (
          // Projects Tab
          unreadCounts.length > 0 ? (
            <div className="space-y-3">
              {unreadCounts.map((count) => {
                const project = projects[count.project_id]
                return (
                  <div
                    key={count.project_id}
                    onClick={() => navigate(`/projects/${count.project_id}/messages`)}
                    className="bg-white rounded-lg border border-secondary-200 p-4 hover:shadow-md transition-shadow cursor-pointer"
                  >
                    <div className="flex justify-between items-center">
                      <div className="flex-1">
                        <h3 className="font-semibold text-secondary-900">
                          {project?.title || `Project ${count.project_id.slice(0, 8)}...`}
                        </h3>
                        {project?.description && (
                          <p className="text-sm text-secondary-600 mt-1 line-clamp-1">
                            {project.description}
                          </p>
                        )}
                      </div>
                      <div className="ml-4">
                        <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-primary-100 text-primary-800">
                          {count.count} unread
                        </span>
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>
          ) : (
            <div className="bg-white rounded-lg border border-secondary-200 p-12 text-center">
              <div className="text-6xl mb-4">💬</div>
              <h3 className="text-lg font-medium text-secondary-900 mb-2">
                No Unread Messages
              </h3>
              <p className="text-secondary-600 mb-6">
                You're all caught up! All your project messages have been read.
              </p>
              <button
                onClick={() => navigate('/projects')}
                className="btn-primary"
              >
                Browse Projects
              </button>
            </div>
          )
        ) : (
          // Direct Messages Tab
          directLoading ? (
            <div className="flex items-center justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
            </div>
          ) : directMessages.length > 0 ? (
            <div className="space-y-3">
              {directMessages.map((message) => (
                <div
                  key={message.id}
                  onClick={() => navigate(`/messages/direct/${message.sender_id}`)}
                  className="bg-white rounded-lg border border-secondary-200 p-4 hover:shadow-md transition-shadow cursor-pointer"
                >
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-1">
                        <h3 className="font-semibold text-secondary-900">
                          {message.sender_name}
                        </h3>
                        {!message.is_read && (
                          <div className="w-2 h-2 bg-primary-600 rounded-full"></div>
                        )}
                      </div>
                      <p className="text-sm text-secondary-600 mb-1">
                        {getMessageSnippet(message)}
                      </p>
                      <p className="text-xs text-secondary-500">
                        {formatTime(message.created_at)}
                      </p>
                    </div>
                    {!message.is_read && (
                      <div className="ml-4">
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-primary-100 text-primary-800">
                          New
                        </span>
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="bg-white rounded-lg border border-secondary-200 p-12 text-center">
              <div className="text-6xl mb-4">📩</div>
              <h3 className="text-lg font-medium text-secondary-900 mb-2">
                No Direct Messages Yet
              </h3>
              <p className="text-secondary-600 mb-6">
                You haven't received any direct messages yet. Start a conversation with someone!
              </p>
            </div>
          )
        )}
      </div>
    </div>
  )
}

