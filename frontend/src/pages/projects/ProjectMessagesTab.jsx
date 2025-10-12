import { useState, useEffect, useRef } from 'react'
import { useAuth } from '../../contexts/AuthContext'
import api from '../../services/api'
import MessageThread from '../../components/MessageThread'

export default function ProjectMessagesTab({ projectId }) {
  const { user } = useAuth()
  const [messages, setMessages] = useState([])
  const [newMessage, setNewMessage] = useState('')
  const [loading, setLoading] = useState(true)
  const [sending, setSending] = useState(false)
  const pollingRef = useRef(null)
  const lastFetchTimeRef = useRef(new Date())

  useEffect(() => {
    fetchMessages()
    
    // Start polling for new messages every 3 seconds
    pollingRef.current = setInterval(() => {
      fetchNewMessages()
    }, 3000)

    return () => {
      if (pollingRef.current) {
        clearInterval(pollingRef.current)
      }
    }
  }, [projectId])

  const fetchMessages = async () => {
    try {
      setLoading(true)
      const response = await api.get(`/projects/${projectId}/messages/recent?count=50`)
      setMessages(response.data.messages || [])
      lastFetchTimeRef.current = new Date()
    } catch (error) {
      console.error('Error fetching messages:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchNewMessages = async () => {
    try {
      const response = await api.get(
        `/projects/${projectId}/messages/new?after=${lastFetchTimeRef.current.toISOString()}`
      )
      if (response.data.messages && response.data.messages.length > 0) {
        setMessages(prev => [...prev, ...response.data.messages])
        lastFetchTimeRef.current = new Date()
      }
    } catch (error) {
      console.error('Error fetching new messages:', error)
    }
  }

  const handleSendMessage = async (e) => {
    e.preventDefault()
    if (!newMessage.trim()) return

    try {
      setSending(true)
      const response = await api.post(`/projects/${projectId}/messages`, {
        message_text: newMessage
      })
      
      // Add the new message optimistically
      setMessages(prev => [...prev, response.data])
      setNewMessage('')
      lastFetchTimeRef.current = new Date()
    } catch (error) {
      console.error('Error sending message:', error)
      alert('Failed to send message')
    } finally {
      setSending(false)
    }
  }

  const handleMarkAllAsRead = async () => {
    try {
      await api.post(`/projects/${projectId}/messages/read-all`)
    } catch (error) {
      console.error('Error marking messages as read:', error)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-[600px] bg-white rounded-lg border border-secondary-200">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-secondary-200">
        <h2 className="text-lg font-semibold text-secondary-900">Project Messages</h2>
        <button
          onClick={handleMarkAllAsRead}
          className="text-sm text-primary-600 hover:text-primary-800"
        >
          Mark all as read
        </button>
      </div>

      {/* Messages */}
      <MessageThread
        messages={messages}
        currentUserId={user?.id}
      />

      {/* Input */}
      <form onSubmit={handleSendMessage} className="p-4 border-t border-secondary-200">
        <div className="flex gap-2">
          <input
            type="text"
            value={newMessage}
            onChange={(e) => setNewMessage(e.target.value)}
            placeholder="Type a message..."
            className="flex-1 px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            disabled={sending}
          />
          <button
            type="submit"
            disabled={sending || !newMessage.trim()}
            className="btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {sending ? 'Sending...' : 'Send'}
          </button>
        </div>
      </form>
    </div>
  )
}

