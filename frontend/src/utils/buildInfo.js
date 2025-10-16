import { getVersionInfo, getVersionString } from '../version.js'

// Build timestamp for cache busting
export const BUILD_TIMESTAMP = new Date().toISOString()
export const BUILD_VERSION = getVersionString()
export const VERSION_INFO = getVersionInfo()

console.log('Build Info:', { BUILD_TIMESTAMP, BUILD_VERSION, VERSION_INFO })
