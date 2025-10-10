import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { useToast } from '../contexts/ToastContext'
import { 
  MapPinIcon, 
  CalendarIcon, 
  ClockIcon,
  UserGroupIcon,
  CheckCircleIcon,
  XCircleIcon
} from '@heroicons/react/24/outline'
import api from '../services/api'

export default function VolunteerPortal() {
  const [initiatives, setInitiatives] = useState([])
  const [loading, setLoading] = useState(true)
  const [statusFilter, setStatusFilter] = useState('active')
  const [locationFilter, setLocationFilter] = useState('')
  const [skillFilter, setSkillFilter] = useState('')
  const [appliedInitiatives, setAppliedInitiatives] = useState(new Set())

  const { user } = useAuth()
  const { showToast } = useToast()

  useEffect(() => {
    fetchInitiatives()
    fetchAppliedInitiatives()
  }, [statusFilter])

  const fetchInitiatives = async () => {
    try {
      setLoading(true)
      const params = new URLSearchParams()
      if (statusFilter) params.append('status', statusFilter)
      
      const response = await api.get(`/initiatives?${params.toString()}`)
      setInitiatives(response.data.initiatives || [])
    } catch (error) {
      showToast('Failed to load initiatives', 'error')
    } finally {
      setLoading(false)
    }
  }

  const fetchAppliedInitiatives = async () => {
    try {
      const response = await api.get('/applications')
      const appliedIds = new Set(response.data.applications?.map(app => app.initiative_id) || [])
      setAppliedInitiatives(appliedIds)
    } catch (error) {
      // Silently fail - applications might not be implemented yet
    }
  }

  const handleApply = async (initiativeId) => {
    try {
      await api.post('/applications', {
        initiative_id: initiativeId,
        volunteer_id: user.id
      })
      
      setAppliedInitiatives(prev => new Set([...prev, initiativeId]))
      showToast('Application submitted successfully!', 'success')
    } catch (error) {
      showToast(error.response?.data?.error || 'Failed to submit application', 'error')
    }
  }

  const filteredInitiatives = initiatives.filter(initiative => {
    if (locationFilter && !initiative.location_address?.toLowerCase().includes(locationFilter.toLowerCase())) {
      return false
    }
    
    if (skillFilter && !initiative.required_skills?.some(skill => 
      skill.toLowerCase().includes(skillFilter.toLowerCase())
    )) {
      return false
    }
    
    return true
  })

  const formatDate = (dateString) => {
    if (!dateString) return 'TBD'
    return new Date(dateString).toLocaleDateString()
  }

  const getStatusBadge = (status) => {
    const styles = {
      active: 'bg-green-100 text-green-800',
      closed: 'bg-red-100 text-red-800',
      draft: 'bg-gray-100 text-gray-800'
    }
    return (
      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${styles[status] || styles.draft}`}>
        {status.charAt(0).toUpperCase() + status.slice(1)}
      </span>
    )
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-secondary-900">
          Volunteer Opportunities
        </h1>
        <p className="text-secondary-600 mt-2">
          Find initiatives that match your skills and interests
        </p>
      </div>

      {/* Filters */}
      <div className="mb-6 grid grid-cols-1 md:grid-cols-4 gap-4">
        <div>
          <label htmlFor="statusFilter" className="block text-sm font-medium text-secondary-700 mb-1">
            Status
          </label>
          <select
            id="statusFilter"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="input-field"
          >
            <option value="">All Statuses</option>
            <option value="active">Active</option>
            <option value="closed">Closed</option>
          </select>
        </div>
        
        <div>
          <label htmlFor="locationFilter" className="block text-sm font-medium text-secondary-700 mb-1">
            Location
          </label>
          <input
            id="locationFilter"
            type="text"
            className="input-field"
            placeholder="Filter by location..."
            value={locationFilter}
            onChange={(e) => setLocationFilter(e.target.value)}
          />
        </div>
        
        <div>
          <label htmlFor="skillFilter" className="block text-sm font-medium text-secondary-700 mb-1">
            Skill
          </label>
          <input
            id="skillFilter"
            type="text"
            className="input-field"
            placeholder="Filter by skill..."
            value={skillFilter}
            onChange={(e) => setSkillFilter(e.target.value)}
          />
        </div>
        
        <div className="flex items-end">
          <button
            onClick={() => {
              setStatusFilter('')
              setLocationFilter('')
              setSkillFilter('')
            }}
            className="btn-secondary w-full"
          >
            Clear Filters
          </button>
        </div>
      </div>

      {/* Initiatives List */}
      {filteredInitiatives.length === 0 ? (
        <div className="text-center py-12">
          <div className="mx-auto h-12 w-12 text-secondary-400">
            <UserGroupIcon className="h-12 w-12" />
          </div>
          <h3 className="mt-2 text-sm font-medium text-secondary-900">No initiatives found</h3>
          <p className="mt-1 text-sm text-secondary-500">
            Try adjusting your filters or check back later for new opportunities.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
          {filteredInitiatives.map((initiative) => {
            const hasApplied = appliedInitiatives.has(initiative.id)
            
            return (
              <div key={initiative.id} className="card hover:shadow-lg transition-shadow">
                <div className="flex justify-between items-start mb-3">
                  <h3 className="text-lg font-semibold text-secondary-900 line-clamp-2">
                    {initiative.title}
                  </h3>
                  {getStatusBadge(initiative.status)}
                </div>
                
                <p className="text-secondary-600 text-sm line-clamp-3 mb-4">
                  {initiative.description}
                </p>
                
                <div className="space-y-2 mb-4">
                  {initiative.location_address && (
                    <div className="flex items-center text-sm text-secondary-500">
                      <MapPinIcon className="h-4 w-4 mr-1" />
                      {initiative.location_address}
                    </div>
                  )}
                  
                  <div className="flex items-center text-sm text-secondary-500">
                    <CalendarIcon className="h-4 w-4 mr-1" />
                    {formatDate(initiative.start_date)} - {formatDate(initiative.end_date)}
                  </div>
                  
                  {initiative.required_skills && initiative.required_skills.length > 0 && (
                    <div className="flex items-center text-sm text-secondary-500">
                      <UserGroupIcon className="h-4 w-4 mr-1" />
                      {initiative.required_skills.length} required skills
                    </div>
                  )}
                </div>
                
                {initiative.required_skills && initiative.required_skills.length > 0 && (
                  <div className="mb-4">
                    <div className="flex flex-wrap gap-1">
                      {initiative.required_skills.slice(0, 3).map(skill => (
                        <span 
                          key={skill}
                          className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-primary-100 text-primary-800"
                        >
                          {skill}
                        </span>
                      ))}
                      {initiative.required_skills.length > 3 && (
                        <span className="text-xs text-secondary-500">
                          +{initiative.required_skills.length - 3} more
                        </span>
                      )}
                    </div>
                  </div>
                )}
                
                <div className="flex justify-between items-center">
                  <Link
                    to={`/initiatives/${initiative.id}`}
                    className="text-primary-600 hover:text-primary-500 text-sm font-medium"
                  >
                    View Details →
                  </Link>
                  
                  {hasApplied ? (
                    <div className="flex items-center text-green-600 text-sm">
                      <CheckCircleIcon className="h-4 w-4 mr-1" />
                      Applied
                    </div>
                  ) : initiative.status === 'active' ? (
                    <button
                      onClick={() => handleApply(initiative.id)}
                      className="btn-primary text-sm"
                    >
                      Apply Now
                    </button>
                  ) : (
                    <button
                      disabled
                      className="btn-secondary text-sm opacity-50 cursor-not-allowed"
                    >
                      {initiative.status === 'closed' ? 'Closed' : 'Draft'}
                    </button>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      )}

      {/* Stats */}
      <div className="mt-8 grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="card text-center">
          <div className="text-2xl font-bold text-primary-600">
            {filteredInitiatives.filter(i => i.status === 'active').length}
          </div>
          <div className="text-sm text-secondary-500">Active Opportunities</div>
        </div>
        <div className="card text-center">
          <div className="text-2xl font-bold text-secondary-600">
            {appliedInitiatives.size}
          </div>
          <div className="text-sm text-secondary-500">Applications Submitted</div>
        </div>
        <div className="card text-center">
          <div className="text-2xl font-bold text-green-600">
            {filteredInitiatives.reduce((acc, i) => acc + (i.required_skills?.length || 0), 0)}
          </div>
          <div className="text-sm text-secondary-500">Total Skills Needed</div>
        </div>
      </div>
    </div>
  )
}
