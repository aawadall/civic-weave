import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'
import api from '../../services/api'

export default function VolunteersPoolPage() {
  const { hasAnyRole } = useAuth()
  const [volunteers, setVolunteers] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [searchTerm, setSearchTerm] = useState('')
  const [skillFilter, setSkillFilter] = useState('')
  const [sortBy, setSortBy] = useState('name')
  const [showTopRated, setShowTopRated] = useState(false)

  useEffect(() => {
    fetchVolunteers()
  }, [showTopRated])

  const fetchVolunteers = async () => {
    try {
      setLoading(true)
      const endpoint = showTopRated ? '/volunteers/top-rated' : '/volunteers'
      const response = await api.get(endpoint)
      setVolunteers(response.data.volunteers || [])
    } catch (err) {
      setError('Failed to fetch volunteers')
      console.error('Error fetching volunteers:', err)
    } finally {
      setLoading(false)
    }
  }

  const filteredVolunteers = volunteers.filter(volunteer => {
    const matchesSearch = volunteer.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         volunteer.email?.toLowerCase().includes(searchTerm.toLowerCase())
    const matchesSkill = !skillFilter || 
                        volunteer.skills?.some(skill => 
                          skill.toLowerCase().includes(skillFilter.toLowerCase())
                        )
    return matchesSearch && matchesSkill
  })

  const sortedVolunteers = [...filteredVolunteers].sort((a, b) => {
    switch (sortBy) {
      case 'name':
        return (a.name || '').localeCompare(b.name || '')
      case 'rating':
        return (b.overall_score || 0) - (a.overall_score || 0)
      case 'ratings_count':
        return (b.total_ratings || 0) - (a.total_ratings || 0)
      default:
        return 0
    }
  })

  // Check if user has permission to view volunteers
  if (!hasAnyRole('team_lead', 'campaign_manager', 'admin')) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-secondary-900 mb-4">Access Denied</h2>
          <p className="text-secondary-600 mb-4">You don't have permission to view the volunteer pool.</p>
          <Link to="/dashboard" className="btn-primary">
            Back to Dashboard
          </Link>
        </div>
      </div>
    )
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
            onClick={fetchVolunteers}
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
          <h1 className="text-3xl font-bold text-secondary-900 mb-2">Volunteer Pool</h1>
          <p className="text-secondary-600">Browse and manage volunteers in your organization.</p>
        </div>

        {/* Filters and Controls */}
        <div className="bg-white rounded-lg shadow-sm border border-secondary-200 p-6 mb-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
            <div>
              <label className="block text-sm font-medium text-secondary-900 mb-2">
                Search
              </label>
              <input
                type="text"
                placeholder="Search volunteers..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-secondary-900 mb-2">
                Filter by Skill
              </label>
              <input
                type="text"
                placeholder="Filter by skill..."
                value={skillFilter}
                onChange={(e) => setSkillFilter(e.target.value)}
                className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-secondary-900 mb-2">
                Sort By
              </label>
              <select
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value)}
                className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              >
                <option value="name">Name</option>
                <option value="rating">Rating</option>
                <option value="ratings_count">Number of Ratings</option>
              </select>
            </div>
            <div className="flex items-end">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={showTopRated}
                  onChange={(e) => setShowTopRated(e.target.checked)}
                  className="mr-2"
                />
                <span className="text-sm text-secondary-900">Show top-rated only</span>
              </label>
            </div>
          </div>
        </div>

        {/* Volunteers Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {sortedVolunteers.map((volunteer) => (
            <div key={volunteer.id} className="bg-white rounded-lg shadow-sm border border-secondary-200 p-6">
              <div className="flex justify-between items-start mb-4">
                <div>
                  <h3 className="text-lg font-semibold text-secondary-900">
                    {volunteer.name || 'Unnamed Volunteer'}
                  </h3>
                  <p className="text-secondary-600 text-sm">{volunteer.email}</p>
                </div>
                {volunteer.overall_score !== undefined && (
                  <div className="text-right">
                    <div className="text-lg font-bold text-primary-600">
                      {volunteer.overall_score.toFixed(1)}
                    </div>
                    <div className="text-xs text-secondary-500">
                      {volunteer.total_ratings} ratings
                    </div>
                  </div>
                )}
              </div>

              {volunteer.skills && volunteer.skills.length > 0 && (
                <div className="mb-4">
                  <h4 className="text-sm font-medium text-secondary-900 mb-2">Skills:</h4>
                  <div className="flex flex-wrap gap-1">
                    {volunteer.skills.slice(0, 3).map((skill, index) => (
                      <span key={index} className="px-2 py-1 bg-primary-100 text-primary-800 text-xs rounded">
                        {skill}
                      </span>
                    ))}
                    {volunteer.skills.length > 3 && (
                      <span className="px-2 py-1 bg-secondary-100 text-secondary-600 text-xs rounded">
                        +{volunteer.skills.length - 3} more
                      </span>
                    )}
                  </div>
                </div>
              )}

              <div className="flex justify-between items-center">
                <div className="text-sm text-secondary-500">
                  Joined: {new Date(volunteer.created_at).toLocaleDateString()}
                </div>
                <div className="flex gap-2">
                  <Link 
                    to={`/volunteers/${volunteer.id}/scorecard`}
                    className="text-primary-600 hover:text-primary-800 text-sm font-medium"
                  >
                    View Scorecard
                  </Link>
                  <Link 
                    to={`/volunteers/${volunteer.id}`}
                    className="text-secondary-600 hover:text-secondary-800 text-sm font-medium"
                  >
                    View Profile
                  </Link>
                </div>
              </div>
            </div>
          ))}
        </div>

        {sortedVolunteers.length === 0 && (
          <div className="text-center py-12">
            <h3 className="text-lg font-medium text-secondary-900 mb-2">No volunteers found</h3>
            <p className="text-secondary-600">
              {searchTerm || skillFilter 
                ? 'Try adjusting your search or filter criteria.'
                : 'No volunteers are available at the moment.'
              }
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
