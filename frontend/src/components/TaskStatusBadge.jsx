export default function TaskStatusBadge({ status }) {
  const styles = {
    todo: 'bg-gray-100 text-gray-800',
    in_progress: 'bg-blue-100 text-blue-800',
    done: 'bg-green-100 text-green-800',
  }

  const labels = {
    todo: 'To Do',
    in_progress: 'In Progress',
    done: 'Done',
  }

  return (
    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${styles[status] || styles.todo}`}>
      {labels[status] || status}
    </span>
  )
}

