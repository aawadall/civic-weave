import { useState } from 'react'
import { getVersionInfo, getVersionString } from '../version.js'
import { getCombinedVersionInfo } from '../utils/backendVersion.js'

export default function VersionToggle() {
  const [isOpen, setIsOpen] = useState(false)
  const [combinedVersion, setCombinedVersion] = useState(null)
  const frontendVersion = getVersionInfo()
  
  const handleToggle = async () => {
    if (!isOpen && !combinedVersion) {
      try {
        const combined = await getCombinedVersionInfo()
        setCombinedVersion(combined)
      } catch (error) {
        console.warn('Failed to fetch combined version info:', error)
      }
    }
    setIsOpen(!isOpen)
  }
  
  return (
    <div className="fixed bottom-4 right-4 z-50">
      <button
        onClick={handleToggle}
        className="bg-blue-600 hover:bg-blue-700 text-white text-xs px-2 py-1 rounded shadow-lg transition-colors"
        title="Show version info"
      >
        v{frontendVersion.version}
      </button>
      
      {isOpen && (
        <div className="absolute bottom-8 right-0 bg-white border border-gray-200 rounded-lg shadow-lg p-3 min-w-80 text-xs">
          <div className="font-medium text-gray-800 mb-3">Version Information</div>
          
          <div className="space-y-3 text-gray-600">
            {/* Frontend Section */}
            <div>
              <div className="font-medium text-blue-700 mb-1">Frontend</div>
              <div className="pl-2 space-y-1">
                <div><strong>Version:</strong> {frontendVersion.version}</div>
                <div><strong>Environment:</strong> {frontendVersion.build_env}</div>
                <div><strong>Git Commit:</strong> {frontendVersion.git_commit}</div>
                <div><strong>Build Time:</strong> {new Date(frontendVersion.build_time).toLocaleString()}</div>
                <div><strong>API URL:</strong> {frontendVersion.api_base_url}</div>
              </div>
            </div>
            
            {/* Backend Section */}
            <div>
              <div className="font-medium text-green-700 mb-1">Backend</div>
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
            </div>
          </div>
          
          <button
            onClick={() => setIsOpen(false)}
            className="mt-3 text-blue-600 hover:text-blue-800 text-xs"
          >
            Close
          </button>
        </div>
      )}
    </div>
  )
}
