import { useState, useEffect } from 'react'
import { getCombinedVersionInfo } from '../utils/backendVersion.js'

export default function VersionDisplay({ compact = false, showInProduction = false }) {
  const [combinedVersion, setCombinedVersion] = useState(null)
  const [loading, setLoading] = useState(true)
  
  useEffect(() => {
    getCombinedVersionInfo()
      .then(setCombinedVersion)
      .catch(console.warn)
      .finally(() => setLoading(false))
  }, [])
  
  // Don't show in production unless explicitly requested
  if (process.env.NODE_ENV === 'production' && !showInProduction) {
    return null
  }
  
  if (loading) {
    return (
      <div className="text-xs text-gray-500">
        Loading version info...
      </div>
    )
  }
  
  if (compact) {
    return (
      <div className="text-xs text-gray-500 space-y-1">
        <div>FE: {combinedVersion?.frontend?.version || 'unknown'}</div>
        <div>BE: {combinedVersion?.backend?.version || 'unknown'}</div>
      </div>
    )
  }
  
  return (
    <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
      <h4 className="text-sm font-medium text-gray-800 mb-3">System Versions</h4>
      
      <div className="space-y-4 text-xs">
        {/* Frontend */}
        <div>
          <div className="font-medium text-blue-700 mb-2">Frontend</div>
          <div className="pl-2 space-y-1 text-gray-600">
            <div><strong>Version:</strong> {combinedVersion?.frontend?.version || 'unknown'}</div>
            <div><strong>Environment:</strong> {combinedVersion?.frontend?.build_env || 'unknown'}</div>
            <div><strong>Git Commit:</strong> {combinedVersion?.frontend?.git_commit || 'unknown'}</div>
            <div><strong>Build Time:</strong> {combinedVersion?.frontend?.build_time ? new Date(combinedVersion.frontend.build_time).toLocaleString() : 'unknown'}</div>
          </div>
        </div>
        
        {/* Backend */}
        <div>
          <div className="font-medium text-green-700 mb-2">Backend</div>
          <div className="pl-2 space-y-1 text-gray-600">
            <div><strong>Version:</strong> {combinedVersion?.backend?.version || 'unknown'}</div>
            <div><strong>Environment:</strong> {combinedVersion?.backend?.build_env || 'unknown'}</div>
            <div><strong>Git Commit:</strong> {combinedVersion?.backend?.git_commit || 'unknown'}</div>
            <div><strong>Build Time:</strong> {combinedVersion?.backend?.build_time ? new Date(combinedVersion.backend.build_time).toLocaleString() : 'unknown'}</div>
            <div><strong>Go Version:</strong> {combinedVersion?.backend?.go_version || 'unknown'}</div>
          </div>
        </div>
        
        {/* Sync Status */}
        <div className="pt-2 border-t border-gray-300">
          <div className="font-medium text-gray-700 mb-1">Sync Status</div>
          <div className="pl-2 text-gray-600">
            <div><strong>Last Updated:</strong> {combinedVersion?.timestamp ? new Date(combinedVersion.timestamp).toLocaleString() : 'unknown'}</div>
            <div><strong>API Connection:</strong> {combinedVersion?.backend ? '✅ Connected' : '❌ Disconnected'}</div>
          </div>
        </div>
      </div>
    </div>
  )
}
