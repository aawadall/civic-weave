// Project lifecycle utility functions

export const PROJECT_STATUSES = {
  DRAFT: 'draft',
  RECRUITING: 'recruiting', 
  ACTIVE: 'active',
  COMPLETED: 'completed',
  ARCHIVED: 'archived'
}

export const STATUS_LABELS = {
  [PROJECT_STATUSES.DRAFT]: 'Draft',
  [PROJECT_STATUSES.RECRUITING]: 'Recruiting',
  [PROJECT_STATUSES.ACTIVE]: 'Active',
  [PROJECT_STATUSES.COMPLETED]: 'Completed',
  [PROJECT_STATUSES.ARCHIVED]: 'Archived'
}

export const STATUS_COLORS = {
  [PROJECT_STATUSES.DRAFT]: 'bg-yellow-100 text-yellow-800',
  [PROJECT_STATUSES.RECRUITING]: 'bg-blue-100 text-blue-800',
  [PROJECT_STATUSES.ACTIVE]: 'bg-green-100 text-green-800',
  [PROJECT_STATUSES.COMPLETED]: 'bg-gray-100 text-gray-800',
  [PROJECT_STATUSES.ARCHIVED]: 'bg-red-100 text-red-800'
}

/**
 * Get available transitions from current status
 * @param {string} currentStatus - Current project status
 * @param {Object} projectData - Project data for validation
 * @returns {Array} Array of valid next statuses
 */
export function getAvailableTransitions(currentStatus, projectData = {}) {
  console.log('getAvailableTransitions: Current status:', currentStatus, 'Project data:', projectData)
  
  const transitions = {
    [PROJECT_STATUSES.DRAFT]: [PROJECT_STATUSES.RECRUITING, PROJECT_STATUSES.ARCHIVED],
    [PROJECT_STATUSES.RECRUITING]: [PROJECT_STATUSES.ACTIVE, PROJECT_STATUSES.ARCHIVED],
    [PROJECT_STATUSES.ACTIVE]: [PROJECT_STATUSES.COMPLETED, PROJECT_STATUSES.ARCHIVED],
    [PROJECT_STATUSES.COMPLETED]: [PROJECT_STATUSES.ARCHIVED],
    [PROJECT_STATUSES.ARCHIVED]: [] // No transitions from archived
  }

  let available = transitions[currentStatus] || []
  console.log('getAvailableTransitions: Available transitions before filtering:', available)
  
  // Filter based on validation requirements
  const filtered = available.filter(status => {
    switch (status) {
      case PROJECT_STATUSES.RECRUITING:
        const hasTeamLead = projectData.team_lead_id != null
        console.log('getAvailableTransitions: Recruiting check - hasTeamLead:', hasTeamLead, 'team_lead_id:', projectData.team_lead_id)
        return hasTeamLead
      case PROJECT_STATUSES.ACTIVE:
        const hasActiveMembers = projectData.active_team_count > 0
        console.log('getAvailableTransitions: Active check - hasActiveMembers:', hasActiveMembers, 'active_team_count:', projectData.active_team_count)
        return hasActiveMembers
      default:
        return true
    }
  })
  
  console.log('getAvailableTransitions: Final filtered transitions:', filtered)
  return filtered
}

/**
 * Get field restrictions for a given status
 * @param {string} status - Project status
 * @returns {Object} Object with field restrictions
 */
