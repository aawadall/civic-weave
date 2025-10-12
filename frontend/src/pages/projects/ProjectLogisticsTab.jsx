import { useState, useEffect } from 'react'
import api from '../../services/api'

export default function ProjectLogisticsTab({ projectId }) {
  const [logistics, setLogistics] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchLogistics()
  }, [projectId])

  const fetchLogistics = async () => {
    try {
      setLoading(true)
      const response = await api.get(`/projects/${projectId}/logistics`)
      setLogistics(response.data)
    } catch (error) {
      console.error('Error fetching logistics:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleApproveApplication = async (applicationId, approve) => {
    try {
      await api.post(`/projects/${projectId}/approve-volunteer`, {
        application_id: applicationId,
        approve: approve
      })
      fetchLogistics()
    } catch (error) {
      console.error('Error processing application:', error)
      alert('Failed to process application')
    }
  }

  const handleRemoveVolunteer = async (volunteerId) => {
    if (!window.confirm('Are you sure you want to remove this volunteer from the project?')) {
      return
    }

    try {
      await api.delete(`/projects/${projectId}/volunteers/${volunteerId}`)
      fetchLogistics()
    } catch (error) {
      console.error('Error removing volunteer:', error)
      alert('Failed to remove volunteer')
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      {/* Budget Section - Under Construction */}
      <div className="bg-white rounded-lg border border-secondary-200 p-6">
        <h2 className="text-xl font-bold text-secondary-900 mb-4">Budget</h2>
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
          <p className="text-yellow-800 font-medium">🚧 Under Construction</p>
          <p className="text-sm text-yellow-700 mt-1">
            Detailed budgeting features are coming soon. For now, you can track basic budget information.
          </p>
        </div>
        
        {logistics?.project && (
          <div className="mt-4 space-y-2">
            <div className="flex justify-between">
              <span className="text-secondary-600">Total Budget:</span>
              <span className="font-medium">${logistics.project.budget_total || 0}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-secondary-600">Spent:</span>
              <span className="font-medium">${logistics.project.budget_spent || 0}</span>
            </div>
          </div>
        )}
      </div>

      {/* Volunteer Management */}
      <div className="bg-white rounded-lg border border-secondary-200 p-6">
        <h2 className="text-xl font-bold text-secondary-900 mb-4">Team Members</h2>
        
        {logistics?.team_members && logistics.team_members.length > 0 ? (
          <div className="space-y-3">
            {logistics.team_members.map(member => (
              <div key={member.id} className="flex items-center justify-between p-3 bg-secondary-50 rounded-lg">
                <div>
                  <p className="font-medium text-secondary-900">Volunteer ID: {member.volunteer_id}</p>
                  <p className="text-sm text-secondary-600">
                    Joined: {new Date(member.joined_at).toLocaleDateString()}
                  </p>
                  <span className={`text-xs px-2 py-1 rounded-full ${
                    member.status === 'active' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                  }`}>
                    {member.status}
                  </span>
                </div>
                {member.status === 'active' && (
                  <button
                    onClick={() => handleRemoveVolunteer(member.volunteer_id)}
                    className="px-4 py-2 text-red-600 hover:bg-red-50 rounded-lg"
                  >
                    Remove
                  </button>
                )}
              </div>
            ))}
          </div>
        ) : (
          <p className="text-secondary-600 text-center py-8">No team members yet</p>
        )}
      </div>

      {/* Pending Applications */}
      <div className="bg-white rounded-lg border border-secondary-200 p-6">
        <h2 className="text-xl font-bold text-secondary-900 mb-4">Pending Applications</h2>
        
        {logistics?.pending_applications && logistics.pending_applications.length > 0 ? (
          <div className="space-y-3">
            {logistics.pending_applications.map(app => (
              <div key={app.id} className="flex items-center justify-between p-4 bg-blue-50 rounded-lg">
                <div>
                  <p className="font-medium text-secondary-900">Volunteer ID: {app.volunteer_id}</p>
                  <p className="text-sm text-secondary-600">
                    Applied: {new Date(app.applied_at).toLocaleDateString()}
                  </p>
                  {app.admin_notes && (
                    <p className="text-sm text-secondary-700 mt-2 italic">"{app.admin_notes}"</p>
                  )}
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => handleApproveApplication(app.id, true)}
                    className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700"
                  >
                    Approve
                  </button>
                  <button
                    onClick={() => handleApproveApplication(app.id, false)}
                    className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
                  >
                    Reject
                  </button>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-secondary-600 text-center py-8">No pending applications</p>
        )}
      </div>

      {/* Campaigns Integration - Coming Soon */}
      <div className="bg-white rounded-lg border border-secondary-200 p-6">
        <h2 className="text-xl font-bold text-secondary-900 mb-4">Campaigns</h2>
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <p className="text-blue-800 font-medium">📧 Coming Soon</p>
          <p className="text-sm text-blue-700 mt-1">
            Integration with the campaigns module will allow you to send targeted communications to your team.
          </p>
        </div>
      </div>
    </div>
  )
}

