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

// Task Time Logging API
export const logTaskTime = (taskId, timeLogData) => {
  return api.post(`/tasks/${taskId}/time-logs`, timeLogData)
}

export const getTaskTimeLogs = (taskId) => {
  return api.get(`/tasks/${taskId}/time-logs`)
}

// Task Status Transitions API
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