export function getFieldRestrictions(status) {
  const restrictions = {
    [PROJECT_STATUSES.DRAFT]: {
      // All fields editable
      title: true,
      description: true,
      required_skills: true,
      location_address: true,
      start_date: true,
      end_date: true,
      team_lead_id: true,
      budget_total: true,
      budget_spent: true
    },
    [PROJECT_STATUSES.RECRUITING]: {
      // Restrict title, description, required_skills
      title: false,
      description: false,
      required_skills: false,
      location_address: true,
      start_date: true,
      end_date: true,
      team_lead_id: true,
      budget_total: true,
      budget_spent: true
    },
    [PROJECT_STATUSES.ACTIVE]: {
      // Restrict title, required_skills, start_date
      title: false,
      description: true,
      required_skills: false,
      location_address: true,
      start_date: false,
      end_date: true,
      team_lead_id: true,
      budget_total: true,
      budget_spent: true
    },
    [PROJECT_STATUSES.COMPLETED]: {
      // Only team_lead_id, end_date, budget_spent editable
      title: false,
      description: false,
      required_skills: false,
      location_address: false,
      start_date: false,
      end_date: true,
      team_lead_id: true,
      budget_total: false,
      budget_spent: true
    },
    [PROJECT_STATUSES.ARCHIVED]: {
      // Only team_lead_id, end_date, budget_spent editable
      title: false,
      description: false,
      required_skills: false,
      location_address: false,
      start_date: false,
      end_date: true,
      team_lead_id: true,
      budget_total: false,
      budget_spent: true
    }
  }

  return restrictions[status] || {}
}

/**
 * Get transition requirements for moving from one status to another
 * @param {string} fromStatus - Current status
 * @param {string} toStatus - Target status
 * @returns {Object} Requirements object
 */
export function getTransitionRequirements(fromStatus, toStatus) {
  const requirements = {
    [PROJECT_STATUSES.RECRUITING]: {
      team_lead_assigned: true,
      message: 'Team lead must be assigned to move to recruiting'
    },
    [PROJECT_STATUSES.ACTIVE]: {
      active_team_members: true,
      message: 'At least one active team member required to move to active'
    }
  }

  return requirements[toStatus] || {}
}

/**
 * Get status color classes for Tailwind
 * @param {string} status - Project status
 * @returns {string} Tailwind classes
 */
export function getStatusColor(status) {
  return STATUS_COLORS[status] || 'bg-gray-100 text-gray-800'
}

/**
 * Get status label
 * @param {string} status - Project status
 * @returns {string} Human-readable label
 */
export function getStatusLabel(status) {
  return STATUS_LABELS[status] || status
}

/**
 * Check if a field is editable for a given status
 * @param {string} fieldName - Field name
 * @param {string} status - Project status
 * @returns {boolean} Whether field is editable
 */
export function isFieldEditable(fieldName, status) {
  const restrictions = getFieldRestrictions(status)
  return restrictions[fieldName] !== false
}

/**
 * Get validation checklist for a project
 * @param {Object} projectData - Project data
 * @returns {Array} Array of validation items
 */
export function getValidationChecklist(projectData) {
  return [
    {
      key: 'team_lead_assigned',
      label: 'Team lead assigned',
      valid: projectData.team_lead_id != null,
      required: true
    },
    {
      key: 'active_team_members',
      label: 'Active team members',
      valid: projectData.active_team_count > 0,
      required: true
    }
  ]
}

/**
 * Get workflow stages for visual display
 * @returns {Array} Array of workflow stages
 */
export function getWorkflowStages() {
  return [
    { status: PROJECT_STATUSES.DRAFT, label: 'Draft', color: STATUS_COLORS[PROJECT_STATUSES.DRAFT] },
    { status: PROJECT_STATUSES.RECRUITING, label: 'Recruiting', color: STATUS_COLORS[PROJECT_STATUSES.RECRUITING] },
    { status: PROJECT_STATUSES.ACTIVE, label: 'Active', color: STATUS_COLORS[PROJECT_STATUSES.ACTIVE] },
    { status: PROJECT_STATUSES.COMPLETED, label: 'Completed', color: STATUS_COLORS[PROJECT_STATUSES.COMPLETED] },
    { status: PROJECT_STATUSES.ARCHIVED, label: 'Archived', color: STATUS_COLORS[PROJECT_STATUSES.ARCHIVED] }
  ]
}
