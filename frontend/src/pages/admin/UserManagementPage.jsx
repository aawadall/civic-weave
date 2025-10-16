import { useState, useEffect, useMemo } from 'react'
import { useAuth } from '../../contexts/AuthContext'
import api from '../../services/api'

export default function UserManagementPage() {
  const { hasRole } = useAuth()
  const [users, setUsers] = useState([])
  const [roles, setRoles] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedUser, setSelectedUser] = useState(null)
  const [showRoleModal, setShowRoleModal] = useState(false)
  const [userRoles, setUserRoles] = useState([])

  useEffect(() => {
    fetchUsers()
    fetchRoles()
  }, [])

  const fetchUsers = async () => {
    try {
      setLoading(true)
      const response = await api.get('/admin/users')
      setUsers(response.data.users || [])
    } catch (err) {
      setError('Failed to fetch users')
      console.error('Error fetching users:', err)
    } finally {
      setLoading(false)
    }
  }

  const fetchRoles = async () => {
    try {
      const response = await api.get('/admin/roles')
      setRoles(response.data.roles || [])
    } catch (err) {
      console.error('Error fetching roles:', err)
    }
  }

  const fetchUserRoles = async (userId) => {
    try {
      const response = await api.get(`/admin/users/${userId}/roles`)
      setUserRoles(response.data.roles || [])
    } catch (err) {
      console.error('Error fetching user roles:', err)
    }
  }

  const availableRoles = useMemo(() => {
    if (!selectedUser) {
      return []
    }

    const assignedRoleIds = new Set(userRoles.map(role => role.id))
    return roles.filter(role => !assignedRoleIds.has(role.id))
  }, [roles, userRoles, selectedUser])

  const handleAssignRole = async (userId, roleId) => {
    try {
      await api.post(`/admin/users/${userId}/roles`, { role_id: roleId })
      fetchUserRoles(userId)
      fetchUsers() // Refresh users list
    } catch (err) {
      console.error('Error assigning role:', err)
      alert('Failed to assign role. Please try again.')
    }
  }

  const handleRevokeRole = async (userId, roleId) => {
    try {
      await api.delete(`/admin/users/${userId}/roles/${roleId}`)
      fetchUserRoles(userId)
      fetchUsers() // Refresh users list
    } catch (err) {
      console.error('Error revoking role:', err)
      alert('Failed to revoke role. Please try again.')
    }
  }

  const openRoleModal = (user) => {
    setSelectedUser(user)
    setShowRoleModal(true)
    fetchUserRoles(user.id)
  }

  const filteredUsers = users.filter(user => 
    user.email?.toLowerCase().includes(searchTerm.toLowerCase()) ||
    user.name?.toLowerCase().includes(searchTerm.toLowerCase())
  )

  if (!hasRole('admin')) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-secondary-900 mb-4">Access Denied</h2>
          <p className="text-secondary-600 mb-4">You don't have permission to manage users.</p>
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
            onClick={fetchUsers}
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
          <h1 className="text-3xl font-bold text-secondary-900 mb-2">User Management</h1>
          <p className="text-secondary-600">Manage users and their role assignments.</p>
        </div>

        {/* Search */}
        <div className="mb-6">
          <input
            type="text"
            placeholder="Search users..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full max-w-md px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
          />
        </div>

        {/* Users Table */}
        <div className="bg-white rounded-lg shadow-sm border border-secondary-200 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-secondary-200">
              <thead className="bg-secondary-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-secondary-500 uppercase tracking-wider">
                    User
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-secondary-500 uppercase tracking-wider">
                    Roles
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-secondary-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-secondary-500 uppercase tracking-wider">
                    Joined
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-secondary-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-secondary-200">
                {filteredUsers.map((user) => (
                  <tr key={user.id}>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div>
                        <div className="text-sm font-medium text-secondary-900">
                          {user.name || 'Unnamed User'}
                        </div>
                        <div className="text-sm text-secondary-500">{user.email}</div>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex flex-wrap gap-1">
                        {user.roles && user.roles.length > 0 ? (
                          user.roles.map((role, index) => (
                            <span
                              key={index}
                              className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-primary-100 text-primary-800"
                            >
                              {role.name}
                            </span>
                          ))
                        ) : (
                          <span className="text-sm text-secondary-500">No roles assigned</span>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                        user.email_verified 
                          ? 'bg-green-100 text-green-800' 
                          : 'bg-yellow-100 text-yellow-800'
                      }`}>
                        {user.email_verified ? 'Verified' : 'Unverified'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-secondary-500">
                      {new Date(user.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <button
                        onClick={() => openRoleModal(user)}
                        className="text-primary-600 hover:text-primary-900"
                      >
                        Manage Roles
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {filteredUsers.length === 0 && (
          <div className="text-center py-12">
            <h3 className="text-lg font-medium text-secondary-900 mb-2">No users found</h3>
            <p className="text-secondary-600">
              {searchTerm ? 'Try adjusting your search criteria.' : 'No users are registered yet.'}
            </p>
          </div>
        )}

        {/* Role Management Modal */}
        {showRoleModal && selectedUser && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
            <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[80vh] overflow-y-auto">
              <div className="p-6">
                <div className="flex justify-between items-center mb-4">
                  <h2 className="text-xl font-semibold text-secondary-900">
                    Manage Roles for {selectedUser.name || selectedUser.email}
                  </h2>
                  <button
                    onClick={() => setShowRoleModal(false)}
                    className="text-secondary-400 hover:text-secondary-600"
                  >
                    ×
                  </button>
                </div>

                {/* Current Roles */}
                <div className="mb-6">
                  <h3 className="text-lg font-medium text-secondary-900 mb-3">Current Roles</h3>
                  {userRoles.length > 0 ? (
                    <div className="space-y-2">
                      {userRoles.map((role) => (
                        <div key={role.id} className="flex justify-between items-center p-3 border border-secondary-200 rounded-lg">
                          <div>
                            <div className="font-medium text-secondary-900">{role.name}</div>
                            <div className="text-sm text-secondary-600">{role.description}</div>
                          </div>
                          <button
                            onClick={() => handleRevokeRole(selectedUser.id, role.id)}
                            className="text-red-600 hover:text-red-800 text-sm font-medium"
                          >
                            Remove
                          </button>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className="text-secondary-600">No roles assigned</p>
                  )}
                </div>

                {/* Available Roles */}
                <div>
                  <h3 className="text-lg font-medium text-secondary-900 mb-3">Available Roles</h3>
                  {availableRoles.length > 0 ? (
                    <div className="space-y-2">
                      {availableRoles.map((role) => (
                        <div key={role.id} className="flex justify-between items-center p-3 border border-secondary-200 rounded-lg">
                          <div>
                            <div className="font-medium text-secondary-900">{role.name}</div>
                            <div className="text-sm text-secondary-600">{role.description}</div>
                          </div>
                          <button
                            onClick={() => handleAssignRole(selectedUser.id, role.id)}
                            className="text-primary-600 hover:text-primary-800 text-sm font-medium"
                          >
                            Assign
                          </button>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className="text-secondary-600">All roles have been assigned</p>
                  )}
                </div>

                <div className="flex justify-end mt-6">
                  <button
                    onClick={() => setShowRoleModal(false)}
                    className="px-4 py-2 border border-secondary-300 text-secondary-700 rounded-lg hover:bg-secondary-50"
                  >
                    Close
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
