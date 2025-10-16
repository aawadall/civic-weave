let cachedBackendVersion = null
let lastFetchTime = 0
const CACHE_DURATION = 5 * 60 * 1000 // 5 minutes

// Get API base URL from environment
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'

export async function getBackendVersion() {
  const now = Date.now()
  
  // Return cached version if still valid
  if (cachedBackendVersion && (now - lastFetchTime) < CACHE_DURATION) {
    return cachedBackendVersion
  }
  
  try {
    const response = await fetch(`${API_BASE_URL.replace('/api', '')}/version`)
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }
    
    const data = await response.json()
    cachedBackendVersion = data.version
    lastFetchTime = now
    
    console.log('Backend Version:', cachedBackendVersion)
    return cachedBackendVersion
  } catch (error) {
    console.warn('Failed to fetch backend version:', error.message)
    return {
      version: 'unknown',
      build_time: 'unknown',
      git_commit: 'unknown',
      build_env: 'unknown',
      go_version: 'unknown'
    }
  }
}

// Get combined version info for both frontend and backend
export async function getCombinedVersionInfo() {
  const frontendVersion = await import('../version.js')
  const backendVersion = await getBackendVersion()
  
  return {
    frontend: {
      version: frontendVersion.VERSION,
      build_time: frontendVersion.BUILD_TIME,
      git_commit: frontendVersion.GIT_COMMIT,
      build_env: frontendVersion.BUILD_ENV,
      api_base_url: frontendVersion.API_BASE_URL
    },
    backend: backendVersion,
    timestamp: Date.now()
  }
}

export async function getBackendHealth() {
  try {
    const response = await fetch(`${API_BASE_URL.replace('/api', '')}/health`)
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }
    
    const data = await response.json()
    return data
  } catch (error) {
    console.warn('Failed to fetch backend health:', error.message)
    return {
      status: 'error',
      version: {
        version: 'unknown',
        build_time: 'unknown',
        git_commit: 'unknown',
        build_env: 'unknown',
        go_version: 'unknown'
      }
    }
  }
}
