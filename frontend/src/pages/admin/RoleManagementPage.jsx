import { useState, useEffect } from 'react'
import { useAuth } from '../../contexts/AuthContext'
import api from '../../services/api'

export default function RoleManagementPage() {
  const { hasRole } = useAuth()
  const [roles, setRoles] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [newRole, setNewRole] = useState({
    name: '',
    description: ''
  })
  const [creating, setCreating] = useState(false)

  useEffect(() => {
    fetchRoles()
  }, [])

  const fetchRoles = async () => {
    try {
      setLoading(true)
      const response = await api.get('/admin/roles')
      setRoles(response.data.roles || [])
    } catch (err) {
      setError('Failed to fetch roles')
      console.error('Error fetching roles:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleCreateRole = async (e) => {
    e.preventDefault()
    setCreating(true)

    try {
      await api.post('/admin/roles', newRole)
      setNewRole({ name: '', description: '' })
      setShowCreateModal(false)
      fetchRoles()
    } catch (err) {
      console.error('Error creating role:', err)
      alert('Failed to create role. Please try again.')
    } finally {
      setCreating(false)
    }
  }

  const handleDeleteRole = async (roleId) => {
    if (!confirm('Are you sure you want to delete this role? This action cannot be undone.')) {
      return
    }

    try {
      await api.delete(`/admin/roles/${roleId}`)
      fetchRoles()
    } catch (err) {
      console.error('Error deleting role:', err)
      alert('Failed to delete role. Please try again.')
    }
  }

  if (!hasRole('admin')) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-secondary-900 mb-4">Access Denied</h2>
          <p className="text-secondary-600 mb-4">You don't have permission to manage roles.</p>
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
            onClick={fetchRoles}
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
            <h1 className="text-3xl font-bold text-secondary-900">Role Management</h1>
            <button
              onClick={() => setShowCreateModal(true)}
              className="btn-primary"
            >
              Create Role
            </button>
          </div>
          <p className="text-secondary-600">Manage system roles and their permissions.</p>
        </div>

        {/* Roles Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {roles.map((role) => (
            <div key={role.id} className="bg-white rounded-lg shadow-sm border border-secondary-200 p-6">
              <div className="flex justify-between items-start mb-4">
                <h3 className="text-lg font-semibold text-secondary-900">{role.name}</h3>
                <button
                  onClick={() => handleDeleteRole(role.id)}
                  className="text-red-600 hover:text-red-800 text-sm font-medium"
                >
                  Delete
                </button>
              </div>
              
              <p className="text-secondary-600 text-sm mb-4">{role.description}</p>
              
              <div className="text-xs text-secondary-500">
                Created: {new Date(role.created_at).toLocaleDateString()}
              </div>
            </div>
          ))}
        </div>

        {roles.length === 0 && (
          <div className="text-center py-12">
            <h3 className="text-lg font-medium text-secondary-900 mb-2">No roles found</h3>
            <p className="text-secondary-600 mb-4">No roles have been created yet.</p>
            <button
              onClick={() => setShowCreateModal(true)}
              className="btn-primary"
            >
              Create First Role
            </button>
          </div>
        )}

        {/* Create Role Modal */}
        {showCreateModal && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
            <div className="bg-white rounded-lg shadow-xl max-w-md w-full">
              <div className="p-6">
                <div className="flex justify-between items-center mb-4">
                  <h2 className="text-xl font-semibold text-secondary-900">Create New Role</h2>
                  <button
                    onClick={() => setShowCreateModal(false)}
                    className="text-secondary-400 hover:text-secondary-600"
                  >
                    ×
                  </button>
                </div>

                <form onSubmit={handleCreateRole} className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-secondary-900 mb-2">
                      Role Name *
                    </label>
                    <input
                      type="text"
                      value={newRole.name}
                      onChange={(e) => setNewRole(prev => ({ ...prev, name: e.target.value }))}
                      required
                      className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                      placeholder="Enter role name (e.g., project_manager)"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-secondary-900 mb-2">
                      Description *
                    </label>
                    <textarea
                      value={newRole.description}
                      onChange={(e) => setNewRole(prev => ({ ...prev, description: e.target.value }))}
                      required
                      rows={3}
                      className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                      placeholder="Describe the role and its responsibilities"
                    />
                  </div>

                  <div className="flex justify-end gap-3">
                    <button
                      type="button"
                      onClick={() => setShowCreateModal(false)}
                      className="px-4 py-2 border border-secondary-300 text-secondary-700 rounded-lg hover:bg-secondary-50"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      disabled={creating}
                      className="btn-primary"
                    >
                      {creating ? 'Creating...' : 'Create Role'}
                    </button>
                  </div>
                </form>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
