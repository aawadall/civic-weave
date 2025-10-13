import { Link } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'

export default function VolunteerDashboard() {
  const { user } = useAuth()

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-secondary-900">
          Welcome back, {user?.name || 'Volunteer'}!
        </h1>
        <p className="text-secondary-600 mt-2">
          Manage your volunteer activities and discover new opportunities.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="card">
          <h3 className="text-lg font-semibold text-secondary-900 mb-2">
            My Applications
          </h3>
          <p className="text-secondary-600 text-sm mb-4">
            Track your volunteer applications and their status.
          </p>
          <div className="text-2xl font-bold text-primary-600">0</div>
          <p className="text-sm text-secondary-500">Pending applications</p>
        </div>

        <div className="card">
          <h3 className="text-lg font-semibold text-secondary-900 mb-2">
            Hours Volunteered
          </h3>
          <p className="text-secondary-600 text-sm mb-4">
            Total hours contributed to community projects.
          </p>
          <div className="text-2xl font-bold text-primary-600">0</div>
          <p className="text-sm text-secondary-500">Hours this year</p>
        </div>

        <div className="card">
          <h3 className="text-lg font-semibold text-secondary-900 mb-2">
            Available Opportunities
          </h3>
          <p className="text-secondary-600 text-sm mb-4">
            New projects matching your skills and interests.
          </p>
          <div className="text-2xl font-bold text-primary-600">0</div>
          <p className="text-sm text-secondary-500">Recommended matches</p>
        </div>
      </div>

      <div className="mt-8 grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="card">
          <h3 className="text-lg font-semibold text-secondary-900 mb-4">
            Recent Activity
          </h3>
          <div className="text-center py-8 text-secondary-500">
            <p>No recent activity yet.</p>
            <p className="text-sm mt-2">Start by browsing available opportunities!</p>
          </div>
        </div>

        <div className="card">
          <h3 className="text-lg font-semibold text-secondary-900 mb-4">
            Quick Actions
          </h3>
          <div className="space-y-3">
            <Link to="/volunteer" className="w-full btn-primary text-left block">
              Browse Opportunities
            </Link>
            <button className="w-full btn-secondary text-left">
              View My Applications
            </button>
            <Link to="/profile" className="w-full btn-secondary text-left block">
              Update Profile
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}
