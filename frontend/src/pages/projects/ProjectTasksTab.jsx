import { useState, useEffect } from 'react'
import { useAuth } from '../../contexts/AuthContext'
import api from '../../services/api'
import TaskCard from '../../components/TaskCard'
import TaskStatusBadge from '../../components/TaskStatusBadge'
import PriorityBadge from '../../components/PriorityBadge'
import TaskDetailModal from '../../components/TaskDetailModal'

export default function ProjectTasksTab({ projectId, isProjectOwner }) {
  const { user } = useAuth()
  const [tasks, setTasks] = useState([])
  const [unassignedTasks, setUnassignedTasks] = useState([])
  const [teamMembers, setTeamMembers] = useState([])
  const [loading, setLoading] = useState(true)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [selectedTask, setSelectedTask] = useState(null)
  const [showTaskDetailModal, setShowTaskDetailModal] = useState(false)
  const [newTask, setNewTask] = useState({
    title: '',
    description: '',
    priority: 'medium',
    due_date: '',
    labels: [],
    assignee_id: null
  })

  useEffect(() => {
    fetchTasks()
    if (isProjectOwner) {
      fetchTeamMembers()
    }
    if (!isProjectOwner) {
      fetchUnassignedTasks()
    }
  }, [projectId, isProjectOwner])

  const fetchTasks = async () => {
    try {
      setLoading(true)
      const response = await api.get(`/projects/${projectId}/tasks`)
      setTasks(response.data.tasks || [])
    } catch (error) {
      console.error('Error fetching tasks:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchUnassignedTasks = async () => {
    try {
      const response = await api.get(`/projects/${projectId}/tasks/unassigned`)
      setUnassignedTasks(response.data.tasks || [])
    } catch (error) {
      console.error('Error fetching unassigned tasks:', error)
    }
  }

  const fetchTeamMembers = async () => {
    try {
      const response = await api.get(`/projects/${projectId}/team-members`)
      setTeamMembers(response.data.team_members || [])
    } catch (error) {
      console.error('Error fetching team members:', error)
    }
  }

  const handleCreateTask = async (e) => {
    e.preventDefault()
    try {
      // Transform the data to match backend expectations
      const taskData = {
        title: newTask.title,
        description: newTask.description,
        priority: newTask.priority,
        labels: newTask.labels || []
      }
      
      // Only include assignee_id if it's provided
      if (newTask.assignee_id) {
        taskData.assignee_id = newTask.assignee_id
      }
      
      // Only include due_date if it's provided and not empty
      // Convert date string to ISO format for Go time.Time parsing
      if (newTask.due_date && newTask.due_date.trim() !== '') {
        // Convert "YYYY-MM-DD" to "YYYY-MM-DDTHH:MM:SSZ" format
        const date = new Date(newTask.due_date + 'T00:00:00Z')
        taskData.due_date = date.toISOString()
      }
      
      await api.post(`/projects/${projectId}/tasks`, taskData)
      setShowCreateModal(false)
      setNewTask({ title: '', description: '', priority: 'medium', due_date: '', labels: [], assignee_id: null })
      fetchTasks()
    } catch (error) {
      console.error('Error creating task:', error)
      alert('Failed to create task')
    }
  }

  const handleStatusChange = async (taskId, newStatus) => {
    try {
      await api.put(`/tasks/${taskId}`, { status: newStatus })
      fetchTasks()
    } catch (error) {
      console.error('Error updating task status:', error)
      alert('Failed to update task status')
    }
  }

  const handleSelfAssign = async (taskId) => {
    try {
      await api.post(`/tasks/${taskId}/assign`)
      fetchTasks()
      fetchUnassignedTasks()
    } catch (error) {
      console.error('Error self-assigning task:', error)
      alert('Failed to assign task')
    }
  }

  const handleTaskClick = (task) => {
    setSelectedTask(task)
    setShowTaskDetailModal(true)
  }

  const handleTaskUpdated = () => {
    fetchTasks()
    if (!isProjectOwner) {
      fetchUnassignedTasks()
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  // Kanban view for project owners
  if (isProjectOwner) {
    const todoTasks = tasks.filter(t => t.status === 'todo')
    const inProgressTasks = tasks.filter(t => t.status === 'in_progress')
    const doneTasks = tasks.filter(t => t.status === 'done')

    return (
      <div className="space-y-6">
        {/* Header */}
        <div className="flex justify-between items-center">
          <h2 className="text-2xl font-bold text-secondary-900">Tasks</h2>
          <button
            onClick={() => setShowCreateModal(true)}
            className="btn-primary"
          >
            + Create Task
          </button>
        </div>

        {/* Kanban Board */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {/* To Do Column */}
          <div className="bg-secondary-50 rounded-lg p-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-semibold text-secondary-900">To Do</h3>
              <span className="bg-secondary-200 text-secondary-700 text-xs font-medium px-2 py-1 rounded-full">
                {todoTasks.length}
              </span>
            </div>
            <div className="space-y-3">
              {todoTasks.map(task => (
                <TaskCard
                  key={task.id}
                  task={task}
                  onClick={() => handleTaskClick(task)}
                  onStatusChange={handleStatusChange}
                  canEdit={false}
                />
              ))}
              {todoTasks.length === 0 && (
                <p className="text-sm text-secondary-500 text-center py-8">No tasks</p>
              )}
            </div>
          </div>

          {/* In Progress Column */}
          <div className="bg-blue-50 rounded-lg p-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-semibold text-secondary-900">In Progress</h3>
              <span className="bg-blue-200 text-blue-700 text-xs font-medium px-2 py-1 rounded-full">
                {inProgressTasks.length}
              </span>
            </div>
            <div className="space-y-3">
              {inProgressTasks.map(task => (
                <TaskCard
                  key={task.id}
                  task={task}
                  onClick={() => handleTaskClick(task)}
                  onStatusChange={handleStatusChange}
                  canEdit={false}
                />
              ))}
              {inProgressTasks.length === 0 && (
                <p className="text-sm text-secondary-500 text-center py-8">No tasks</p>
              )}
            </div>
          </div>

          {/* Done Column */}
          <div className="bg-green-50 rounded-lg p-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-semibold text-secondary-900">Done</h3>
              <span className="bg-green-200 text-green-700 text-xs font-medium px-2 py-1 rounded-full">
                {doneTasks.length}
              </span>
            </div>
            <div className="space-y-3">
              {doneTasks.map(task => (
                <TaskCard
                  key={task.id}
                  task={task}
                  onClick={() => handleTaskClick(task)}
                  canEdit={false}
                />
              ))}
              {doneTasks.length === 0 && (
                <p className="text-sm text-secondary-500 text-center py-8">No tasks</p>
              )}
            </div>
          </div>
        </div>

        {/* Create Task Modal */}
        {showCreateModal && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
            <div className="bg-white rounded-lg max-w-2xl w-full p-6">
              <h3 className="text-xl font-bold text-secondary-900 mb-4">Create New Task</h3>
              <form onSubmit={handleCreateTask} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-secondary-900 mb-2">
                    Title *
                  </label>
                  <input
                    type="text"
                    value={newTask.title}
                    onChange={(e) => setNewTask({ ...newTask, title: e.target.value })}
                    required
                    className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-secondary-900 mb-2">
                    Description
                  </label>
                  <textarea
                    value={newTask.description}
                    onChange={(e) => setNewTask({ ...newTask, description: e.target.value })}
                    rows={3}
                    className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500"
                  />
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-secondary-900 mb-2">
                      Priority
                    </label>
                    <select
                      value={newTask.priority}
                      onChange={(e) => setNewTask({ ...newTask, priority: e.target.value })}
                      className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500"
                    >
                      <option value="low">Low</option>
                      <option value="medium">Medium</option>
                      <option value="high">High</option>
                    </select>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-secondary-900 mb-2">
                      Due Date <span className="text-secondary-500 text-sm">(optional)</span>
                    </label>
                    <input
                      type="date"
                      value={newTask.due_date}
                      onChange={(e) => setNewTask({ ...newTask, due_date: e.target.value })}
                      placeholder="Select due date"
                      className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500"
                    />
                  </div>
                </div>

                {/* Assignee Selection - Only for TLs */}
                {isProjectOwner && teamMembers.length > 0 && (
                  <div>
                    <label className="block text-sm font-medium text-secondary-900 mb-2">
                      Assign to <span className="text-secondary-500 text-sm">(optional)</span>
                    </label>
                    <select
                      value={newTask.assignee_id || ''}
                      onChange={(e) => setNewTask({ ...newTask, assignee_id: e.target.value || null })}
                      className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500"
                    >
                      <option value="">Unassigned</option>
                      {teamMembers.map(member => (
                        <option key={member.volunteer_id} value={member.volunteer_id}>
                          {member.volunteer_name || member.volunteer_email}
                        </option>
                      ))}
                    </select>
                  </div>
                )}

                <div className="flex justify-end gap-4">
                  <button
                    type="button"
                    onClick={() => setShowCreateModal(false)}
                    className="px-6 py-2 border border-secondary-300 text-secondary-700 rounded-lg hover:bg-secondary-50"
                  >
                    Cancel
                  </button>
                  <button type="submit" className="btn-primary">
                    Create Task
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}
      </div>
    )
  }

  // Volunteer view - My Tasks + Available Tasks
  const myTasks = tasks.filter(t => t.assignee_id)
  
  return (
    <div className="space-y-8">
      {/* My Tasks */}
      <div>
        <h2 className="text-2xl font-bold text-secondary-900 mb-4">My Tasks</h2>
        {myTasks.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {myTasks.map(task => (
              <TaskCard
                key={task.id}
                task={task}
                onClick={() => handleTaskClick(task)}
                onStatusChange={handleStatusChange}
                canEdit={true}
              />
            ))}
          </div>
        ) : (
          <div className="text-center py-12 bg-secondary-50 rounded-lg">
            <p className="text-secondary-600">You don't have any assigned tasks yet.</p>
            <p className="text-sm text-secondary-500 mt-2">Check available tasks below to self-assign!</p>
          </div>
        )}
      </div>

      {/* Available Tasks */}
      {unassignedTasks.length > 0 && (
        <div>
          <h2 className="text-2xl font-bold text-secondary-900 mb-4">Available Tasks</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {unassignedTasks.map(task => (
              <div key={task.id} className="relative">
                <TaskCard
                  task={task}
                  onClick={() => handleTaskClick(task)}
                />
                <button
                  onClick={() => handleSelfAssign(task.id)}
                  className="absolute top-2 right-2 px-3 py-1 bg-primary-600 text-white text-xs rounded hover:bg-primary-700"
                >
                  Assign to Me
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Task Detail Modal */}
      <TaskDetailModal
        task={selectedTask}
        isOpen={showTaskDetailModal}
        onClose={() => {
          setShowTaskDetailModal(false)
          setSelectedTask(null)
        }}
        onTaskUpdated={handleTaskUpdated}
      />
    </div>
  )
}

