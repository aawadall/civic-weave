import { useState, useEffect } from 'react'
import { getVersionString } from '../version.js'
import { getCombinedVersionInfo } from '../utils/backendVersion.js'

export default function AppFooter() {
  const [combinedVersion, setCombinedVersion] = useState(null)
  const frontendVersion = getVersionString()
  
  useEffect(() => {
    getCombinedVersionInfo()
      .then(setCombinedVersion)
      .catch(console.warn)
  }, [])
  
  return (
    <footer className="bg-gray-50 border-t border-gray-200 py-4 px-6">
      <div className="max-w-7xl mx-auto flex justify-between items-center text-sm text-gray-600">
        <div>
          © 2024 CivicWeave. All rights reserved.
        </div>
        <div className="text-xs space-y-1">
          <div>FE: {combinedVersion?.frontend?.version || 'loading...'}</div>
          <div>BE: {combinedVersion?.backend?.version || 'loading...'}</div>
        </div>
      </div>
    </footer>
  )
}
