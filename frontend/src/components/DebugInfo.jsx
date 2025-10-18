import { useState, useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { getVersionInfo, getVersionString } from '../version.js'
import { getCombinedVersionInfo } from '../utils/backendVersion.js'

export default function DebugInfo({ project, title = "Debug Info" }) {
  const { user } = useAuth()
  const versionInfo = getVersionInfo()
  const [combinedVersion, setCombinedVersion] = useState(null)
  
  useEffect(() => {
    getCombinedVersionInfo().then(setCombinedVersion)
  }, [])
  
  if (process.env.NODE_ENV === 'production') {
    return null
  }
  
  return (
    <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-4">
      <h3 className="text-sm font-medium text-yellow-800 mb-2">{title}</h3>
      <div className="text-xs text-yellow-700 space-y-1">
        {/* Frontend Version Info */}
        <div className="font-medium text-blue-700">Frontend</div>
        <div className="pl-2 space-y-1">
          <div><strong>Version:</strong> {versionInfo.version}</div>
          <div><strong>Environment:</strong> {versionInfo.build_env}</div>
          <div><strong>Git Commit:</strong> {versionInfo.git_commit}</div>
          <div><strong>Build Time:</strong> {new Date(versionInfo.build_time).toLocaleString()}</div>
          <div><strong>API URL:</strong> {versionInfo.api_base_url}</div>
        </div>
        
        <hr className="my-2 border-yellow-300" />
        
        {/* Backend Version Info */}
        <div className="font-medium text-green-700">Backend</div>
        {combinedVersion?.backend ? (
          <div className="pl-2 space-y-1">
            <div><strong>Version:</strong> {combinedVersion.backend.version}</div>
            <div><strong>Environment:</strong> {combinedVersion.backend.build_env}</div>
            <div><strong>Git Commit:</strong> {combinedVersion.backend.git_commit}</div>
            <div><strong>Build Time:</strong> {new Date(combinedVersion.backend.build_time).toLocaleString()}</div>
            <div><strong>Go Version:</strong> {combinedVersion.backend.go_version}</div>
          </div>
        ) : (
          <div className="pl-2 text-gray-500">Loading...</div>
        )}
        
        <hr className="my-2 border-yellow-300" />
        
        {/* User & Project Info */}
        <div className="font-medium text-purple-700">User & Project</div>
        <div className="pl-2 space-y-1">
          <div><strong>User:</strong> {user ? `${user.name} (${user.id})` : 'Not logged in'}</div>
          <div><strong>User Roles:</strong> {user?.roles?.join(', ') || 'None'}</div>
          <div><strong>Project:</strong> {project ? 'Loaded' : 'Not loaded'}</div>
          {project && (
            <>
              <div><strong>Project ID:</strong> {project.id}</div>
              <div><strong>Project Status:</strong> {project.project_status || 'Unknown'}</div>
              <div><strong>Team Lead ID:</strong> {project.team_lead_id || 'None'}</div>
              <div><strong>Created By:</strong> {project.created_by_admin_id || 'Unknown'}</div>
              <div><strong>Active Team Count:</strong> {project.active_team_count || 0}</div>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
