export default function NotificationBadge({ count }) {
  if (!count || count === 0) return null

  return (
    <span className="absolute -top-1 -right-1 flex h-5 w-5 items-center justify-center rounded-full bg-red-600 text-xs font-semibold text-white">
      {count > 99 ? '99+' : count}
    </span>
  )
}

