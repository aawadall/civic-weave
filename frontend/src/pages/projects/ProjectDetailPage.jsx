import { useState, useEffect } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'
import api from '../../services/api'
import RichTextEditor from '../../components/RichTextEditor'
import ProjectTasksTab from './ProjectTasksTab'
import ProjectMessagesTab from './ProjectMessagesTab'
import ProjectLogisticsTab from './ProjectLogisticsTab'
import ProjectStatusTransition from '../../components/ProjectStatusTransition'
import DebugInfo from '../../components/DebugInfo'

export default function ProjectDetailPage() {
  const { id, tab } = useParams()
  const navigate = useNavigate()
  const { user, hasAnyRole, hasRole } = useAuth()
  const [project, setProject] = useState(null)
  const [signups, setSignups] = useState([])
  const [teamMembers, setTeamMembers] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [activeTab, setActiveTab] = useState(tab || 'overview')
  const [applying, setApplying] = useState(false)
  const [isTeamMember, setIsTeamMember] = useState(false)
  const [volunteerProfile, setVolunteerProfile] = useState(null)
  const [availableTeamLeads, setAvailableTeamLeads] = useState([])
  const [showTeamLeadModal, setShowTeamLeadModal] = useState(false)

  useEffect(() => {
    if (id) {
      fetchProjectDetails()
      // Only check team membership if user is logged in
      if (user) {
        checkTeamMembership()
      }
    }
  }, [id, user])

  useEffect(() => {
    if (tab) {
      setActiveTab(tab)
    }
  }, [tab])

  const fetchProjectDetails = async () => {
    try {
      setLoading(true)
      const response = await api.get(`/projects/${id}`)
      console.log('Project data:', response.data)
      setProject(response.data)
      
      // Fetch additional data for team leads and admins
      if (hasAnyRole('team_lead', 'admin')) {
        try {
          const [signupsResponse, teamResponse] = await Promise.all([
            api.get(`/projects/${id}/signups`),
            api.get(`/projects/${id}/team-members`)
          ])
          setSignups(signupsResponse.data.signups || [])
          setTeamMembers(teamResponse.data.team_members || [])
        } catch (err) {
          console.warn('Could not fetch team data:', err)
        }
      }
    } catch (err) {
      setError('Failed to fetch project details')
      console.error('Error fetching project:', err)
    } finally {
      setLoading(false)
    }
  }

  const checkTeamMembership = async () => {
    try {
      // Check if user is in project team directly
      const teamResponse = await api.get(`/projects/${id}/team-members`)
      const members = teamResponse.data.team_members || []
      
      // Check if current user is a team member
      const isMember = members.some(m => m.user_id === user?.id && m.status === 'active')
      setIsTeamMember(isMember)
      
      // Set volunteer profile if found in team
      const myMember = members.find(m => m.user_id === user?.id)
      if (myMember) {
        setVolunteerProfile(myMember)
      }
    } catch (err) {
      console.warn('Could not check team membership:', err)
      // Don't set any volunteer profile if there's an error
      setVolunteerProfile(null)
      setIsTeamMember(false)
    }
  }

  const handleApply = async () => {
    try {
      setApplying(true)
      await api.post(`/projects/${id}/apply`)
      // Refresh project data to show updated application status
      fetchProjectDetails()
    } catch (err) {
      console.error('Error applying to project:', err)
      alert('Failed to apply to project. Please try again.')
    } finally {
      setApplying(false)
    }
  }

  const canManageProject = () => {
    if (!project || !user) {
      console.log('canManageProject: No project or user', { project: !!project, user: !!user })
      return false
    }
    
    // Admin can always manage
    if (hasAnyRole('admin')) {
      console.log('canManageProject: User is admin')
      return true
    }
    
    // Team lead can manage if they're the team lead for this project
    if (hasAnyRole('team_lead') && project?.team_lead_id === user?.id) {
      console.log('canManageProject: User is team lead for this project')
      return true
    }
    
    // Project creator can manage
    if (project?.created_by_admin_id === user?.id) {
      console.log('canManageProject: User is project creator')
      return true
    }
    
    console.log('canManageProject: No permissions', { 
      userRoles: user?.roles, 
      teamLeadId: project?.team_lead_id, 
      createdBy: project?.created_by_admin_id,
      userId: user?.id 
    })
    return false
  }

  const canApply = () => {
    return hasRole('volunteer') && project?.status === 'recruiting'
  }

  const handleStatusChange = (updatedProject) => {
    if (updatedProject) {
      setProject(updatedProject)
    }
  }

  const fetchAvailableTeamLeads = async () => {
    try {
      // Get users with team_lead role
      const response = await api.get('/admin/users')
      const users = response.data.users || []
      
      // Filter users with team_lead role
      const teamLeads = []
      for (const user of users) {
        try {
          const rolesResponse = await api.get(`/admin/users/${user.id}/roles`)
          const roles = rolesResponse.data.roles || []
          const hasTeamLeadRole = roles.some(role => role.name === 'team_lead')
          
          if (hasTeamLeadRole) {
            teamLeads.push({
              id: user.id,
              name: user.name || user.email,
              email: user.email
            })
          }
        } catch (err) {
          console.warn(`Failed to get roles for user ${user.id}:`, err)
        }
      }
      
      setAvailableTeamLeads(teamLeads)
      setShowTeamLeadModal(true)
    } catch (err) {
      console.error('Error fetching team leads:', err)
      alert('Failed to fetch available team leads. Please try again.')
    }
  }

  const handleAssignTeamLead = async (teamLeadId) => {
    try {
      await api.put(`/projects/${id}/team-lead`, {
        team_lead_id: teamLeadId
      })
      
      // Refresh project data
      await fetchProjectDetails()
      
      setShowTeamLeadModal(false)
      alert('Team lead assigned successfully! You can now transition to recruiting stage.')
    } catch (err) {
      console.error('Error assigning team lead:', err)
      alert('Failed to assign team lead. Please try again.')
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  if (error || !project) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-secondary-900 mb-4">Project Not Found</h2>
          <p className="text-secondary-600 mb-4">{error || 'The requested project could not be found.'}</p>
          <Link to="/projects" className="btn-primary">
            Back to Projects
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-secondary-50 py-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Debug Info */}
        <DebugInfo project={project} title="Project Detail Page Debug" />
        
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between mb-4">
            <Link to="/projects" className="text-primary-600 hover:text-primary-800 text-sm font-medium">
              ← Back to Projects
            </Link>
            {canManageProject() && (
              <div className="flex gap-2">
                <Link to={`/projects/${id}/edit`} className="btn-secondary">
                  Edit Project
                </Link>
                {hasAnyRole('team_lead', 'admin') && (
                  <Link to={`/projects/${id}/logistics`} className="btn-primary">
                    Manage Logistics
                  </Link>
                )}
              </div>
            )}
          </div>

          <div className="bg-white rounded-lg shadow-sm border border-secondary-200 p-6">
            <div className="flex justify-between items-start mb-4">
              <h1 className="text-3xl font-bold text-secondary-900">{project.title}</h1>
              <div className="flex items-center gap-3">
                {canManageProject() && project ? (
                  <ProjectStatusTransition 
                    project={project} 
                    onStatusChange={handleStatusChange}
                    compact={true}
                  />
                ) : (
                  <span className={`px-3 py-1 text-sm font-medium rounded-full ${
                    (project.status || project.project_status) === 'active' ? 'bg-green-100 text-green-800' :
                    (project.status || project.project_status) === 'recruiting' ? 'bg-blue-100 text-blue-800' :
                    (project.status || project.project_status) === 'completed' ? 'bg-gray-100 text-gray-800' :
                    'bg-yellow-100 text-yellow-800'
                  }`}>
                    {project.status || project.project_status || 'draft'}
                  </span>
                )}
                
                {/* Quick Assign Team Lead Button for Admins */}
                {hasAnyRole('admin') && !project.team_lead_id && (
                  <button
                    onClick={fetchAvailableTeamLeads}
                    className="px-3 py-1 text-sm bg-orange-100 text-orange-800 rounded-full hover:bg-orange-200 transition-colors"
                    title="Assign a team lead to enable recruiting stage"
                  >
                    Assign Team Lead
                  </button>
                )}
              </div>
            </div>

            <p className="text-secondary-600 text-lg mb-6">{project.description}</p>

            {/* Project Info */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
              {project.location_address && (
                <div>
                  <h3 className="text-sm font-medium text-secondary-900 mb-2">Location</h3>
                  <p className="text-secondary-600">📍 {project.location_address}</p>
                </div>
              )}
              {project.start_date && (
                <div>
                  <h3 className="text-sm font-medium text-secondary-900 mb-2">Start Date</h3>
                  <p className="text-secondary-600">{new Date(project.start_date).toLocaleDateString()}</p>
                </div>
              )}
              {project.end_date && (
                <div>
                  <h3 className="text-sm font-medium text-secondary-900 mb-2">End Date</h3>
                  <p className="text-secondary-600">{new Date(project.end_date).toLocaleDateString()}</p>
                </div>
              )}
            </div>

            {/* Action Buttons */}
            <div className="flex gap-4">
              {canApply() && (
                <button
                  onClick={handleApply}
                  disabled={applying}
                  className="btn-primary"
                >
                  {applying ? 'Applying...' : 'Apply to Project'}
                </button>
              )}
              {project.status === 'recruiting' && hasRole('volunteer') && (
                <p className="text-sm text-secondary-600 self-center">
                  This project is currently recruiting volunteers
                </p>
              )}
            </div>
          </div>
        </div>

        {/* Tabs */}
        <div className="bg-white rounded-lg shadow-sm border border-secondary-200">
          <div className="border-b border-secondary-200">
            <nav className="flex space-x-8 px-6 overflow-x-auto">
              <button
                onClick={() => setActiveTab('overview')}
                className={`py-4 px-1 border-b-2 font-medium text-sm whitespace-nowrap ${
                  activeTab === 'overview'
                    ? 'border-primary-500 text-primary-600'
                    : 'border-transparent text-secondary-500 hover:text-secondary-700'
                }`}
              >
                Overview
              </button>
              {project.required_skills && project.required_skills.length > 0 && (
                <button
                  onClick={() => setActiveTab('skills')}
                  className={`py-4 px-1 border-b-2 font-medium text-sm whitespace-nowrap ${
                    activeTab === 'skills'
                      ? 'border-primary-500 text-primary-600'
                      : 'border-transparent text-secondary-500 hover:text-secondary-700'
                  }`}
                >
                  Required Skills
                </button>
              )}
              {isTeamMember && (
                <>
                  <button
                    onClick={() => setActiveTab('tasks')}
                    className={`py-4 px-1 border-b-2 font-medium text-sm whitespace-nowrap ${
                      activeTab === 'tasks'
                        ? 'border-primary-500 text-primary-600'
                        : 'border-transparent text-secondary-500 hover:text-secondary-700'
                    }`}
                  >
                    Tasks
                  </button>
                  <button
                    onClick={() => setActiveTab('messages')}
                    className={`py-4 px-1 border-b-2 font-medium text-sm whitespace-nowrap ${
                      activeTab === 'messages'
                        ? 'border-primary-500 text-primary-600'
                        : 'border-transparent text-secondary-500 hover:text-secondary-700'
                    }`}
                  >
                    Messages
                  </button>
                </>
              )}
              {canManageProject() && (
                <>
                  <button
                    onClick={() => setActiveTab('logistics')}
                    className={`py-4 px-1 border-b-2 font-medium text-sm whitespace-nowrap ${
                      activeTab === 'logistics'
                        ? 'border-primary-500 text-primary-600'
                        : 'border-transparent text-secondary-500 hover:text-secondary-700'
                    }`}
                  >
                    Logistics
                  </button>
                  <button
                    onClick={() => setActiveTab('signups')}
                    className={`py-4 px-1 border-b-2 font-medium text-sm whitespace-nowrap ${
                      activeTab === 'signups'
                        ? 'border-primary-500 text-primary-600'
                        : 'border-transparent text-secondary-500 hover:text-secondary-700'
                    }`}
                  >
                    Signups ({signups.length})
                  </button>
                  <button
                    onClick={() => setActiveTab('team')}
                    className={`py-4 px-1 border-b-2 font-medium text-sm whitespace-nowrap ${
                      activeTab === 'team'
                        ? 'border-primary-500 text-primary-600'
                        : 'border-transparent text-secondary-500 hover:text-secondary-700'
                    }`}
                  >
                    Team ({teamMembers.length})
                  </button>
                </>
              )}
            </nav>
          </div>

          <div className="p-6">
            {activeTab === 'overview' && (
              <div>
                <h3 className="text-lg font-semibold text-secondary-900 mb-4">Project Description</h3>
                <div className="prose max-w-none">
                  {project.content_json ? (
                    <RichTextEditor value={project.content_json} readOnly={true} />
                  ) : (
                    <p className="text-secondary-600 whitespace-pre-wrap">{project.description}</p>
                  )}
                </div>
              </div>
            )}

            {activeTab === 'skills' && project.required_skills && (
              <div>
                <h3 className="text-lg font-semibold text-secondary-900 mb-4">Required Skills</h3>
                <div className="flex flex-wrap gap-2">
                  {project.required_skills.map((skill, index) => (
                    <span key={index} className="px-3 py-1 bg-primary-100 text-primary-800 text-sm rounded-full">
                      {skill}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {activeTab === 'tasks' && isTeamMember && (
              <ProjectTasksTab projectId={id} isProjectOwner={canManageProject()} />
            )}

            {activeTab === 'messages' && isTeamMember && (
              <ProjectMessagesTab projectId={id} />
            )}

            {activeTab === 'logistics' && canManageProject() && (
              <ProjectLogisticsTab projectId={id} />
            )}

            {activeTab === 'signups' && canManageProject() && (
              <div>
                <h3 className="text-lg font-semibold text-secondary-900 mb-4">Project Signups</h3>
                {signups.length > 0 ? (
                  <div className="space-y-4">
                    {signups.map((signup) => (
                      <div key={signup.id} className="border border-secondary-200 rounded-lg p-4">
                        <div className="flex justify-between items-start">
                          <div>
                            <h4 className="font-medium text-secondary-900">{signup.volunteer_name}</h4>
                            <p className="text-sm text-secondary-600">{signup.volunteer_email}</p>
                            <p className="text-sm text-secondary-500">Applied: {new Date(signup.applied_at).toLocaleDateString()}</p>
                          </div>
                          <span className={`px-2 py-1 text-xs font-medium rounded-full ${
                            signup.status === 'accepted' ? 'bg-green-100 text-green-800' :
                            signup.status === 'rejected' ? 'bg-red-100 text-red-800' :
                            'bg-yellow-100 text-yellow-800'
                          }`}>
                            {signup.status}
                          </span>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-secondary-600">No signups yet.</p>
                )}
              </div>
            )}

            {activeTab === 'team' && canManageProject() && (
              <div>
                <h3 className="text-lg font-semibold text-secondary-900 mb-4">Team Members</h3>
                {teamMembers.length > 0 ? (
                  <div className="space-y-4">
                    {teamMembers.map((member) => (
                      <div key={member.volunteer_id} className="border border-secondary-200 rounded-lg p-4">
                        <div className="flex justify-between items-start">
                          <div>
                            <h4 className="font-medium text-secondary-900">{member.name}</h4>
                            <p className="text-sm text-secondary-600">{member.email}</p>
                            <p className="text-sm text-secondary-500">Joined: {new Date(member.joined_at).toLocaleDateString()}</p>
                          </div>
                          <span className={`px-2 py-1 text-xs font-medium rounded-full ${
                            member.status === 'active' ? 'bg-green-100 text-green-800' :
                            member.status === 'completed' ? 'bg-blue-100 text-blue-800' :
                            'bg-yellow-100 text-yellow-800'
                          }`}>
                            {member.status}
                          </span>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-secondary-600">No team members yet.</p>
                )}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Team Lead Selection Modal */}
      {showTeamLeadModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-secondary-900 mb-4">Assign Team Lead</h3>
            <p className="text-secondary-600 mb-4">Select a team lead for this project:</p>
            
            {availableTeamLeads.length > 0 ? (
              <div className="space-y-2 mb-4">
                {availableTeamLeads.map((teamLead) => (
                  <button
                    key={teamLead.id}
                    onClick={() => handleAssignTeamLead(teamLead.id)}
                    className="w-full text-left p-3 border border-secondary-200 rounded-lg hover:bg-secondary-50 transition-colors"
                  >
                    <div className="font-medium text-secondary-900">{teamLead.name}</div>
                    <div className="text-sm text-secondary-600">{teamLead.email}</div>
                  </button>
                ))}
              </div>
            ) : (
              <div className="text-secondary-600 mb-4">
                No team leads available. Users need to have the 'team_lead' role to be assigned as team leads.
              </div>
            )}
            
            <div className="flex justify-end gap-3">
              <button
                onClick={() => setShowTeamLeadModal(false)}
                className="px-4 py-2 text-secondary-600 hover:text-secondary-800 transition-colors"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
