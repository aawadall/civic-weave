import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'
import api from '../../services/api'

export default function CampaignsListPage() {
  const { hasAnyRole, hasRole, user } = useAuth()
  const [campaigns, setCampaigns] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [statusFilter, setStatusFilter] = useState('')

  useEffect(() => {
    fetchCampaigns()
  }, [])

  const fetchCampaigns = async () => {
    try {
      setLoading(true)
      const params = statusFilter ? { status: statusFilter } : {}
      const response = await api.get('/campaigns', { params })
      setCampaigns(response.data.campaigns || [])
    } catch (err) {
      setError('Failed to fetch campaigns')
      console.error('Error fetching campaigns:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleSendCampaign = async (campaignId) => {
    if (!confirm('Are you sure you want to send this campaign? This action cannot be undone.')) {
      return
    }

    try {
      await api.post(`/campaigns/${campaignId}/send`)
      alert('Campaign sent successfully!')
      fetchCampaigns() // Refresh the list
    } catch (err) {
      console.error('Error sending campaign:', err)
      alert('Failed to send campaign. Please try again.')
    }
  }

  const canCreateCampaign = hasAnyRole('campaign_manager', 'admin')
  const canManageCampaign = (campaign) => {
    return hasRole('admin') || campaign.created_by_user_id === user?.id
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
            onClick={fetchCampaigns}
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
            <h1 className="text-3xl font-bold text-secondary-900">Campaigns</h1>
            {canCreateCampaign && (
              <Link to="/campaigns/create" className="btn-primary">
                Create Campaign
              </Link>
            )}
          </div>

          {/* Filter */}
          <div className="mb-6">
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            >
              <option value="">All Statuses</option>
              <option value="draft">Draft</option>
              <option value="scheduled">Scheduled</option>
              <option value="sending">Sending</option>
              <option value="sent">Sent</option>
              <option value="failed">Failed</option>
            </select>
          </div>
        </div>

        {/* Campaigns List */}
        <div className="space-y-6">
          {campaigns.map((campaign) => (
            <div key={campaign.id} className="bg-white rounded-lg shadow-sm border border-secondary-200 p-6">
              <div className="flex justify-between items-start mb-4">
                <div>
                  <h3 className="text-xl font-semibold text-secondary-900 mb-2">
                    {campaign.name}
                  </h3>
                  <p className="text-secondary-600 mb-2">{campaign.description}</p>
                  <div className="flex items-center gap-4 text-sm text-secondary-500">
                    <span>Subject: {campaign.email_subject}</span>
                    <span>Created: {new Date(campaign.created_at).toLocaleDateString()}</span>
                    {campaign.sent_at && (
                      <span>Sent: {new Date(campaign.sent_at).toLocaleDateString()}</span>
                    )}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <span className={`px-3 py-1 text-sm font-medium rounded-full ${
                    campaign.status === 'sent' ? 'bg-green-100 text-green-800' :
                    campaign.status === 'sending' ? 'bg-blue-100 text-blue-800' :
                    campaign.status === 'scheduled' ? 'bg-yellow-100 text-yellow-800' :
                    campaign.status === 'failed' ? 'bg-red-100 text-red-800' :
                    'bg-gray-100 text-gray-800'
                  }`}>
                    {campaign.status}
                  </span>
                </div>
              </div>

              {/* Target Roles */}
              {campaign.target_roles && campaign.target_roles.length > 0 && (
                <div className="mb-4">
                  <h4 className="text-sm font-medium text-secondary-900 mb-2">Target Roles:</h4>
                  <div className="flex flex-wrap gap-1">
                    {campaign.target_roles.map((role, index) => (
                      <span key={index} className="px-2 py-1 bg-primary-100 text-primary-800 text-xs rounded">
                        {role}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {/* Actions */}
              <div className="flex justify-between items-center">
                <div className="flex gap-2">
                  <Link 
                    to={`/campaigns/${campaign.id}`}
                    className="text-primary-600 hover:text-primary-800 text-sm font-medium"
                  >
                    View Details
                  </Link>
                  {canManageCampaign(campaign) && (
                    <>
                      <Link 
                        to={`/campaigns/${campaign.id}/edit`}
                        className="text-secondary-600 hover:text-secondary-800 text-sm font-medium"
                      >
                        Edit
                      </Link>
                      <Link 
                        to={`/campaigns/${campaign.id}/preview`}
                        className="text-secondary-600 hover:text-secondary-800 text-sm font-medium"
                      >
                        Preview
                      </Link>
                    </>
                  )}
                </div>
                <div className="flex gap-2">
                  {campaign.status === 'draft' && canManageCampaign(campaign) && (
                    <button
                      onClick={() => handleSendCampaign(campaign.id)}
                      className="px-4 py-2 bg-green-600 text-white text-sm rounded-lg hover:bg-green-700"
                    >
                      Send Now
                    </button>
                  )}
                  {campaign.status === 'sent' && (
                    <Link 
                      to={`/campaigns/${campaign.id}/stats`}
                      className="px-4 py-2 bg-blue-600 text-white text-sm rounded-lg hover:bg-blue-700"
                    >
                      View Stats
                    </Link>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>

        {campaigns.length === 0 && (
          <div className="text-center py-12">
            <h3 className="text-lg font-medium text-secondary-900 mb-2">No campaigns found</h3>
            <p className="text-secondary-600 mb-4">
              {statusFilter 
                ? 'No campaigns match the selected status filter.'
                : 'No campaigns have been created yet.'
              }
            </p>
            {canCreateCampaign && (
              <Link to="/campaigns/create" className="btn-primary">
                Create First Campaign
              </Link>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
