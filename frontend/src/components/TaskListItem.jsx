import React from 'react'
import { Link } from 'react-router-dom'
import { 
  CheckCircleIcon, 
  ClockIcon, 
  ExclamationTriangleIcon,
  CalendarIcon,
  UserIcon
} from '@heroicons/react/24/outline'
import { startTask } from '../services/api'

const TaskListItem = ({ task, showProject = true, showActions = true, onTaskUpdated }) => {
  const handleStartTask = async () => {
    try {
      await startTask(task.id)
      onTaskUpdated?.()
    } catch (error) {
      console.error('Error starting task:', error)
      alert('Failed to start task')
    }
  }
  const getStatusColor = (status) => {
    switch (status) {
      case 'todo':
        return 'bg-gray-100 text-gray-800'
      case 'in_progress':
        return 'bg-blue-100 text-blue-800'
      case 'blocked':
        return 'bg-red-100 text-red-800'
      case 'done':
        return 'bg-green-100 text-green-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const getPriorityColor = (priority) => {
    switch (priority) {
      case 'high':
        return 'bg-red-100 text-red-800'
      case 'medium':
        return 'bg-yellow-100 text-yellow-800'
      case 'low':
        return 'bg-green-100 text-green-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const getStatusIcon = (status) => {
    switch (status) {
      case 'todo':
        return <ClockIcon className="h-4 w-4" />
      case 'in_progress':
        return <ClockIcon className="h-4 w-4" />
      case 'blocked':
        return <ExclamationTriangleIcon className="h-4 w-4" />
      case 'done':
        return <CheckCircleIcon className="h-4 w-4" />
      default:
        return <ClockIcon className="h-4 w-4" />
    }
  }

  const formatDate = (dateString) => {
    if (!dateString) return null
    return new Date(dateString).toLocaleDateString()
  }

  const isOverdue = (dueDate, status) => {
    if (!dueDate || status === 'done') return false
    return new Date(dueDate) < new Date()
  }

  const getDueDateColor = (dueDate, status) => {
    if (!dueDate) return 'text-gray-500'
    if (status === 'done') return 'text-green-600'
    if (isOverdue(dueDate, status)) return 'text-red-600'
    return 'text-gray-600'
  }

  return (
    <div className="bg-white border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          {/* Task Header */}
          <div className="flex items-center space-x-3 mb-2">
            <div className="flex items-center space-x-2">
              {getStatusIcon(task.status)}
              <h3 className="text-sm font-medium text-gray-900 truncate">
                {task.title}
              </h3>
            </div>
            <div className="flex items-center space-x-2">
              <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(task.status)}`}>
                {task.status.replace('_', ' ')}
              </span>
              <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getPriorityColor(task.priority)}`}>
                {task.priority}
              </span>
            </div>
          </div>

          {/* Task Description */}
          {task.description && (
            <p className="text-sm text-gray-600 mb-3 line-clamp-2">
              {task.description}
            </p>
          )}

          {/* Task Details */}
          <div className="flex items-center space-x-4 text-sm text-gray-500 mb-3">
            {task.due_date && (
              <div className={`flex items-center ${getDueDateColor(task.due_date, task.status)}`}>
                <CalendarIcon className="h-4 w-4 mr-1" />
                <span className={isOverdue(task.due_date, task.status) ? 'font-medium' : ''}>
                  Due {formatDate(task.due_date)}
                  {isOverdue(task.due_date, task.status) && ' (Overdue)'}
                </span>
              </div>
            )}
            
            {task.assignee_name && (
              <div className="flex items-center text-gray-600">
                <UserIcon className="h-4 w-4 mr-1" />
                <span>{task.assignee_name}</span>
              </div>
            )}
          </div>

          {/* Project Link */}
          {showProject && task.project_title && (
            <div className="text-sm text-gray-500 mb-3">
              <Link 
                to={`/projects/${task.project_id}`}
                className="text-blue-600 hover:text-blue-800 hover:underline"
              >
                {task.project_title}
              </Link>
            </div>
          )}

          {/* Labels */}
          {task.labels && task.labels.length > 0 && (
            <div className="flex flex-wrap gap-1 mb-3">
              {task.labels.map((label, index) => (
                <span 
                  key={index}
                  className="inline-flex items-center px-2 py-1 rounded-md text-xs font-medium bg-gray-50 text-gray-700"
                >
                  {label}
                </span>
              ))}
            </div>
          )}

          {/* Task Actions */}
          {showActions && (
            <div className="flex items-center space-x-2">
              <Link
                to={`/tasks/${task.id}`}
                className="inline-flex items-center px-3 py-1.5 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
              >
                View Details
              </Link>
              
              {task.status === 'todo' && (
                <button 
                  onClick={handleStartTask}
                  className="inline-flex items-center px-3 py-1.5 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                >
                  Start Task
                </button>
              )}
              
              {task.status === 'in_progress' && (
                <button className="inline-flex items-center px-3 py-1.5 border border-transparent text-sm font-medium rounded-md text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500">
                  Mark Done
                </button>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default TaskListItem
