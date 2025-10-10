import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useToast } from '../../contexts/ToastContext'
import { 
  PlusIcon, 
  PencilIcon, 
  TrashIcon,
  EyeIcon 
} from '@heroicons/react/24/outline'
import api from '../../services/api'

export default function InitiativesListPage() {
  const [initiatives, setInitiatives] = useState([])
  const [loading, setLoading] = useState(true)
  const [statusFilter, setStatusFilter] = useState('')

  const { showToast } = useToast()

  useEffect(() => {
    fetchInitiatives()
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

  const handleDelete = async (id, title) => {
    if (!window.confirm(`Are you sure you want to delete "${title}"?`)) {
      return
    }

    try {
      await api.delete(`/initiatives/${id}`)
      showToast('Initiative deleted successfully', 'success')
      fetchInitiatives()
    } catch (error) {
      showToast('Failed to delete initiative', 'error')
    }
  }

  const getStatusBadge = (status) => {
    const styles = {
      draft: 'bg-gray-100 text-gray-800',
      active: 'bg-green-100 text-green-800',
      closed: 'bg-red-100 text-red-800'
    }
    return (
      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${styles[status] || styles.draft}`}>
        {status.charAt(0).toUpperCase() + status.slice(1)}
      </span>
    )
  }

  const formatDate = (dateString) => {
    if (!dateString) return 'Not set'
    return new Date(dateString).toLocaleDateString()
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
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold text-secondary-900">
              Manage Initiatives
            </h1>
            <p className="text-secondary-600 mt-2">
              Create, edit, and manage volunteer opportunities
            </p>
          </div>
          <Link
            to="/admin/initiatives/create"
            className="btn-primary inline-flex items-center"
          >
            <PlusIcon className="h-5 w-5 mr-2" />
            Create Initiative
          </Link>
        </div>
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
            <option value="draft">Draft</option>
            <option value="active">Active</option>
            <option value="closed">Closed</option>
          </select>
        </div>
      </div>

      {/* Initiatives List */}
      {initiatives.length === 0 ? (
        <div className="text-center py-12">
          <div className="mx-auto h-12 w-12 text-secondary-400">
            <PlusIcon className="h-12 w-12" />
          </div>
          <h3 className="mt-2 text-sm font-medium text-secondary-900">No initiatives</h3>
          <p className="mt-1 text-sm text-secondary-500">
            Get started by creating your first initiative.
          </p>
          <div className="mt-6">
            <Link
              to="/admin/initiatives/create"
              className="btn-primary inline-flex items-center"
            >
              <PlusIcon className="h-5 w-5 mr-2" />
              Create Initiative
            </Link>
          </div>
        </div>
      ) : (
        <div className="bg-white shadow overflow-hidden sm:rounded-md">
          <ul className="divide-y divide-secondary-200">
            {initiatives.map((initiative) => (
              <li key={initiative.id}>
                <div className="px-4 py-4 sm:px-6">
                  <div className="flex items-center justify-between">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center space-x-3">
                        <h3 className="text-lg font-medium text-secondary-900 truncate">
                          {initiative.title}
                        </h3>
                        {getStatusBadge(initiative.status)}
                      </div>
                      <p className="mt-1 text-sm text-secondary-600 line-clamp-2">
                        {initiative.description}
                      </p>
                      <div className="mt-2 flex items-center space-x-4 text-sm text-secondary-500">
                        <span>📍 {initiative.location_address || 'No location'}</span>
                        <span>📅 {formatDate(initiative.start_date)} - {formatDate(initiative.end_date)}</span>
                        <span>🎯 {initiative.required_skills?.length || 0} required skills</span>
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Link
                        to={`/admin/initiatives/${initiative.id}`}
                        className="inline-flex items-center p-2 text-secondary-400 hover:text-secondary-600"
                        title="View details"
                      >
                        <EyeIcon className="h-5 w-5" />
                      </Link>
                      <Link
                        to={`/admin/initiatives/${initiative.id}/edit`}
                        className="inline-flex items-center p-2 text-secondary-400 hover:text-secondary-600"
                        title="Edit initiative"
                      >
                        <PencilIcon className="h-5 w-5" />
                      </Link>
                      <button
                        onClick={() => handleDelete(initiative.id, initiative.title)}
                        className="inline-flex items-center p-2 text-red-400 hover:text-red-600"
                        title="Delete initiative"
                      >
                        <TrashIcon className="h-5 w-5" />
                      </button>
                    </div>
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
          <div className="text-2xl font-bold text-primary-600">
            {initiatives.filter(i => i.status === 'active').length}
          </div>
          <div className="text-sm text-secondary-500">Active</div>
        </div>
        <div className="card text-center">
          <div className="text-2xl font-bold text-gray-600">
            {initiatives.filter(i => i.status === 'draft').length}
          </div>
          <div className="text-sm text-secondary-500">Draft</div>
        </div>
        <div className="card text-center">
          <div className="text-2xl font-bold text-red-600">
            {initiatives.filter(i => i.status === 'closed').length}
          </div>
          <div className="text-sm text-secondary-500">Closed</div>
        </div>
        <div className="card text-center">
          <div className="text-2xl font-bold text-secondary-600">
            {initiatives.length}
          </div>
          <div className="text-sm text-secondary-500">Total</div>
        </div>
      </div>
    </div>
  )
}
