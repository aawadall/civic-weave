import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { getConversation, sendMessage, markMessageAsRead } from '../services/api'
import MessageThread from '../components/MessageThread'

export default function DirectConversationPage() {
  const { conversationId } = useParams()
  const { user } = useAuth()
  const navigate = useNavigate()
  const [messages, setMessages] = useState([])
  const [loading, setLoading] = useState(true)
  const [sending, setSending] = useState(false)
  const [error, setError] = useState('')
  const [replyForm, setReplyForm] = useState({
    subject: '',
    messageText: ''
  })

  useEffect(() => {
    fetchConversation()
  }, [conversationId])

  const fetchConversation = async () => {
    try {
      setLoading(true)
      setError('')
      const response = await getConversation(conversationId, { limit: 50 })
      const conversationMessages = response.data.messages || []
      setMessages(conversationMessages)

      // Mark all unread messages as read
      const unreadMessages = conversationMessages.filter(msg => !msg.is_read)
      if (unreadMessages.length > 0) {
        // Mark messages as read in parallel
        const markAsReadPromises = unreadMessages.map(msg => 
          markMessageAsRead(msg.id).catch(err => {
            console.error(`Failed to mark message ${msg.id} as read:`, err)
            return null // Don't block on individual failures
          })
        )
        
        await Promise.all(markAsReadPromises)
        
        // Optimistic update: mark messages as read in local state
        setMessages(prev => prev.map(msg => ({
          ...msg,
          is_read: true
        })))
      }
    } catch (error) {
      console.error('Error fetching conversation:', error)
      setError('Failed to load conversation')
    } finally {
      setLoading(false)
    }
  }

  const handleInputChange = (e) => {
    const { name, value } = e.target
    setReplyForm(prev => ({
      ...prev,
      [name]: value
    }))
  }

  const handleSendReply = async (e) => {
    e.preventDefault()
    if (!replyForm.messageText.trim()) {
      setError('Message text is required')
      return
    }

    try {
      setSending(true)
      setError('')
      
      await sendMessage(
        'user',
        conversationId,
        replyForm.subject || null,
        replyForm.messageText
      )
      
      // Clear form
      setReplyForm({
        subject: '',
        messageText: ''
      })
      
      // Refresh conversation to show new message
      await fetchConversation()
    } catch (error) {
      console.error('Failed to send reply:', error)
      setError(error.response?.data?.error || 'Failed to send message')
    } finally {
      setSending(false)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-secondary-50">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  if (error && messages.length === 0) {
    return (
      <div className="min-h-screen bg-secondary-50 py-8">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="bg-white rounded-lg border border-red-200 p-8 text-center">
            <div className="text-6xl mb-4">❌</div>
            <h3 className="text-lg font-medium text-red-900 mb-2">
              Error Loading Conversation
            </h3>
            <p className="text-red-600 mb-6">{error}</p>
            <button
              onClick={() => navigate('/messages')}
              className="btn-primary"
            >
              Back to Messages
            </button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-secondary-50 py-8">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-6">
          <button
            onClick={() => navigate('/messages')}
            className="flex items-center text-secondary-600 hover:text-secondary-800 mb-4"
          >
            <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
            Back to Messages
          </button>
          <h1 className="text-3xl font-bold text-secondary-900">Direct Message</h1>
          <p className="text-secondary-600 mt-2">
            Conversation with {messages[0]?.sender_name || 'Unknown User'}
          </p>
        </div>

        {/* Messages Thread */}
        <div className="bg-white rounded-lg border border-secondary-200 mb-6">
          <div className="h-[500px] flex flex-col">
            <MessageThread
              messages={messages}
              currentUserId={user?.id}
            />
          </div>
        </div>

        {/* Reply Composer */}
        <div className="bg-white rounded-lg border border-secondary-200 p-6">
          <h3 className="text-lg font-semibold text-secondary-900 mb-4">
            Send Reply
          </h3>
          
          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-sm text-red-600">{error}</p>
            </div>
          )}

          <form onSubmit={handleSendReply}>
            {/* Subject */}
            <div className="mb-4">
              <label className="block text-sm font-medium text-secondary-700 mb-2">
                Subject (optional)
              </label>
              <input
                type="text"
                name="subject"
                value={replyForm.subject}
                onChange={handleInputChange}
                placeholder="Enter message subject"
                className="w-full px-3 py-2 border border-secondary-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500"
              />
            </div>

            {/* Message Text */}
            <div className="mb-4">
              <label className="block text-sm font-medium text-secondary-700 mb-2">
                Message
              </label>
              <textarea
                name="messageText"
                value={replyForm.messageText}
                onChange={handleInputChange}
                rows={4}
                placeholder="Type your message here..."
                className="w-full px-3 py-2 border border-secondary-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                required
              />
            </div>

            {/* Send Button */}
            <div className="flex justify-end">
              <button
                type="submit"
                disabled={sending || !replyForm.messageText.trim()}
                className="btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {sending ? 'Sending...' : 'Send Reply'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}
