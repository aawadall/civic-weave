import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'
import api from '../../services/api'
import { 
  getFieldRestrictions, 
  getStatusLabel, 
  getStatusColor,
  isFieldEditable 
} from '../../utils/projectLifecycle'
import DebugInfo from '../../components/DebugInfo'
import SkillChipInput from '../../components/SkillChipInput'
import RichTextEditor from '../../components/RichTextEditor'

// Utility function to extract plain text from TipTap JSON
const extractPlainText = (json) => {
  if (!json || !json.content) return ''
  
  const extractText = (node) => {
    if (node.type === 'text') {
      return node.text || ''
    }
    if (node.content) {
      return node.content.map(extractText).join('')
    }
    return ''
  }
  
  return json.content.map(extractText).join('\n').trim()
}

export default function EditProjectPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { user, hasAnyRole } = useAuth()
  const [project, setProject] = useState(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState(null)
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    content_json: null,
    location_address: '',
    start_date: '',
    end_date: '',
    team_lead_id: '',
    budget_total: '',
    budget_spent: ''
  })
  const [projectSkills, setProjectSkills] = useState([])
  const [skillsLoading, setSkillsLoading] = useState(false)

  useEffect(() => {
    if (id) {
      fetchProject()
      fetchProjectSkills()
    }
  }, [id])

  const fetchProject = async () => {
    try {
      setLoading(true)
      const response = await api.get(`/projects/${id}`)
      const projectData = response.data
      console.log('EditProjectPage: Project data:', projectData)
      setProject(projectData)
      
      // Populate form with current data
      setFormData({
        title: projectData.title || '',
        description: projectData.description || '',
        content_json: projectData.content_json || null,
        location_address: projectData.location_address || '',
        start_date: projectData.start_date ? new Date(projectData.start_date).toISOString().split('T')[0] : '',
        end_date: projectData.end_date ? new Date(projectData.end_date).toISOString().split('T')[0] : '',
        team_lead_id: projectData.team_lead_id || '',
        budget_total: projectData.budget_total || '',
        budget_spent: projectData.budget_spent || ''
      })
    } catch (err) {
      setError('Failed to fetch project details')
      console.error('Error fetching project:', err)
    } finally {
      setLoading(false)
    }
  }

  const fetchProjectSkills = async () => {
    try {
      setSkillsLoading(true)
      const response = await api.get(`/projects/${id}/skills`)
      const skills = response.data.skills || []
      setProjectSkills(skills.map(skill => skill.skill_name))
    } catch (err) {
      console.error('Error fetching project skills:', err)
      // If the endpoint doesn't exist yet or returns error, fall back to legacy field
      if (project?.required_skills) {
        setProjectSkills(project.required_skills)
      }
    } finally {
      setSkillsLoading(false)
    }
  }

  const handleInputChange = (e) => {
    const { name, value } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: value
    }))
  }

  const handleSkillsChange = (newSkills) => {
    setProjectSkills(newSkills)
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setSaving(true)
    setError(null)

    try {
      // Convert form data to API format
      const updateData = {
        ...formData,
        start_date: formData.start_date || null,
        end_date: formData.end_date || null,
        budget_total: formData.budget_total ? parseFloat(formData.budget_total) : null,
        budget_spent: formData.budget_spent ? parseFloat(formData.budget_spent) : null,
        team_lead_id: formData.team_lead_id || null
      }

      // Update project basic info
      await api.put(`/projects/${id}`, updateData)

      // Update project skills separately if they can be edited
      if (isFieldEditable('required_skills', currentStatus) || isAdmin) {
        try {
          await api.put(`/projects/${id}/skills`, {
            skill_names: projectSkills
          })
        } catch (skillsErr) {
          console.error('Error updating project skills:', skillsErr)
          // Don't fail the entire form submission for skills error
        }
      }

      navigate(`/projects/${id}`)
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to update project')
      console.error('Error updating project:', err)
    } finally {
      setSaving(false)
    }
  }

  const canEdit = () => {
    if (!project || !user) {
      console.log('EditProjectPage canEdit: No project or user', { project: !!project, user: !!user })
      return false
    }
    
    // Admin can always edit
    if (hasAnyRole('admin')) {
      console.log('EditProjectPage canEdit: User is admin')
      return true
    }
    
    // Team lead can edit if they're the team lead for this project
    if (hasAnyRole('team_lead') && project?.team_lead_id === user?.id) {
      console.log('EditProjectPage canEdit: User is team lead for this project')
      return true
    }
    
    // Project creator can edit
    if (project?.created_by_admin_id === user?.id) {
      console.log('EditProjectPage canEdit: User is project creator')
      return true
    }
    
    console.log('EditProjectPage canEdit: No permissions', { 
      userRoles: user?.roles, 
      teamLeadId: project?.team_lead_id, 
      createdBy: project?.created_by_admin_id,
      userId: user?.id 
    })
    return false
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  if (error && !project) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-secondary-900 mb-4">Project Not Found</h2>
          <p className="text-secondary-600 mb-4">{error}</p>
          <button 
            onClick={() => navigate('/projects')} 
            className="btn-primary"
          >
            Back to Projects
          </button>
        </div>
      </div>
    )
  }

  if (!canEdit()) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-secondary-900 mb-4">Access Denied</h2>
          <p className="text-secondary-600 mb-4">You don't have permission to edit this project.</p>
          <button 
            onClick={() => navigate(`/projects/${id}`)} 
            className="btn-primary"
          >
            Back to Project
          </button>
        </div>
      </div>
    )
  }

  const currentStatus = project?.project_status || project?.status
  const fieldRestrictions = getFieldRestrictions(currentStatus)
  const isAdmin = hasAnyRole('admin')

  return (
    <div className="min-h-screen bg-secondary-50 py-8">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Debug Info */}
        <DebugInfo project={project} title="Edit Project Page Debug" />
        
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between mb-4">
            <button 
              onClick={() => navigate(`/projects/${id}`)} 
              className="text-primary-600 hover:text-primary-800 text-sm font-medium"
            >
              ← Back to Project
            </button>
            <div className="flex items-center gap-3">
              <span className="text-sm text-gray-500">Current Status:</span>
              <span className={`px-3 py-1 text-sm font-medium rounded-full ${getStatusColor(currentStatus)}`}>
                {getStatusLabel(currentStatus)}
              </span>
            </div>
          </div>

          <h1 className="text-3xl font-bold text-secondary-900">Edit Project</h1>
          <p className="text-secondary-600 mt-2">
            Some fields may be restricted based on the current project status.
          </p>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="space-y-8">
          <div className="bg-white rounded-lg shadow-sm border border-secondary-200 p-6">
            <h2 className="text-xl font-bold text-secondary-900 mb-6">Project Details</h2>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Title */}
              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Project Title
                  {!isFieldEditable('title', currentStatus) && !isAdmin && (
                    <span className="text-xs text-gray-500 ml-2">(Locked in {getStatusLabel(currentStatus)} stage)</span>
                  )}
                </label>
                <input
                  type="text"
                  name="title"
                  value={formData.title}
                  onChange={handleInputChange}
                  disabled={!isFieldEditable('title', currentStatus) && !isAdmin}
                  className={`w-full px-3 py-2 border rounded-md ${
                    (!isFieldEditable('title', currentStatus) && !isAdmin) 
                      ? 'bg-gray-100 border-gray-300 text-gray-500' 
                      : 'border-gray-300 focus:ring-primary-500 focus:border-primary-500'
                  }`}
                  placeholder="Enter project title"
                />
                {!isFieldEditable('title', currentStatus) && !isAdmin && (
                  <p className="text-xs text-gray-500 mt-1">
                    Title cannot be changed once project moves beyond draft stage
                  </p>
                )}
              </div>

              {/* Description */}
              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Description
                  {!isFieldEditable('description', currentStatus) && !isAdmin && (
                    <span className="text-xs text-gray-500 ml-2">(Locked in {getStatusLabel(currentStatus)} stage)</span>
                  )}
                </label>
                <RichTextEditor 
                  value={formData.content_json} 
                  onChange={(json) => {
                    setFormData(prev => ({ ...prev, content_json: json }))
                    // Extract plain text for description field
                    if (json && json.content) {
                      const plainText = extractPlainText(json)
                      setFormData(prev => ({ ...prev, description: plainText }))
                    }
                  }}
                  readOnly={!isFieldEditable('description', currentStatus) && !isAdmin}
                  placeholder="Describe the project goals and requirements"
                />
              </div>

              {/* Required Skills */}
              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Required Skills
                  {!isFieldEditable('required_skills', currentStatus) && !isAdmin && (
                    <span className="text-xs text-gray-500 ml-2">(Locked in {getStatusLabel(currentStatus)} stage)</span>
                  )}
                </label>
                {skillsLoading ? (
                  <div className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-100">
                    <span className="text-gray-500">Loading skills...</span>
                  </div>
                ) : (
                  <SkillChipInput
                    selectedSkills={projectSkills}
                    onChange={handleSkillsChange}
                    disabled={!isFieldEditable('required_skills', currentStatus) && !isAdmin}
                    maxSkills={50}
                    placeholder="Add required skills..."
                  />
                )}
                {!isFieldEditable('required_skills', currentStatus) && !isAdmin && (
                  <p className="text-xs text-gray-500 mt-1">
                    Skills cannot be changed once project moves beyond draft stage
                  </p>
                )}
              </div>

              {/* Location */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Location Address
                </label>
                <input
                  type="text"
                  name="location_address"
                  value={formData.location_address}
                  onChange={handleInputChange}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                  placeholder="Enter project location"
                />
              </div>

              {/* Start Date */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Start Date
                  {!isFieldEditable('start_date', currentStatus) && !isAdmin && (
                    <span className="text-xs text-gray-500 ml-2">(Locked in {getStatusLabel(currentStatus)} stage)</span>
                  )}
                </label>
                <input
                  type="date"
                  name="start_date"
                  value={formData.start_date}
                  onChange={handleInputChange}
                  disabled={!isFieldEditable('start_date', currentStatus) && !isAdmin}
                  className={`w-full px-3 py-2 border rounded-md ${
                    (!isFieldEditable('start_date', currentStatus) && !isAdmin) 
                      ? 'bg-gray-100 border-gray-300 text-gray-500' 
                      : 'border-gray-300 focus:ring-primary-500 focus:border-primary-500'
                  }`}
                />
              </div>

              {/* End Date */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  End Date
                </label>
                <input
                  type="date"
                  name="end_date"
                  value={formData.end_date}
                  onChange={handleInputChange}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                />
              </div>

              {/* Budget Total */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Total Budget
                  {!isFieldEditable('budget_total', currentStatus) && !isAdmin && (
                    <span className="text-xs text-gray-500 ml-2">(Locked in {getStatusLabel(currentStatus)} stage)</span>
                  )}
                </label>
                <input
                  type="number"
                  name="budget_total"
                  value={formData.budget_total}
                  onChange={handleInputChange}
                  disabled={!isFieldEditable('budget_total', currentStatus) && !isAdmin}
                  className={`w-full px-3 py-2 border rounded-md ${
                    (!isFieldEditable('budget_total', currentStatus) && !isAdmin) 
                      ? 'bg-gray-100 border-gray-300 text-gray-500' 
                      : 'border-gray-300 focus:ring-primary-500 focus:border-primary-500'
                  }`}
                  placeholder="0.00"
                  step="0.01"
                  min="0"
                />
              </div>

              {/* Budget Spent */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Budget Spent
                </label>
                <input
                  type="number"
                  name="budget_spent"
                  value={formData.budget_spent}
                  onChange={handleInputChange}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-primary-500 focus:border-primary-500"
                  placeholder="0.00"
                  step="0.01"
                  min="0"
                />
              </div>
            </div>
          </div>

          {/* Error Display */}
          {error && (
            <div className="bg-red-50 border border-red-200 rounded-md p-4">
              <p className="text-sm text-red-600">{error}</p>
            </div>
          )}

          {/* Submit Button */}
          <div className="flex justify-end space-x-4">
            <button
              type="button"
              onClick={() => navigate(`/projects/${id}`)}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={saving}
              className="px-4 py-2 text-sm font-medium text-white bg-primary-600 border border-transparent rounded-md hover:bg-primary-700 disabled:opacity-50"
            >
              {saving ? 'Saving...' : 'Save Changes'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
