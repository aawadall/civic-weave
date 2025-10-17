import React, { useState } from 'react'
import { 
  DocumentIcon, 
  LinkIcon, 
  ArrowDownTrayIcon,
  EyeIcon,
  CalendarIcon,
  UserIcon
} from '@heroicons/react/24/outline'
import { downloadResource } from '../services/api'

const ResourceCard = ({ resource, showActions = true, showUploader = true }) => {
  const [isDownloading, setIsDownloading] = useState(false)

  const getResourceTypeIcon = (type) => {
    switch (type) {
      case 'file':
        return <DocumentIcon className="h-5 w-5" />
      case 'link':
        return <LinkIcon className="h-5 w-5" />
      case 'document':
        return <DocumentIcon className="h-5 w-5" />
      default:
        return <DocumentIcon className="h-5 w-5" />
    }
  }

  const getResourceTypeColor = (type) => {
    switch (type) {
      case 'file':
        return 'bg-blue-100 text-blue-800'
      case 'link':
        return 'bg-green-100 text-green-800'
      case 'document':
        return 'bg-purple-100 text-purple-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const getScopeColor = (scope) => {
    switch (scope) {
      case 'global':
        return 'bg-purple-100 text-purple-800'
      case 'project_specific':
        return 'bg-blue-100 text-blue-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const formatFileSize = (bytes) => {
    if (!bytes) return null
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(1024))
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i]
  }

  const formatDate = (dateString) => {
    if (!dateString) return null
    return new Date(dateString).toLocaleDateString()
  }

  const handleDownload = async () => {
    if (isDownloading) return

    setIsDownloading(true)
    try {
      const response = await downloadResource(resource.id)
      
      // Create blob URL and trigger download
      const blob = new Blob([response.data])
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = resource.title
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
    } catch (error) {
      console.error('Failed to download resource:', error)
      // For links, open in new tab
      if (resource.resource_type === 'link') {
        window.open(resource.file_url, '_blank')
      }
    } finally {
      setIsDownloading(false)
    }
  }

  const handleView = () => {
    if (resource.resource_type === 'link') {
      window.open(resource.file_url, '_blank')
    } else {
      handleDownload()
    }
  }

  return (
    <div className="bg-white border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center space-x-3">
          <div className={`p-2 rounded-lg ${getResourceTypeColor(resource.resource_type)}`}>
            {getResourceTypeIcon(resource.resource_type)}
          </div>
          <div className="flex-1 min-w-0">
            <h3 className="text-lg font-medium text-gray-900 truncate">
              {resource.title}
            </h3>
            {resource.description && (
              <p className="text-sm text-gray-600 mt-1 line-clamp-2">
                {resource.description}
              </p>
            )}
          </div>
        </div>
        
        <div className="flex items-center space-x-2">
          <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getResourceTypeColor(resource.resource_type)}`}>
            {resource.resource_type}
          </span>
          <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getScopeColor(resource.scope)}`}>
            {resource.scope.replace('_', ' ')}
          </span>
        </div>
      </div>

      {/* Resource Details */}
      <div className="space-y-2 mb-4">
        {resource.file_size && (
          <div className="text-sm text-gray-600">
            Size: {formatFileSize(resource.file_size)}
          </div>
        )}
        
        {resource.mime_type && (
          <div className="text-sm text-gray-600">
            Type: {resource.mime_type}
          </div>
        )}

        {resource.project_title && (
          <div className="text-sm text-gray-600">
            Project: {resource.project_title}
          </div>
        )}

        {resource.tags && resource.tags.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {resource.tags.map((tag, index) => (
              <span 
                key={index}
                className="inline-flex items-center px-2 py-1 rounded-md text-xs font-medium bg-gray-50 text-gray-700"
              >
                {tag}
              </span>
            ))}
          </div>
        )}
      </div>

      {/* Stats and Metadata */}
      <div className="flex items-center justify-between text-sm text-gray-500 mb-4">
        <div className="flex items-center space-x-4">
          <div className="flex items-center">
            <ArrowDownTrayIcon className="h-4 w-4 mr-1" />
            <span>{resource.download_count} downloads</span>
          </div>
          
          <div className="flex items-center">
            <CalendarIcon className="h-4 w-4 mr-1" />
            <span>{formatDate(resource.created_at)}</span>
          </div>
        </div>
        
        {showUploader && resource.uploader_name && (
          <div className="flex items-center">
            <UserIcon className="h-4 w-4 mr-1" />
            <span>{resource.uploader_name}</span>
          </div>
        )}
      </div>

      {/* Actions */}
      {showActions && (
        <div className="flex items-center space-x-2">
          <button
            onClick={handleView}
            className="inline-flex items-center px-3 py-1.5 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            <EyeIcon className="h-4 w-4 mr-1" />
            {resource.resource_type === 'link' ? 'Open Link' : 'View'}
          </button>
          
          <button
            onClick={handleDownload}
            disabled={isDownloading}
            className="inline-flex items-center px-3 py-1.5 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <ArrowDownTrayIcon className="h-4 w-4 mr-1" />
            {isDownloading ? 'Downloading...' : 'Download'}
          </button>
        </div>
      )}
    </div>
  )
}

export default ResourceCard
