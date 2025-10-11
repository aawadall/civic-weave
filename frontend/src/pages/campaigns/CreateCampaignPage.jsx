import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'
import api from '../../services/api'

export default function CreateCampaignPage() {
  const navigate = useNavigate()
  const { hasAnyRole } = useAuth()
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    target_roles: [],
    email_subject: '',
    email_body: '',
    scheduled_at: ''
  })
  const [newRole, setNewRole] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [previewData, setPreviewData] = useState(null)
  const [showPreview, setShowPreview] = useState(false)

  // Check if user has permission to create campaigns
  if (!hasAnyRole('campaign_manager', 'admin')) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-secondary-900 mb-4">Access Denied</h2>
          <p className="text-secondary-600 mb-4">You don't have permission to create campaigns.</p>
          <button onClick={() => navigate('/campaigns')} className="btn-primary">
            Back to Campaigns
          </button>
        </div>
      </div>
    )
  }

  const handleInputChange = (e) => {
    const { name, value } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: value
    }))
  }

  const handleAddRole = () => {
    if (newRole.trim() && !formData.target_roles.includes(newRole.trim())) {
      setFormData(prev => ({
        ...prev,
        target_roles: [...prev.target_roles, newRole.trim()]
      }))
      setNewRole('')
    }
  }

  const handleRemoveRole = (roleToRemove) => {
    setFormData(prev => ({
      ...prev,
      target_roles: prev.target_roles.filter(role => role !== roleToRemove)
    }))
  }

  const handlePreview = async () => {
    try {
      const response = await api.get('/campaigns/preview', {
        params: {
          target_roles: formData.target_roles.join(',')
        }
      })
      setPreviewData(response.data)
      setShowPreview(true)
    } catch (err) {
      console.error('Error getting preview:', err)
      alert('Failed to get preview data')
    }
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError(null)

    try {
      const response = await api.post('/campaigns', formData)
      navigate(`/campaigns/${response.data.id}`)
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to create campaign')
      console.error('Error creating campaign:', err)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-secondary-50 py-8">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mb-8">
          <button 
            onClick={() => navigate('/campaigns')}
            className="text-primary-600 hover:text-primary-800 text-sm font-medium mb-4"
          >
            ← Back to Campaigns
          </button>
          <h1 className="text-3xl font-bold text-secondary-900">Create New Campaign</h1>
          <p className="text-secondary-600 mt-2">Create an email campaign to reach out to volunteers.</p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Form */}
          <div className="lg:col-span-2">
            <div className="bg-white rounded-lg shadow-sm border border-secondary-200 p-6">
              {error && (
                <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
                  <p className="text-red-800">{error}</p>
                </div>
              )}

              <form onSubmit={handleSubmit} className="space-y-6">
                {/* Campaign Name */}
                <div>
                  <label htmlFor="name" className="block text-sm font-medium text-secondary-900 mb-2">
                    Campaign Name *
                  </label>
                  <input
                    type="text"
                    id="name"
                    name="name"
                    value={formData.name}
                    onChange={handleInputChange}
                    required
                    className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    placeholder="Enter campaign name"
                  />
                </div>

                {/* Description */}
                <div>
                  <label htmlFor="description" className="block text-sm font-medium text-secondary-900 mb-2">
                    Description
                  </label>
                  <textarea
                    id="description"
                    name="description"
                    value={formData.description}
                    onChange={handleInputChange}
                    rows={3}
                    className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    placeholder="Brief description of this campaign"
                  />
                </div>

                {/* Target Roles */}
                <div>
                  <label className="block text-sm font-medium text-secondary-900 mb-2">
                    Target Roles *
                  </label>
                  <div className="flex gap-2 mb-3">
                    <input
                      type="text"
                      value={newRole}
                      onChange={(e) => setNewRole(e.target.value)}
                      onKeyPress={(e) => e.key === 'Enter' && (e.preventDefault(), handleAddRole())}
                      className="flex-1 px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                      placeholder="Add a target role (e.g., volunteer, team_lead)"
                    />
                    <button
                      type="button"
                      onClick={handleAddRole}
                      className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700"
                    >
                      Add
                    </button>
                  </div>
                  {formData.target_roles.length > 0 && (
                    <div className="flex flex-wrap gap-2">
                      {formData.target_roles.map((role, index) => (
                        <span
                          key={index}
                          className="inline-flex items-center px-3 py-1 bg-primary-100 text-primary-800 text-sm rounded-full"
                        >
                          {role}
                          <button
                            type="button"
                            onClick={() => handleRemoveRole(role)}
                            className="ml-2 text-primary-600 hover:text-primary-800"
                          >
                            ×
                          </button>
                        </span>
                      ))}
                    </div>
                  )}
                </div>

                {/* Email Subject */}
                <div>
                  <label htmlFor="email_subject" className="block text-sm font-medium text-secondary-900 mb-2">
                    Email Subject *
                  </label>
                  <input
                    type="text"
                    id="email_subject"
                    name="email_subject"
                    value={formData.email_subject}
                    onChange={handleInputChange}
                    required
                    className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    placeholder="Enter email subject line"
                  />
                </div>

                {/* Email Body */}
                <div>
                  <label htmlFor="email_body" className="block text-sm font-medium text-secondary-900 mb-2">
                    Email Body *
                  </label>
                  <textarea
                    id="email_body"
                    name="email_body"
                    value={formData.email_body}
                    onChange={handleInputChange}
                    required
                    rows={8}
                    className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    placeholder="Write your email content here..."
                  />
                </div>

                {/* Schedule */}
                <div>
                  <label htmlFor="scheduled_at" className="block text-sm font-medium text-secondary-900 mb-2">
                    Schedule (Optional)
                  </label>
                  <input
                    type="datetime-local"
                    id="scheduled_at"
                    name="scheduled_at"
                    value={formData.scheduled_at}
                    onChange={handleInputChange}
                    className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                  />
                  <p className="text-sm text-secondary-500 mt-1">
                    Leave empty to save as draft. Set a future date to schedule the campaign.
                  </p>
                </div>

                {/* Submit Buttons */}
                <div className="flex justify-end gap-4 pt-6 border-t border-secondary-200">
                  <button
                    type="button"
                    onClick={() => navigate('/campaigns')}
                    className="px-6 py-2 border border-secondary-300 text-secondary-700 rounded-lg hover:bg-secondary-50"
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    onClick={handlePreview}
                    className="px-6 py-2 border border-primary-300 text-primary-700 rounded-lg hover:bg-primary-50"
                  >
                    Preview
                  </button>
                  <button
                    type="submit"
                    disabled={loading}
                    className="btn-primary"
                  >
                    {loading ? 'Creating...' : 'Create Campaign'}
                  </button>
                </div>
              </form>
            </div>
          </div>

          {/* Preview Sidebar */}
          <div className="lg:col-span-1">
            <div className="bg-white rounded-lg shadow-sm border border-secondary-200 p-6 sticky top-8">
              <h3 className="text-lg font-semibold text-secondary-900 mb-4">Campaign Preview</h3>
              
              {formData.name && (
                <div className="mb-4">
                  <h4 className="text-sm font-medium text-secondary-900 mb-1">Name</h4>
                  <p className="text-secondary-600">{formData.name}</p>
                </div>
              )}

              {formData.email_subject && (
                <div className="mb-4">
                  <h4 className="text-sm font-medium text-secondary-900 mb-1">Subject</h4>
                  <p className="text-secondary-600">{formData.email_subject}</p>
                </div>
              )}

              {formData.target_roles.length > 0 && (
                <div className="mb-4">
                  <h4 className="text-sm font-medium text-secondary-900 mb-2">Target Roles</h4>
                  <div className="flex flex-wrap gap-1">
                    {formData.target_roles.map((role, index) => (
                      <span key={index} className="px-2 py-1 bg-primary-100 text-primary-800 text-xs rounded">
                        {role}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {formData.email_body && (
                <div className="mb-4">
                  <h4 className="text-sm font-medium text-secondary-900 mb-2">Email Preview</h4>
                  <div className="border border-secondary-200 rounded-lg p-3 bg-secondary-50 text-sm">
                    <div className="font-medium mb-2">Subject: {formData.email_subject}</div>
                    <div className="whitespace-pre-wrap text-secondary-600">
                      {formData.email_body}
                    </div>
                  </div>
                </div>
              )}

              {formData.scheduled_at && (
                <div className="mb-4">
                  <h4 className="text-sm font-medium text-secondary-900 mb-1">Scheduled For</h4>
                  <p className="text-secondary-600">
                    {new Date(formData.scheduled_at).toLocaleString()}
                  </p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
