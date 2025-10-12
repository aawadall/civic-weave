export default function PriorityBadge({ priority }) {
  const styles = {
    low: 'bg-gray-100 text-gray-700',
    medium: 'bg-yellow-100 text-yellow-800',
    high: 'bg-red-100 text-red-800',
  }

  const icons = {
    low: '↓',
    medium: '=',
    high: '↑',
  }

  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${styles[priority] || styles.medium}`}>
      <span className="mr-1">{icons[priority] || icons.medium}</span>
      {priority ? priority.charAt(0).toUpperCase() + priority.slice(1) : 'Medium'}
    </span>
  )
}

