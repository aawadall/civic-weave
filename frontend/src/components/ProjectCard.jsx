import React from 'react'
import { Link } from 'react-router-dom'
import { CalendarIcon, MapPinIcon, UsersIcon, ChatBubbleLeftIcon, CheckCircleIcon } from '@heroicons/react/24/outline'

const ProjectCard = ({ project, showStats = true }) => {
  const getStatusColor = (status) => {
    switch (status) {
      case 'recruiting':
        return 'bg-blue-100 text-blue-800'
      case 'active':
        return 'bg-green-100 text-green-800'
      case 'completed':
        return 'bg-gray-100 text-gray-800'
      case 'cancelled':
        return 'bg-red-100 text-red-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const getPriorityColor = (priority) => {
    switch (priority) {
      case 'urgent':
        return 'bg-red-100 text-red-800'
      case 'high':
        return 'bg-orange-100 text-orange-800'
      case 'normal':
        return 'bg-blue-100 text-blue-800'
      case 'low':
        return 'bg-gray-100 text-gray-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const formatDate = (dateString) => {
    if (!dateString) return 'TBD'
    return new Date(dateString).toLocaleDateString()
  }

  return (
    <div className="bg-white rounded-lg shadow-md border border-gray-200 hover:shadow-lg transition-shadow duration-200">
      <div className="p-6">
        {/* Header */}
        <div className="flex items-start justify-between mb-4">
          <div className="flex-1">
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
              <Link 
                to={`/projects/${project.id}`}
                className="hover:text-blue-600 transition-colors"
              >
                {project.title}
              </Link>
            </h3>
            <p className="text-gray-600 text-sm line-clamp-2">
              {project.description}
            </p>
          </div>
          <div className="flex flex-col items-end space-y-2">
            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(project.project_status)}`}>
              {project.project_status}
            </span>
            {project.team_lead && (
              <span className="text-xs text-gray-500">
                Team Lead: {project.team_lead.name}
              </span>
            )}
          </div>
        </div>

        {/* Project Details */}
        <div className="space-y-3 mb-4">
          {project.location_address && (
            <div className="flex items-center text-sm text-gray-600">
              <MapPinIcon className="h-4 w-4 mr-2 flex-shrink-0" />
              <span className="truncate">{project.location_address}</span>
            </div>
          )}
          
          <div className="flex items-center text-sm text-gray-600">
            <CalendarIcon className="h-4 w-4 mr-2 flex-shrink-0" />
            <span>
              {formatDate(project.start_date)} - {formatDate(project.end_date)}
            </span>
          </div>

          {project.required_skills && project.required_skills.length > 0 && (
            <div className="flex flex-wrap gap-1">
              {project.required_skills.slice(0, 3).map((skill, index) => (
                <span 
                  key={index}
                  className="inline-flex items-center px-2 py-1 rounded-md text-xs font-medium bg-blue-50 text-blue-700"
                >
                  {skill}
                </span>
              ))}
              {project.required_skills.length > 3 && (
                <span className="inline-flex items-center px-2 py-1 rounded-md text-xs font-medium bg-gray-50 text-gray-600">
                  +{project.required_skills.length - 3} more
                </span>
              )}
            </div>
          )}
        </div>

        {/* Stats */}
        {showStats && (
          <div className="grid grid-cols-2 gap-4 mb-4">
            <div className="flex items-center text-sm text-gray-600">
              <UsersIcon className="h-4 w-4 mr-2" />
              <span>{project.active_team_count} members</span>
            </div>
            <div className="flex items-center text-sm text-gray-600">
              <CheckCircleIcon className="h-4 w-4 mr-2" />
              <span>{project.signup_count} applications</span>
            </div>
          </div>
        )}

        {/* Message and Task Stats */}
        {(project.unread_message_count > 0 || project.assigned_tasks_count > 0 || project.overdue_tasks_count > 0) && (
          <div className="flex items-center justify-between pt-3 border-t border-gray-200">
            <div className="flex items-center space-x-4 text-sm">
              {project.unread_message_count > 0 && (
                <div className="flex items-center text-blue-600">
                  <ChatBubbleLeftIcon className="h-4 w-4 mr-1" />
                  <span className="font-medium">{project.unread_message_count}</span>
                  <span className="ml-1">unread</span>
                </div>
              )}
              {project.assigned_tasks_count > 0 && (
                <div className="flex items-center text-gray-600">
                  <CheckCircleIcon className="h-4 w-4 mr-1" />
                  <span className="font-medium">{project.assigned_tasks_count}</span>
                  <span className="ml-1">tasks</span>
                </div>
              )}
              {project.overdue_tasks_count > 0 && (
                <div className="flex items-center text-red-600">
                  <span className="font-medium">{project.overdue_tasks_count}</span>
                  <span className="ml-1">overdue</span>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Action Button */}
        <div className="mt-4">
          <Link
            to={`/projects/${project.id}`}
            className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors"
          >
            View Project
          </Link>
        </div>
      </div>
    </div>
  )
}

export default ProjectCard
