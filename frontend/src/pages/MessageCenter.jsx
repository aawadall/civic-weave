import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import api from '../services/api'

export default function MessageCenter() {
  const { user } = useAuth()
  const [unreadCounts, setUnreadCounts] = useState([])
  const [projects, setProjects] = useState({})
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()

  useEffect(() => {
    fetchUnreadCounts()
  }, [])

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
            View all your project messages in one place
          </p>
        </div>

        {unreadCounts.length > 0 ? (
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
        )}
      </div>
    </div>
  )
}

