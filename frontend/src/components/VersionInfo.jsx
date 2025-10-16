import { getVersionInfo, getVersionString } from '../version.js'

export default function VersionInfo({ compact = false, showInProduction = false, showBackend = false }) {
  const versionInfo = getVersionInfo()
  
  // Don't show in production unless explicitly requested
  if (process.env.NODE_ENV === 'production' && !showInProduction) {
    return null
  }
  
  if (compact) {
    return (
      <div className="text-xs text-gray-500">
        v{versionInfo.version} ({versionInfo.build_env})
      </div>
    )
  }
  
  return (
    <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
      <h4 className="text-sm font-medium text-blue-800 mb-2">Version Information</h4>
      <div className="text-xs text-blue-700 space-y-1">
        <div><strong>Version:</strong> {versionInfo.version}</div>
        <div><strong>Environment:</strong> {versionInfo.build_env}</div>
        <div><strong>Build Time:</strong> {new Date(versionInfo.build_time).toLocaleString()}</div>
        <div><strong>Git Commit:</strong> {versionInfo.git_commit}</div>
        <div><strong>API Base URL:</strong> {versionInfo.api_base_url}</div>
      </div>
    </div>
  )
}
