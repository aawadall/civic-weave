import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { useToast } from '../contexts/ToastContext'
import { 
  MapPinIcon, 
  CalendarIcon, 
  ClockIcon,
  UserGroupIcon,
  ArrowLeftIcon,
  CheckCircleIcon
} from '@heroicons/react/24/outline'
import api from '../services/api'

export default function InitiativeDetailPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { user } = useAuth()
  const { showToast } = useToast()
  
  const [initiative, setInitiative] = useState(null)
  const [loading, setLoading] = useState(true)
  const [applying, setApplying] = useState(false)
  const [hasApplied, setHasApplied] = useState(false)

  useEffect(() => {
    fetchInitiative()
    checkApplicationStatus()
  }, [id])

  const fetchInitiative = async () => {
    try {
      setLoading(true)
      const response = await api.get(`/initiatives/${id}`)
      setInitiative(response.data)
    } catch (error) {
      showToast('Failed to load initiative details', 'error')
      navigate('/volunteer')
    } finally {
      setLoading(false)
    }
  }

  const checkApplicationStatus = async () => {
    try {
      const response = await api.get('/applications')
      const applications = response.data.applications || []
      const userApplication = applications.find(app => 
        app.initiative_id === id && app.volunteer_id === user.id
      )
      setHasApplied(!!userApplication)
    } catch (error) {
      // Silently fail - applications might not be implemented yet
    }
  }

  const handleApply = async () => {
    if (!user) {
      showToast('Please log in to apply', 'error')
      navigate('/login')
      return
    }

    try {
      setApplying(true)
      await api.post('/applications', {
        initiative_id: id,
        volunteer_id: user.id
      })
      
      setHasApplied(true)
      showToast('Application submitted successfully!', 'success')
    } catch (error) {
      showToast(error.response?.data?.error || 'Failed to submit application', 'error')
    } finally {
      setApplying(false)
    }
  }

  const formatDate = (dateString) => {
    if (!dateString) return 'TBD'
    return new Date(dateString).toLocaleDateString('en-US', {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    })
  }

  const formatDateRange = () => {
    if (!initiative.start_date && !initiative.end_date) return 'Dates TBD'
    if (!initiative.start_date) return `Until ${formatDate(initiative.end_date)}`
    if (!initiative.end_date) return `Starting ${formatDate(initiative.start_date)}`
    if (initiative.start_date === initiative.end_date) {
      return formatDate(initiative.start_date)
    }
    return `${formatDate(initiative.start_date)} - ${formatDate(initiative.end_date)}`
  }

  const getStatusBadge = (status) => {
    const styles = {
      active: 'bg-green-100 text-green-800',
      closed: 'bg-red-100 text-red-800',
      draft: 'bg-gray-100 text-gray-800'
    }
    return (
      <span className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${styles[status] || styles.draft}`}>
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

  if (!initiative) {
    return (
      <div className="text-center py-12">
        <h2 className="text-2xl font-bold text-secondary-900">Initiative not found</h2>
        <p className="text-secondary-600 mt-2">The initiative you're looking for doesn't exist.</p>
        <Link to="/volunteer" className="btn-primary mt-4 inline-flex items-center">
          <ArrowLeftIcon className="h-4 w-4 mr-2" />
          Back to Opportunities
        </Link>
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-6">
        <button
          onClick={() => navigate('/volunteer')}
          className="inline-flex items-center text-primary-600 hover:text-primary-500 mb-4"
        >
          <ArrowLeftIcon className="h-4 w-4 mr-1" />
          Back to Opportunities
        </button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Main Content */}
        <div className="lg:col-span-2">
          <div className="card">
            <div className="flex justify-between items-start mb-6">
              <h1 className="text-3xl font-bold text-secondary-900">
                {initiative.title}
              </h1>
              {getStatusBadge(initiative.status)}
            </div>

            <div className="prose max-w-none">
              <p className="text-lg text-secondary-600 mb-6">
                {initiative.description}
              </p>
            </div>

            {/* Initiative Details */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mt-8">
              {initiative.location_address && (
                <div className="flex items-start space-x-3">
                  <MapPinIcon className="h-6 w-6 text-secondary-400 mt-0.5" />
                  <div>
                    <h3 className="font-semibold text-secondary-900">Location</h3>
                    <p className="text-secondary-600">{initiative.location_address}</p>
                  </div>
                </div>
              )}

              <div className="flex items-start space-x-3">
                <CalendarIcon className="h-6 w-6 text-secondary-400 mt-0.5" />
                <div>
                  <h3 className="font-semibold text-secondary-900">Duration</h3>
                  <p className="text-secondary-600">{formatDateRange()}</p>
                </div>
              </div>

              {initiative.required_skills && initiative.required_skills.length > 0 && (
                <div className="md:col-span-2 flex items-start space-x-3">
                  <UserGroupIcon className="h-6 w-6 text-secondary-400 mt-0.5" />
                  <div className="flex-1">
                    <h3 className="font-semibold text-secondary-900 mb-2">Required Skills</h3>
                    <div className="flex flex-wrap gap-2">
                      {initiative.required_skills.map(skill => (
                        <span 
                          key={skill}
                          className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-primary-100 text-primary-800"
                        >
                          {skill}
                        </span>
                      ))}
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Sidebar */}
        <div className="lg:col-span-1">
          <div className="card sticky top-8">
            <h3 className="text-lg font-semibold text-secondary-900 mb-4">
              Apply to Volunteer
            </h3>

            {hasApplied ? (
              <div className="text-center py-6">
                <CheckCircleIcon className="h-12 w-12 text-green-500 mx-auto mb-3" />
                <h4 className="font-semibold text-secondary-900 mb-2">
                  Application Submitted!
                </h4>
                <p className="text-sm text-secondary-600">
                  Your application has been received. We'll review it and get back to you soon.
                </p>
              </div>
            ) : initiative.status === 'active' ? (
              <div>
                <p className="text-sm text-secondary-600 mb-4">
                  Ready to make a difference? Apply now to volunteer for this initiative.
                </p>
                <button
                  onClick={handleApply}
                  disabled={applying}
                  className="w-full btn-primary"
                >
                  {applying ? 'Applying...' : 'Apply Now'}
                </button>
              </div>
            ) : (
              <div className="text-center py-6">
                <div className="text-secondary-400 mb-3">
                  {initiative.status === 'closed' ? (
                    <ClockIcon className="h-12 w-12 mx-auto" />
                  ) : (
                    <UserGroupIcon className="h-12 w-12 mx-auto" />
                  )}
                </div>
                <h4 className="font-semibold text-secondary-900 mb-2">
                  {initiative.status === 'closed' ? 'Applications Closed' : 'Not Yet Available'}
                </h4>
                <p className="text-sm text-secondary-600">
                  {initiative.status === 'closed' 
                    ? 'This initiative is no longer accepting applications.'
                    : 'This initiative is still being prepared.'
                  }
                </p>
              </div>
            )}

            <div className="mt-6 pt-6 border-t border-secondary-200">
              <h4 className="font-semibold text-secondary-900 mb-2">Need Help?</h4>
              <p className="text-sm text-secondary-600 mb-3">
                Have questions about this initiative or the application process?
              </p>
              <button className="w-full btn-secondary">
                Contact Admin
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
