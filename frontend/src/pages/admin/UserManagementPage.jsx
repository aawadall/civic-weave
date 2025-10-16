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
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [showVerificationModal, setShowVerificationModal] = useState(false)
  const [showPasswordModal, setShowPasswordModal] = useState(false)
  const [userRoles, setUserRoles] = useState([])
  const [newPassword, setNewPassword] = useState('')
  const [passwordError, setPasswordError] = useState('')

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

  const openDeleteModal = (user) => {
    setSelectedUser(user)
    setShowDeleteModal(true)
  }

  const openVerificationModal = (user) => {
    setSelectedUser(user)
    setShowVerificationModal(true)
  }

  const openPasswordModal = (user) => {
    setSelectedUser(user)
    setNewPassword('')
    setPasswordError('')
    setShowPasswordModal(true)
  }

  const handleDeleteUser = async () => {
    if (!selectedUser) return

    try {
      await api.delete(`/admin/users/${selectedUser.id}`)
      setShowDeleteModal(false)
      setSelectedUser(null)
      fetchUsers() // Refresh users list
      alert('User deleted successfully')
    } catch (err) {
      console.error('Error deleting user:', err)
      let errorMessage = 'Failed to delete user'
      
      if (err.response?.status === 403) {
        errorMessage = 'Access denied. You may not have permission to delete this user.'
      } else if (err.response?.status === 500) {
        errorMessage = 'Server error. Please try again later or contact support.'
      } else if (err.response?.data?.error) {
        errorMessage = err.response.data.error
      }
      
      alert(errorMessage)
    }
  }

  const handleToggleVerification = async (verified) => {
    if (!selectedUser) return

    try {
      await api.put(`/admin/users/${selectedUser.id}/verification`, {
        email_verified: verified
      })
      setShowVerificationModal(false)
      setSelectedUser(null)
      fetchUsers() // Refresh users list
      alert(`User email ${verified ? 'verified' : 'unverified'} successfully`)
    } catch (err) {
      console.error('Error updating verification status:', err)
      let errorMessage = 'Failed to update verification status'
      
      if (err.response?.status === 403) {
        errorMessage = 'Access denied. You may not have permission to modify this user.'
      } else if (err.response?.status === 500) {
        errorMessage = 'Server error. Please try again later or contact support.'
      } else if (err.response?.data?.error) {
        errorMessage = err.response.data.error
      }
      
      alert(errorMessage)
    }
  }

  const handleChangePassword = async () => {
    if (!selectedUser || !newPassword) return

    // Validate password
    if (newPassword.length < 8) {
      setPasswordError('Password must be at least 8 characters long')
      return
    }

    try {
      await api.put(`/admin/users/${selectedUser.id}/password`, {
        new_password: newPassword
      })
      setShowPasswordModal(false)
      setSelectedUser(null)
      setNewPassword('')
      setPasswordError('')
      alert('Password changed successfully')
    } catch (err) {
      console.error('Error changing password:', err)
      let errorMessage = 'Failed to change password'
      
      if (err.response?.status === 403) {
        errorMessage = 'Access denied. You may not have permission to modify this user.'
      } else if (err.response?.status === 500) {
        errorMessage = 'Server error. Please try again later or contact support.'
      } else if (err.response?.data?.error) {
        errorMessage = err.response.data.error
      }
      
      setPasswordError(errorMessage)
    }
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
                      <div className="flex flex-col space-y-1">
                        <button
                          onClick={() => openRoleModal(user)}
                          className="text-primary-600 hover:text-primary-900 text-left"
                        >
                          Manage Roles
                        </button>
                        <button
                          onClick={() => openVerificationModal(user)}
                          className="text-blue-600 hover:text-blue-900 text-left"
                        >
                          {user.email_verified ? 'Unverify Email' : 'Verify Email'}
                        </button>
                        <button
                          onClick={() => openPasswordModal(user)}
                          className="text-orange-600 hover:text-orange-900 text-left"
                        >
                          Change Password
                        </button>
                        <button
                          onClick={() => openDeleteModal(user)}
                          className="text-red-600 hover:text-red-900 text-left"
                        >
                          Delete User
                        </button>
                      </div>
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

        {/* Delete User Modal */}
        {showDeleteModal && selectedUser && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
            <div className="bg-white rounded-lg shadow-xl max-w-md w-full">
              <div className="p-6">
                <div className="flex justify-between items-center mb-4">
                  <h2 className="text-xl font-semibold text-secondary-900">
                    Delete User
                  </h2>
                  <button
                    onClick={() => setShowDeleteModal(false)}
                    className="text-secondary-400 hover:text-secondary-600"
                  >
                    ×
                  </button>
                </div>
                <p className="text-secondary-600 mb-6">
                  Are you sure you want to delete <strong>{selectedUser.name || selectedUser.email}</strong>? 
                  This action cannot be undone and will permanently remove the user and all associated data.
                </p>
                <div className="flex justify-end space-x-3">
                  <button
                    onClick={() => setShowDeleteModal(false)}
                    className="px-4 py-2 border border-secondary-300 text-secondary-700 rounded-lg hover:bg-secondary-50"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleDeleteUser}
                    className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
                  >
                    Delete User
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Verification Status Modal */}
        {showVerificationModal && selectedUser && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
            <div className="bg-white rounded-lg shadow-xl max-w-md w-full">
              <div className="p-6">
                <div className="flex justify-between items-center mb-4">
                  <h2 className="text-xl font-semibold text-secondary-900">
                    Change Email Verification Status
                  </h2>
                  <button
                    onClick={() => setShowVerificationModal(false)}
                    className="text-secondary-400 hover:text-secondary-600"
                  >
                    ×
                  </button>
                </div>
                <p className="text-secondary-600 mb-6">
                  Current status for <strong>{selectedUser.name || selectedUser.email}</strong>: 
                  <span className={`ml-2 px-2 py-1 rounded-full text-xs font-medium ${
                    selectedUser.email_verified 
                      ? 'bg-green-100 text-green-800' 
                      : 'bg-yellow-100 text-yellow-800'
                  }`}>
                    {selectedUser.email_verified ? 'Verified' : 'Unverified'}
                  </span>
                </p>
                <div className="flex justify-end space-x-3">
                  <button
                    onClick={() => setShowVerificationModal(false)}
                    className="px-4 py-2 border border-secondary-300 text-secondary-700 rounded-lg hover:bg-secondary-50"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={() => handleToggleVerification(!selectedUser.email_verified)}
                    className={`px-4 py-2 rounded-lg text-white ${
                      selectedUser.email_verified 
                        ? 'bg-yellow-600 hover:bg-yellow-700' 
                        : 'bg-green-600 hover:bg-green-700'
                    }`}
                  >
                    {selectedUser.email_verified ? 'Mark as Unverified' : 'Mark as Verified'}
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Change Password Modal */}
        {showPasswordModal && selectedUser && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
            <div className="bg-white rounded-lg shadow-xl max-w-md w-full">
              <div className="p-6">
                <div className="flex justify-between items-center mb-4">
                  <h2 className="text-xl font-semibold text-secondary-900">
                    Change Password
                  </h2>
                  <button
                    onClick={() => setShowPasswordModal(false)}
                    className="text-secondary-400 hover:text-secondary-600"
                  >
                    ×
                  </button>
                </div>
                <p className="text-secondary-600 mb-4">
                  Set a new password for <strong>{selectedUser.name || selectedUser.email}</strong>
                </p>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-secondary-700 mb-2">
                    New Password
                  </label>
                  <input
                    type="password"
                    value={newPassword}
                    onChange={(e) => {
                      setNewPassword(e.target.value)
                      setPasswordError('')
                    }}
                    className="w-full px-3 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    placeholder="Enter new password (min 8 characters)"
                  />
                  {passwordError && (
                    <p className="mt-1 text-sm text-red-600">{passwordError}</p>
                  )}
                </div>
                <div className="flex justify-end space-x-3">
                  <button
                    onClick={() => setShowPasswordModal(false)}
                    className="px-4 py-2 border border-secondary-300 text-secondary-700 rounded-lg hover:bg-secondary-50"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleChangePassword}
                    disabled={!newPassword || newPassword.length < 8}
                    className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Change Password
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
