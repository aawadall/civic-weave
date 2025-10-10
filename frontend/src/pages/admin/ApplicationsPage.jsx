import { useState, useEffect } from 'react'
import { useToast } from '../../contexts/ToastContext'
import { 
  CheckCircleIcon, 
  XCircleIcon,
  ClockIcon,
  UserIcon,
  MapPinIcon,
  CalendarIcon
} from '@heroicons/react/24/outline'
import api from '../../services/api'

export default function ApplicationsPage() {
  const [applications, setApplications] = useState([])
  const [loading, setLoading] = useState(true)
  const [statusFilter, setStatusFilter] = useState('pending')
  const [processing, setProcessing] = useState(null)

  const { showToast } = useToast()

  useEffect(() => {
    fetchApplications()
  }, [statusFilter])

  const fetchApplications = async () => {
    try {
      setLoading(true)
      const params = new URLSearchParams()
      if (statusFilter) params.append('status', statusFilter)
      
      const response = await api.get(`/applications?${params.toString()}`)
      setApplications(response.data.applications || [])
    } catch (error) {
      showToast('Failed to load applications', 'error')
    } finally {
      setLoading(false)
    }
  }

  const handleStatusChange = async (applicationId, newStatus, notes = '') => {
    try {
      setProcessing(applicationId)
      await api.put(`/applications/${applicationId}`, {
        status: newStatus,
        admin_notes: notes
      })
      
      showToast(`Application ${newStatus} successfully`, 'success')
      fetchApplications()
    } catch (error) {
      showToast('Failed to update application status', 'error')
    } finally {
      setProcessing(null)
    }
  }

  const getStatusBadge = (status) => {
    const styles = {
      pending: 'bg-yellow-100 text-yellow-800',
      accepted: 'bg-green-100 text-green-800',
      rejected: 'bg-red-100 text-red-800'
    }
    return (
      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${styles[status] || styles.pending}`}>
        {status.charAt(0).toUpperCase() + status.slice(1)}
      </span>
    )
  }

  const getStatusIcon = (status) => {
    switch (status) {
      case 'accepted':
        return <CheckCircleIcon className="h-5 w-5 text-green-500" />
      case 'rejected':
        return <XCircleIcon className="h-5 w-5 text-red-500" />
      default:
        return <ClockIcon className="h-5 w-5 text-yellow-500" />
    }
  }

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
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
          Volunteer Applications
        </h1>
        <p className="text-secondary-600 mt-2">
          Review and manage volunteer applications for initiatives
        </p>
      </div>

      {/* Filters */}
      <div className="mb-6">
        <div className="flex items-center space-x-4">
          <label htmlFor="statusFilter" className="text-sm font-medium text-secondary-700">
            Filter by status:
          </label>
          <select
            id="statusFilter"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="input-field w-auto"
          >
            <option value="">All Statuses</option>
            <option value="pending">Pending</option>
            <option value="accepted">Accepted</option>
            <option value="rejected">Rejected</option>
          </select>
        </div>
      </div>

      {/* Applications List */}
      {applications.length === 0 ? (
        <div className="text-center py-12">
          <ClockIcon className="mx-auto h-12 w-12 text-secondary-400" />
          <h3 className="mt-2 text-sm font-medium text-secondary-900">No applications</h3>
          <p className="mt-1 text-sm text-secondary-500">
            {statusFilter === 'pending' 
              ? 'No pending applications at the moment.'
              : `No ${statusFilter} applications found.`
            }
          </p>
        </div>
      ) : (
        <div className="bg-white shadow overflow-hidden sm:rounded-md">
          <ul className="divide-y divide-secondary-200">
            {applications.map((application) => (
              <li key={application.id}>
                <div className="px-4 py-4 sm:px-6">
                  <div className="flex items-center justify-between">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center space-x-3 mb-2">
                        {getStatusIcon(application.status)}
                        <h3 className="text-lg font-medium text-secondary-900">
                          Application #{application.id.slice(0, 8)}
                        </h3>
                        {getStatusBadge(application.status)}
                      </div>
                      
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm text-secondary-600">
                        <div className="flex items-center space-x-2">
                          <UserIcon className="h-4 w-4" />
                          <span>Volunteer ID: {application.volunteer_id.slice(0, 8)}...</span>
                        </div>
                        <div className="flex items-center space-x-2">
                          <CalendarIcon className="h-4 w-4" />
                          <span>Applied: {formatDate(application.applied_at)}</span>
                        </div>
                      </div>

                      {application.admin_notes && (
                        <div className="mt-3 p-3 bg-secondary-50 rounded-lg">
                          <h4 className="text-sm font-medium text-secondary-900 mb-1">
                            Volunteer Message:
                          </h4>
                          <p className="text-sm text-secondary-600">
                            {application.admin_notes}
                          </p>
                        </div>
                      )}
                    </div>

                    {application.status === 'pending' && (
                      <div className="flex items-center space-x-2 ml-4">
                        <button
                          onClick={() => {
                            const notes = prompt('Add notes (optional):')
                            handleStatusChange(application.id, 'accepted', notes || '')
                          }}
                          disabled={processing === application.id}
                          className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 disabled:opacity-50"
                        >
                          {processing === application.id ? 'Processing...' : 'Accept'}
                        </button>
                        <button
                          onClick={() => {
                            const notes = prompt('Add rejection reason:')
                            if (notes) {
                              handleStatusChange(application.id, 'rejected', notes)
                            }
                          }}
                          disabled={processing === application.id}
                          className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50"
                        >
                          Reject
                        </button>
                      </div>
                    )}
                  </div>
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Stats */}
      <div className="mt-8 grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="card text-center">
          <div className="text-2xl font-bold text-yellow-600">
            {applications.filter(a => a.status === 'pending').length}
          </div>
          <div className="text-sm text-secondary-500">Pending Review</div>
        </div>
        <div className="card text-center">
          <div className="text-2xl font-bold text-green-600">
            {applications.filter(a => a.status === 'accepted').length}
          </div>
          <div className="text-sm text-secondary-500">Accepted</div>
        </div>
        <div className="card text-center">
          <div className="text-2xl font-bold text-red-600">
            {applications.filter(a => a.status === 'rejected').length}
          </div>
          <div className="text-sm text-secondary-500">Rejected</div>
        </div>
        <div className="card text-center">
          <div className="text-2xl font-bold text-secondary-600">
            {applications.length}
          </div>
          <div className="text-sm text-secondary-500">Total Applications</div>
        </div>
      </div>
    </div>
  )
}
