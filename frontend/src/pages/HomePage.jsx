import { Link } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'

export default function HomePage() {
  const { isAuthenticated } = useAuth()

  return (
    <div className="bg-white">
      {/* Hero Section */}
      <div className="relative bg-gradient-to-r from-primary-600 to-primary-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24">
          <div className="text-center">
            <h1 className="text-4xl md:text-6xl font-extrabold text-white mb-6">
              Connect Communities,
              <br />
              <span className="text-primary-200">Create Impact</span>
            </h1>
            <p className="text-xl text-primary-100 mb-8 max-w-3xl mx-auto">
              CivicWeave connects passionate volunteers with meaningful civic initiatives. 
              Join our community and make a real difference in your neighborhood.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              {!isAuthenticated ? (
                <>
                  <Link
                    to="/register"
                    className="bg-white text-primary-600 px-8 py-3 rounded-lg font-semibold hover:bg-primary-50 transition-colors"
                  >
                    Join as Volunteer
                  </Link>
                  <Link
                    to="/login"
                    className="border-2 border-white text-white px-8 py-3 rounded-lg font-semibold hover:bg-white hover:text-primary-600 transition-colors"
                  >
                    Sign In
                  </Link>
                </>
              ) : (
                <Link
                  to="/dashboard"
                  className="bg-white text-primary-600 px-8 py-3 rounded-lg font-semibold hover:bg-primary-50 transition-colors"
                >
                  Go to Dashboard
                </Link>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Features Section */}
      <div className="py-16 bg-secondary-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl font-extrabold text-secondary-900 mb-4">
              How CivicWeave Works
            </h2>
            <p className="text-xl text-secondary-600 max-w-2xl mx-auto">
              Simple, effective volunteer management for modern communities
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-8">
            <div className="text-center">
              <div className="bg-primary-100 w-16 h-16 rounded-full flex items-center justify-center mx-auto mb-4">
                <span className="text-2xl">👥</span>
              </div>
              <h3 className="text-xl font-semibold text-secondary-900 mb-2">
                Find Your Match
              </h3>
              <p className="text-secondary-600">
                Get matched with volunteer opportunities that align with your skills, 
                interests, and availability.
              </p>
            </div>

            <div className="text-center">
              <div className="bg-primary-100 w-16 h-16 rounded-full flex items-center justify-center mx-auto mb-4">
                <span className="text-2xl">🚀</span>
              </div>
              <h3 className="text-xl font-semibold text-secondary-900 mb-2">
                Make an Impact
              </h3>
              <p className="text-secondary-600">
                Contribute to meaningful civic initiatives and track your impact 
                in your community.
              </p>
            </div>

            <div className="text-center">
              <div className="bg-primary-100 w-16 h-16 rounded-full flex items-center justify-center mx-auto mb-4">
                <span className="text-2xl">📈</span>
              </div>
              <h3 className="text-xl font-semibold text-secondary-900 mb-2">
                Grow Together
              </h3>
              <p className="text-secondary-600">
                Build connections, develop new skills, and be part of a growing 
                community of changemakers.
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* CTA Section */}
      <div className="py-16 bg-white">
        <div className="max-w-4xl mx-auto text-center px-4 sm:px-6 lg:px-8">
          <h2 className="text-3xl font-extrabold text-secondary-900 mb-4">
            Ready to Make a Difference?
          </h2>
          <p className="text-xl text-secondary-600 mb-8">
            Join thousands of volunteers who are already creating positive change in their communities.
          </p>
          {!isAuthenticated && (
            <Link
              to="/register"
              className="btn-primary text-lg px-8 py-3"
            >
              Get Started Today
            </Link>
          )}
        </div>
      </div>
    </div>
  )
}
