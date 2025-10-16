// Frontend version information
export const VERSION = import.meta.env.VITE_VERSION || "1.0.0"
export const BUILD_TIME = new Date().toISOString()
export const GIT_COMMIT = import.meta.env.VITE_GIT_COMMIT || "unknown"
export const BUILD_ENV = import.meta.env.VITE_BUILD_ENV || "development"
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8081/api"

// Get version info as an object
export function getVersionInfo() {
  return {
    version: VERSION,
    build_time: BUILD_TIME,
    git_commit: GIT_COMMIT,
    build_env: BUILD_ENV,
    api_base_url: API_BASE_URL,
    user_agent: navigator.userAgent,
    timestamp: Date.now()
  }
}

// Get version string
export function getVersionString() {
  return `${VERSION}-${BUILD_ENV} (${GIT_COMMIT}) built at ${BUILD_TIME}`
}

// Log version info
console.log('Frontend Version:', getVersionString())
console.log('Version Info:', getVersionInfo())
