import React, { useState } from 'react'
import { 
  ExclamationTriangleIcon, 
  InformationCircleIcon, 
  XMarkIcon,
  CheckIcon
} from '@heroicons/react/24/outline'
import { markBroadcastRead } from '../services/api'

const BroadcastCard = ({ broadcast, onMarkAsRead, showActions = true }) => {
  const [isRead, setIsRead] = useState(broadcast.is_read)
  const [isDismissed, setIsDismissed] = useState(false)

  const getPriorityColor = (priority) => {
    switch (priority) {
      case 'urgent':
        return 'bg-red-100 text-red-800 border-red-200'
      case 'high':
        return 'bg-orange-100 text-orange-800 border-orange-200'
      case 'normal':
        return 'bg-blue-100 text-blue-800 border-blue-200'
      case 'low':
        return 'bg-gray-100 text-gray-800 border-gray-200'
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200'
    }
  }

  const getPriorityIcon = (priority) => {
    switch (priority) {
      case 'urgent':
        return <ExclamationTriangleIcon className="h-5 w-5 text-red-600" />
      case 'high':
        return <ExclamationTriangleIcon className="h-5 w-5 text-orange-600" />
      default:
        return <InformationCircleIcon className="h-5 w-5 text-blue-600" />
    }
  }

  const getTargetAudienceColor = (audience) => {
    switch (audience) {
      case 'all_users':
        return 'bg-purple-100 text-purple-800'
      case 'volunteers_only':
        return 'bg-green-100 text-green-800'
      case 'admins_only':
        return 'bg-red-100 text-red-800'
      case 'team_leads_only':
        return 'bg-blue-100 text-blue-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const formatDate = (dateString) => {
    if (!dateString) return null
    return new Date(dateString).toLocaleDateString()
  }

  const formatDateTime = (dateString) => {
    if (!dateString) return null
    return new Date(dateString).toLocaleString()
  }

  const handleMarkAsRead = async () => {
    if (isRead) return

    try {
      await markBroadcastRead(broadcast.id)
      setIsRead(true)
      if (onMarkAsRead) {
        onMarkAsRead(broadcast.id)
      }
    } catch (error) {
      console.error('Failed to mark broadcast as read:', error)
    }
  }

  const handleDismiss = () => {
    setIsDismissed(true)
  }

  if (isDismissed) {
    return null
  }

  return (
    <div className={`rounded-lg border-2 p-4 mb-4 ${getPriorityColor(broadcast.priority)} ${!isRead ? 'ring-2 ring-blue-500' : ''}`}>
      {/* Header */}
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center space-x-2">
          {getPriorityIcon(broadcast.priority)}
          <h3 className="text-lg font-semibold">
            {broadcast.title}
          </h3>
          {!isRead && (
            <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
              New
            </span>
          )}
        </div>
        
        <div className="flex items-center space-x-2">
          <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getTargetAudienceColor(broadcast.target_audience)}`}>
            {broadcast.target_audience.replace('_', ' ')}
          </span>
          {showActions && (
            <button
              onClick={handleDismiss}
              className="text-gray-400 hover:text-gray-600 transition-colors"
            >
              <XMarkIcon className="h-5 w-5" />
            </button>
          )}
        </div>
      </div>

      {/* Content */}
      <div className="mb-4">
        <p className="text-gray-700 whitespace-pre-wrap">
          {broadcast.content}
        </p>
      </div>

      {/* Footer */}
      <div className="flex items-center justify-between text-sm text-gray-600">
        <div className="flex items-center space-x-4">
          <span>By {broadcast.author_name}</span>
          <span>•</span>
          <span>{formatDateTime(broadcast.created_at)}</span>
          {broadcast.expires_at && (
            <>
              <span>•</span>
              <span>Expires {formatDate(broadcast.expires_at)}</span>
            </>
          )}
        </div>
        
        {showActions && !isRead && (
          <button
            onClick={handleMarkAsRead}
            className="inline-flex items-center px-3 py-1.5 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors"
          >
            <CheckIcon className="h-4 w-4 mr-1" />
            Mark as Read
          </button>
        )}
        
        {isRead && (
          <div className="flex items-center text-green-600">
            <CheckIcon className="h-4 w-4 mr-1" />
            <span>Read</span>
          </div>
        )}
      </div>
    </div>
  )
}

export default BroadcastCard
