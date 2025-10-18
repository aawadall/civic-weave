import TaskStatusBadge from './TaskStatusBadge'
import PriorityBadge from './PriorityBadge'

export default function TaskCard({ task, onClick, onStatusChange, canEdit = false, totalHours = 0 }) {
  const formatDate = (dateString) => {
    if (!dateString) return null
    return new Date(dateString).toLocaleDateString()
  }

  const handleStatusClick = (e, newStatus) => {
    e.stopPropagation()
    if (onStatusChange) {
      onStatusChange(task.id, newStatus)
    }
  }

  return (
    <div
      className={`bg-white rounded-lg border border-secondary-200 p-4 hover:shadow-md transition-shadow cursor-pointer ${
        task.status === 'done' ? 'opacity-75' : ''
      }`}
      onClick={onClick}
    >
      {/* Header */}
      <div className="flex items-start justify-between mb-2">
        <h3 className="font-medium text-secondary-900 flex-1">{task.title}</h3>
        <PriorityBadge priority={task.priority} />
      </div>

      {/* Description */}
      {task.description && (
        <p className="text-sm text-secondary-600 mb-3 line-clamp-2">{task.description}</p>
      )}

      {/* Labels */}
      {task.labels && task.labels.length > 0 && (
        <div className="flex flex-wrap gap-1 mb-3">
          {task.labels.map((label, index) => (
            <span
              key={index}
              className="inline-flex items-center px-2 py-0.5 rounded text-xs bg-secondary-100 text-secondary-700"
            >
              {label}
            </span>
          ))}
        </div>
      )}

      {/* Footer */}
      <div className="flex items-center justify-between text-sm">
        <div className="flex items-center gap-3">
          <TaskStatusBadge status={task.status} />
          
          {totalHours > 0 && (
            <span className="text-xs bg-blue-50 text-blue-700 px-2 py-1 rounded">
              ⏱️ {totalHours.toFixed(1)}h
            </span>
          )}
          
          {task.due_date && (
            <span className={`text-xs ${
              new Date(task.due_date) < new Date() && task.status !== 'done'
                ? 'text-red-600 font-medium'
                : 'text-secondary-500'
            }`}>
              📅 {formatDate(task.due_date)}
            </span>
          )}
        </div>

        {/* Quick status change for assignee */}
        {canEdit && task.status !== 'done' && task.status !== 'blocked' && task.status !== 'takeover_requested' && (
          <div className="flex gap-1">
            {task.status === 'todo' && (
              <button
                onClick={(e) => handleStatusClick(e, 'in_progress')}
                className="text-xs px-2 py-1 bg-blue-50 text-blue-700 rounded hover:bg-blue-100"
              >
                Start
              </button>
            )}
            {task.status === 'in_progress' && (
              <button
                onClick={(e) => handleStatusClick(e, 'done')}
                className="text-xs px-2 py-1 bg-green-50 text-green-700 rounded hover:bg-green-100"
              >
                Complete
              </button>
            )}
          </div>
        )}
      </div>

      {/* Assignee and project info */}
      <div className="mt-2 pt-2 border-t border-secondary-100">
        {task.assignee_id && (
          <div className="text-xs text-secondary-500 mb-1">
            <span className="font-medium">Assigned to:</span> {task.assignee_name || 'Unknown'}
            {task.assignee_email && (
              <span className="ml-1 text-secondary-400">({task.assignee_email})</span>
            )}
          </div>
        )}
        {task.project_title && (
          <div className="text-xs text-secondary-500 mb-1">
            <span className="font-medium">Project:</span> {task.project_title}
            {task.project_status && (
              <span className={`ml-1 px-1 py-0.5 rounded text-xs ${
                task.project_status === 'active' ? 'bg-green-100 text-green-700' :
                task.project_status === 'recruiting' ? 'bg-blue-100 text-blue-700' :
                task.project_status === 'completed' ? 'bg-gray-100 text-gray-700' :
                'bg-yellow-100 text-yellow-700'
              }`}>
                {task.project_status}
              </span>
            )}
          </div>
        )}
        
        {/* Timeline badges */}
        {(task.blocked_at || task.completed_at || task.takeover_requested_at) && (
          <div className="flex flex-wrap gap-1 mt-1">
            {task.blocked_at && (
              <span className="text-xs bg-red-50 text-red-700 px-1 py-0.5 rounded">
                🚫 Blocked {new Date(task.blocked_at).toLocaleDateString()}
              </span>
            )}
            {task.completed_at && (
              <span className="text-xs bg-green-50 text-green-700 px-1 py-0.5 rounded">
                ✅ Completed {new Date(task.completed_at).toLocaleDateString()}
              </span>
            )}
            {task.takeover_requested_at && (
              <span className="text-xs bg-orange-50 text-orange-700 px-1 py-0.5 rounded">
                🔄 Takeover requested {new Date(task.takeover_requested_at).toLocaleDateString()}
              </span>
            )}
          </div>
        )}
      </div>
    </div>
  )
}

