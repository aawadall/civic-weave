import { Link } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'

export default function AdminDashboard() {
  const { user } = useAuth()

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-secondary-900">
          Admin Dashboard
        </h1>
        <p className="text-secondary-600 mt-2">
          Manage initiatives, volunteers, and applications.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <div className="card">
          <h3 className="text-lg font-semibold text-secondary-900 mb-2">
            Total Volunteers
          </h3>
          <div className="text-2xl font-bold text-primary-600">0</div>
          <p className="text-sm text-secondary-500">Registered users</p>
        </div>

        <div className="card">
          <h3 className="text-lg font-semibold text-secondary-900 mb-2">
            Active Initiatives
          </h3>
          <div className="text-2xl font-bold text-primary-600">0</div>
          <p className="text-sm text-secondary-500">Current campaigns</p>
        </div>

        <div className="card">
          <h3 className="text-lg font-semibold text-secondary-900 mb-2">
            Pending Applications
          </h3>
          <div className="text-2xl font-bold text-primary-600">0</div>
          <p className="text-sm text-secondary-500">Awaiting review</p>
        </div>

        <div className="card">
          <h3 className="text-lg font-semibold text-secondary-900 mb-2">
            Completed Hours
          </h3>
          <div className="text-2xl font-bold text-primary-600">0</div>
          <p className="text-sm text-secondary-500">Volunteer hours</p>
        </div>
      </div>

      <div className="mt-8 grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="card">
          <h3 className="text-lg font-semibold text-secondary-900 mb-4">
            Recent Applications
          </h3>
          <div className="text-center py-8 text-secondary-500">
            <p>No applications to review yet.</p>
          </div>
        </div>

        <div className="card">
          <h3 className="text-lg font-semibold text-secondary-900 mb-4">
            Quick Actions
          </h3>
          <div className="space-y-3">
            <Link to="/admin/initiatives/create" className="w-full btn-primary text-left block">
              Create New Initiative
            </Link>
            <Link to="/admin/initiatives" className="w-full btn-secondary text-left block">
              Manage Initiatives
            </Link>
            <Link to="/admin/applications" className="w-full btn-secondary text-left block">
              Review Applications
            </Link>
            <Link to="/admin/skills" className="w-full btn-secondary text-left block">
              Manage Skill Claims
            </Link>
            <button className="w-full btn-secondary text-left">
              View All Volunteers
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
