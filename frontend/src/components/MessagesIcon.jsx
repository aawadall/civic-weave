import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { getUnreadMessageCounts } from '../services/api'
import NotificationBadge from './NotificationBadge'

export default function MessagesIcon() {
  const [unreadCount, setUnreadCount] = useState(0)

  const fetchUnreadCounts = async () => {
    try {
      const response = await getUnreadMessageCounts()
      // Sum up all unread counts across projects
      const total = response.data.unread_counts?.reduce((sum, item) => sum + item.count, 0) || 0
      setUnreadCount(total)
    } catch (error) {
      console.error('Failed to fetch unread message counts:', error)
    }
  }

  useEffect(() => {
    // Fetch immediately
    fetchUnreadCounts()

    // Poll every 30 seconds
    const interval = setInterval(fetchUnreadCounts, 30000)

    return () => clearInterval(interval)
  }, [])

  return (
    <Link to="/messages" className="relative p-2 text-secondary-600 hover:text-secondary-900 transition-colors">
      <svg
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        strokeWidth={1.5}
        stroke="currentColor"
        className="w-6 h-6"
        aria-label="Messages"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M21.75 6.75v10.5a2.25 2.25 0 01-2.25 2.25h-15a2.25 2.25 0 01-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25m19.5 0v.243a2.25 2.25 0 01-1.07 1.916l-7.5 4.615a2.25 2.25 0 01-2.36 0L3.32 8.91a2.25 2.25 0 01-1.07-1.916V6.75"
        />
      </svg>
      <NotificationBadge count={unreadCount} />
    </Link>
  )
}

