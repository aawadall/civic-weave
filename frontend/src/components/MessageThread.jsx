import { useEffect, useRef } from 'react'

export default function MessageThread({ messages, currentUserId, onMessageClick }) {
  const messagesEndRef = useRef(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const formatTime = (dateString) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now - date
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMs / 3600000)
    const diffDays = Math.floor(diffMs / 86400000)

    if (diffMins < 1) return 'Just now'
    if (diffMins < 60) return `${diffMins}m ago`
    if (diffHours < 24) return `${diffHours}h ago`
    if (diffDays < 7) return `${diffDays}d ago`
    return date.toLocaleDateString()
  }

  if (!messages || messages.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 text-secondary-500">
        <div className="text-center">
          <p className="text-lg mb-2">💬</p>
          <p>No messages yet</p>
          <p className="text-sm">Be the first to start the conversation!</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex-1 overflow-y-auto space-y-4 p-4">
      {messages.map((message) => {
        const isOwnMessage = message.sender_id === currentUserId
        
        return (
          <div
            key={message.id}
            className={`flex ${isOwnMessage ? 'justify-end' : 'justify-start'}`}
          >
            <div className={`max-w-[70%] ${isOwnMessage ? 'order-2' : 'order-1'}`}>
              {/* Sender name (only for others' messages) */}
              {!isOwnMessage && (
                <div className="text-xs text-secondary-600 mb-1 px-3">
                  {message.sender_name}
                </div>
              )}
              
              {/* Message bubble */}
              <div
                className={`rounded-lg px-4 py-2 ${
                  isOwnMessage
                    ? 'bg-primary-600 text-white'
                    : 'bg-secondary-100 text-secondary-900'
                }`}
                onClick={() => onMessageClick && onMessageClick(message)}
              >
                <p className="text-sm whitespace-pre-wrap break-words">
                  {message.message_text}
                </p>
                
                {/* Timestamp and edited indicator */}
                <div className={`text-xs mt-1 flex items-center gap-2 ${
                  isOwnMessage ? 'text-primary-100' : 'text-secondary-500'
                }`}>
                  <span>{formatTime(message.created_at)}</span>
                  {message.edited_at && (
                    <span className="italic">(edited)</span>
                  )}
                </div>
              </div>
            </div>
          </div>
        )
      })}
      <div ref={messagesEndRef} />
    </div>
  )
}

