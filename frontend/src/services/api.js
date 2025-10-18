import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api',
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor to handle auth errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Token expired or invalid
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// API functions
export const getUnreadMessageCounts = () => {
  return api.get('/messages/unread-counts')
}

// Task Comments API
export const addTaskComment = (taskId, commentText) => {
  return api.post(`/tasks/${taskId}/comments`, { comment_text: commentText })
}

export const getTaskComments = (taskId) => {
  return api.get(`/tasks/${taskId}/comments`)
}

// Dashboard API
export const getUserProjects = () => {
  return api.get('/users/me/projects')
}

export const getUserTasks = () => {
  return api.get('/users/me/tasks')
}

export const getDashboardData = (params = {}) => {
  return api.get('/users/me/dashboard', { params })
}

// Universal Messaging API
export const sendMessage = (recipientType, recipientId, subject, messageText) => {
  return api.post('/messages', {
    recipient_type: recipientType,
    recipient_id: recipientId,
    subject,
    message_text: messageText
  })
}

export const getInbox = (params = {}) => {
  return api.get('/messages/inbox', { params })
}

export const searchRecipients = (query) => {
  return api.get('/messages/recipients/search', { params: { q: query } })
}

export const getSentMessages = (params = {}) => {
  return api.get('/messages/sent', { params })
}

export const getConversations = (params = {}) => {
  return api.get('/messages/conversations', { params })
}

export const getConversation = (conversationId, params = {}) => {
  return api.get(`/messages/conversations/${conversationId}`, { params })
}

export const getUniversalUnreadCount = () => {
  return api.get('/messages/unread-count')
}

export const markMessageAsRead = (messageId) => {
  return api.post(`/messages/${messageId}/read`)
}

// Broadcast API
export const getBroadcasts = (params = {}) => {
  return api.get('/broadcasts', { params })
}

export const getBroadcast = (id) => {
  return api.get(`/broadcasts/${id}`)
}

export const createBroadcast = (data) => {
  return api.post('/broadcasts', data)
}

export const updateBroadcast = (id, data) => {
  return api.put(`/broadcasts/${id}`, data)
}

export const deleteBroadcast = (id) => {
  return api.delete(`/broadcasts/${id}`)
}

export const markBroadcastRead = (id) => {
  return api.post(`/broadcasts/${id}/read`)
}

export const getBroadcastStats = () => {
  return api.get('/broadcasts/stats')
}

// Resource Library API
export const getResources = (params = {}) => {
  return api.get('/resources', { params })
}

export const getResource = (id) => {
  return api.get(`/resources/${id}`)
}

export const uploadResource = (formData) => {
  return api.post('/resources', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export const updateResource = (id, data) => {
  return api.put(`/resources/${id}`, data)
}

export const deleteResource = (id) => {
  return api.delete(`/resources/${id}`)
}

export const downloadResource = (id) => {
  return api.get(`/resources/${id}/download`, {
    responseType: 'blob'
  })
}

export const getResourceStats = () => {
  return api.get('/resources/stats')
}

export const getRecentResources = (params = {}) => {
  return api.get('/resources/recent', { params })
}

// Task Time Logging API
export const logTaskTime = (taskId, timeLogData) => {
  return api.post(`/tasks/${taskId}/time-logs`, timeLogData)
}

export const getTaskTimeLogs = (taskId) => {
  return api.get(`/tasks/${taskId}/time-logs`)
}

// Task Status Transitions API
export const startTask = (taskId) => {
  return api.post(`/tasks/${taskId}/start`)
}

export const markTaskBlocked = (taskId, reason) => {
  return api.post(`/tasks/${taskId}/mark-blocked`, { reason })
}

export const requestTaskTakeover = (taskId, reason) => {
  return api.post(`/tasks/${taskId}/request-takeover`, { reason })
}

export const markTaskDone = (taskId, completionNote) => {
  return api.post(`/tasks/${taskId}/mark-done`, { completion_note: completionNote })
}

// Task Assignment API
export const selfAssignTask = (taskId) => {
  return api.post(`/tasks/${taskId}/assign`)
}

export const assignTask = (taskId, volunteerId) => {
  return api.put(`/tasks/${taskId}/assign`, { volunteer_id: volunteerId })
}

export const unassignTask = (taskId) => {
  return api.put(`/tasks/${taskId}/assign`, { volunteer_id: null })
}

export default api
