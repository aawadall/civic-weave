import { useState, useEffect } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'
import api from '../../services/api'

export default function ProjectsListPage() {
  const { user, hasAnyRole } = useAuth()
  const location = useLocation()
  const [projects, setProjects] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState('')

  useEffect(() => {
    fetchProjects()
  }, [])

  // Refresh projects when returning from detail page (detected by state change)
  useEffect(() => {
    if (location.state?.refreshProjects) {
      fetchProjects()
      // Clear the refresh flag to prevent unnecessary re-fetches
      window.history.replaceState({}, document.title, location.pathname)
    }
  }, [location])

  const fetchProjects = async () => {
    try {
      setLoading(true)
      const response = await api.get('/projects')
      setProjects(response.data.projects || [])
    } catch (err) {
      setError('Failed to fetch projects')
      console.error('Error fetching projects:', err)
    } finally {
      setLoading(false)
    }
  }

  const filteredProjects = projects.filter(project => {
    const matchesSearch = project.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         project.description.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesStatus = !statusFilter || project.status === statusFilter
    return matchesSearch && matchesStatus
  })

  const canCreateProject = hasAnyRole('team_lead', 'admin')
  const canManageProject = (project) => {
    return hasAnyRole('admin') || 
           (hasAnyRole('team_lead') && project.created_by_user_id === user?.id)
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-secondary-900 mb-4">Error</h2>
          <p className="text-secondary-600">{error}</p>
          <button 
            onClick={fetchProjects}
            className="mt-4 btn-primary"
          >
            Try Again
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-secondary-50 py-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mb-8">
          <div className="flex justify-between items-center mb-6">
            <h1 className="text-3xl font-bold text-secondary-900">Projects</h1>
            {canCreateProject && (
              <Link to="/projects/create" className="btn-primary">
                Create Project
              </Link>
            )}
          </div>

          {/* Search and Filter */}
          <div className="flex flex-col sm:flex-row gap-4 mb-6">
            <div className="flex-1">
              <input
                type="text"
                placeholder="Search projects..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              />
            </div>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            >
              <option value="">All Statuses</option>
              <option value="draft">Draft</option>
              <option value="recruiting">Recruiting</option>
              <option value="active">Active</option>
              <option value="completed">Completed</option>
              <option value="archived">Archived</option>
            </select>
          </div>
        </div>

        {/* Projects Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredProjects.map((project) => (
            <div key={project.id} className="bg-white rounded-lg shadow-sm border border-secondary-200 p-6">
              <div className="flex justify-between items-start mb-4">
                <h3 className="text-lg font-semibold text-secondary-900 line-clamp-2">
                  {project.title}
                </h3>
                <span className={`px-2 py-1 text-xs font-medium rounded-full ${
                  project.status === 'active' ? 'bg-green-100 text-green-800' :
                  project.status === 'recruiting' ? 'bg-blue-100 text-blue-800' :
                  project.status === 'completed' ? 'bg-gray-100 text-gray-800' :
                  'bg-yellow-100 text-yellow-800'
                }`}>
                  {project.status}
                </span>
              </div>

              <p className="text-secondary-600 text-sm mb-4 line-clamp-3">
                {project.description}
              </p>

              {project.required_skills && project.required_skills.length > 0 && (
                <div className="mb-4">
                  <h4 className="text-sm font-medium text-secondary-900 mb-2">Required Skills:</h4>
                  <div className="flex flex-wrap gap-1">
                    {project.required_skills.slice(0, 3).map((skill, index) => (
                      <span key={index} className="px-2 py-1 bg-primary-100 text-primary-800 text-xs rounded">
                        {skill}
                      </span>
                    ))}
                    {project.required_skills.length > 3 && (
                      <span className="px-2 py-1 bg-secondary-100 text-secondary-600 text-xs rounded">
                        +{project.required_skills.length - 3} more
                      </span>
                    )}
                  </div>
                </div>
              )}

              {project.location_address && (
                <p className="text-secondary-500 text-sm mb-4">
                  📍 {project.location_address}
                </p>
              )}

              <div className="flex justify-between items-center">
                <div className="text-sm text-secondary-500">
                  {project.start_date && (
                    <span>Starts: {new Date(project.start_date).toLocaleDateString()}</span>
                  )}
                </div>
                <div className="flex gap-2">
                  <Link 
                    to={`/projects/${project.id}`}
                    className="text-primary-600 hover:text-primary-800 text-sm font-medium"
                  >
                    View Details
                  </Link>
                  {canManageProject(project) && (
                    <Link 
                      to={`/projects/${project.id}/edit`}
                      className="text-secondary-600 hover:text-secondary-800 text-sm font-medium"
                    >
                      Edit
                    </Link>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>

        {filteredProjects.length === 0 && (
          <div className="text-center py-12">
            <h3 className="text-lg font-medium text-secondary-900 mb-2">No projects found</h3>
            <p className="text-secondary-600 mb-4">
              {searchTerm || statusFilter 
                ? 'Try adjusting your search or filter criteria.'
                : 'No projects are available at the moment.'
              }
            </p>
            {canCreateProject && (
              <Link to="/projects/create" className="btn-primary">
                Create First Project
              </Link>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
