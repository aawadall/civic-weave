import React, { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { getDashboardData, getBroadcasts, getRecentResources } from '../services/api'
import ProjectCard from '../components/ProjectCard'
import TaskListItem from '../components/TaskListItem'
import BroadcastCard from '../components/BroadcastCard'
import ResourceCard from '../components/ResourceCard'
import ComposeMessageModal from '../components/ComposeMessageModal'
import { 
  PlusIcon, 
  ChatBubbleLeftIcon, 
  BellIcon,
  DocumentTextIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon
} from '@heroicons/react/24/outline'

export default function VolunteerDashboard() {
  const { user } = useAuth()
  const [dashboardData, setDashboardData] = useState(null)
  const [broadcasts, setBroadcasts] = useState([])
  const [recentResources, setRecentResources] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showComposeModal, setShowComposeModal] = useState(false)

  useEffect(() => {
    loadDashboardData()
  }, [])

  const loadDashboardData = async () => {
    try {
      setLoading(true)
      const [dashboardResponse, broadcastsResponse, resourcesResponse] = await Promise.all([
        getDashboardData({ projects_limit: 5, tasks_limit: 10, broadcasts_limit: 5, resources_limit: 5 }),
        getBroadcasts({ limit: 5 }),
        getRecentResources({ limit: 5 })
      ])
      
      setDashboardData(dashboardResponse.data)
      setBroadcasts(broadcastsResponse.data.broadcasts || [])
      setRecentResources(resourcesResponse.data.resources || [])
    } catch (err) {
      console.error('Failed to load dashboard data:', err)
      setError('Failed to load dashboard data')
    } finally {
      setLoading(false)
    }
  }

  const handleBroadcastMarkAsRead = (broadcastId) => {
    setBroadcasts(prev => 
      prev.map(broadcast => 
        broadcast.id === broadcastId 
          ? { ...broadcast, is_read: true }
          : broadcast
      )
    )
  }

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="animate-pulse">
          <div className="h-8 bg-gray-200 rounded w-1/3 mb-4"></div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {[1, 2, 3].map(i => (
              <div key={i} className="h-32 bg-gray-200 rounded"></div>
            ))}
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="text-center">
          <p className="text-red-600">{error}</p>
          <button 
            onClick={loadDashboardData}
            className="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-secondary-900">
          Welcome back, {user?.name || 'Volunteer'}!
        </h1>
        <p className="text-secondary-600 mt-2">
          Manage your volunteer activities and discover new opportunities.
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
        <div className="card">
          <h3 className="text-lg font-semibold text-secondary-900 mb-2">
            My Projects
          </h3>
          <p className="text-secondary-600 text-sm mb-4">
            Projects you're enrolled in
          </p>
          <div className="text-2xl font-bold text-primary-600">
            {dashboardData?.stats?.total_projects || 0}
          </div>
          <p className="text-sm text-secondary-500">Active projects</p>
        </div>

        <div className="card">
          <h3 className="text-lg font-semibold text-secondary-900 mb-2">
            My Tasks
          </h3>
          <p className="text-secondary-600 text-sm mb-4">
            Tasks assigned to you
          </p>
          <div className="text-2xl font-bold text-primary-600">
            {dashboardData?.stats?.active_tasks || 0}
          </div>
          <p className="text-sm text-secondary-500">
            {dashboardData?.stats?.overdue_tasks || 0} overdue
          </p>
        </div>

        <div className="card">
          <h3 className="text-lg font-semibold text-secondary-900 mb-2">
            Messages
          </h3>
          <p className="text-secondary-600 text-sm mb-4">
            Unread messages
          </p>
          <div className="text-2xl font-bold text-primary-600">
            {dashboardData?.messages?.total || 0}
          </div>
          <p className="text-sm text-secondary-500">Unread messages</p>
        </div>

        <div className="card">
          <h3 className="text-lg font-semibold text-secondary-900 mb-2">
            Announcements
          </h3>
          <p className="text-secondary-600 text-sm mb-4">
            System announcements
          </p>
          <div className="text-2xl font-bold text-primary-600">
            {dashboardData?.stats?.unread_broadcasts || 0}
          </div>
          <p className="text-sm text-secondary-500">Unread announcements</p>
        </div>
      </div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
        {/* My Projects */}
        <div className="card">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-secondary-900">
              My Projects
            </h3>
            <Link 
              to="/projects" 
              className="text-sm text-blue-600 hover:text-blue-800"
            >
              View all
            </Link>
          </div>
          {dashboardData?.projects?.length > 0 ? (
            <div className="space-y-4">
              {dashboardData.projects.map(project => (
                <ProjectCard key={project.id} project={project} showStats={true} />
              ))}
            </div>
          ) : (
            <div className="text-center py-8">
              <p className="text-gray-500 mb-4">You're not enrolled in any projects yet</p>
              <Link 
                to="/projects"
                className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
              >
                <PlusIcon className="h-4 w-4 mr-2" />
                Browse Projects
              </Link>
            </div>
          )}
        </div>

        {/* My Tasks */}
        <div className="card">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-secondary-900">
              My Tasks
            </h3>
            <Link 
              to="/tasks" 
              className="text-sm text-blue-600 hover:text-blue-800"
            >
              View all
            </Link>
          </div>
          {dashboardData?.tasks?.length > 0 ? (
            <div className="space-y-3">
              {dashboardData.tasks.map(task => (
                <TaskListItem key={task.id} task={task} showProject={true} showActions={true} />
              ))}
            </div>
          ) : (
            <div className="text-center py-8">
              <CheckCircleIcon className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <p className="text-gray-500">No tasks assigned to you</p>
            </div>
          )}
        </div>
      </div>

      {/* Announcements */}
      {broadcasts.length > 0 && (
        <div className="mb-8">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-secondary-900">
              Announcements
            </h3>
            <Link 
              to="/messages" 
              className="text-sm text-blue-600 hover:text-blue-800"
            >
              View all
            </Link>
          </div>
          <div className="space-y-4">
            {broadcasts.map(broadcast => (
              <BroadcastCard 
                key={broadcast.id} 
                broadcast={broadcast} 
                onMarkAsRead={handleBroadcastMarkAsRead}
                showActions={true}
              />
            ))}
          </div>
        </div>
      )}

      {/* Resource Library */}
      <div className="mb-8">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-secondary-900">
            Resource Library
          </h3>
          <Link 
            to="/resources" 
            className="text-sm text-blue-600 hover:text-blue-800"
          >
            View all
          </Link>
        </div>
        {recentResources.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {recentResources.map(resource => (
              <ResourceCard key={resource.id} resource={resource} showActions={true} />
            ))}
          </div>
        ) : (
          <div className="text-center py-8">
            <DocumentTextIcon className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <p className="text-gray-500">No resources available yet</p>
          </div>
        )}
      </div>

      {/* Quick Actions */}
      <div className="card">
        <h3 className="text-lg font-semibold text-secondary-900 mb-4">
          Quick Actions
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Link 
            to="/projects"
            className="flex items-center p-4 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
          >
            <PlusIcon className="h-6 w-6 text-blue-600 mr-3" />
            <div>
              <h4 className="font-medium text-gray-900">Browse Projects</h4>
              <p className="text-sm text-gray-500">Find new opportunities</p>
            </div>
          </Link>
          
          <button 
            onClick={() => setShowComposeModal(true)}
            className="flex items-center p-4 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
          >
            <ChatBubbleLeftIcon className="h-6 w-6 text-green-600 mr-3" />
            <div>
              <h4 className="font-medium text-gray-900">Send Message</h4>
              <p className="text-sm text-gray-500">Contact team members</p>
            </div>
          </button>
          
          <Link 
            to="/resources"
            className="flex items-center p-4 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
          >
            <DocumentTextIcon className="h-6 w-6 text-purple-600 mr-3" />
            <div>
              <h4 className="font-medium text-gray-900">Resource Library</h4>
              <p className="text-sm text-gray-500">Browse resources</p>
            </div>
          </Link>
        </div>
      </div>

      {/* Compose Message Modal */}
      <ComposeMessageModal
        isOpen={showComposeModal}
        onClose={() => setShowComposeModal(false)}
        onSent={() => {
          setShowComposeModal(false)
          loadDashboardData() // Refresh data
        }}
      />
    </div>
  )
}
