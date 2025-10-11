import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'
import api from '../../services/api'

export default function VolunteerScorecardPage() {
  const { id } = useParams()
  const { hasAnyRole, hasRole } = useAuth()
  const [volunteer, setVolunteer] = useState(null)
  const [scorecard, setScorecard] = useState(null)
  const [ratings, setRatings] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [showAddRating, setShowAddRating] = useState(false)
  const [newRating, setNewRating] = useState({
    rating_type: 'thumbs_up',
    skill_name: '',
    comment: ''
  })
  const [submittingRating, setSubmittingRating] = useState(false)

  useEffect(() => {
    if (id) {
      fetchVolunteerData()
    }
  }, [id])

  const fetchVolunteerData = async () => {
    try {
      setLoading(true)
      const [volunteerResponse, scorecardResponse, ratingsResponse] = await Promise.all([
        api.get(`/volunteers/${id}`),
        api.get(`/volunteers/${id}/scorecard`),
        api.get(`/volunteers/${id}/ratings`)
      ])
      setVolunteer(volunteerResponse.data)
      setScorecard(scorecardResponse.data)
      setRatings(ratingsResponse.data.ratings || [])
    } catch (err) {
      setError('Failed to fetch volunteer data')
      console.error('Error fetching volunteer data:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleAddRating = async (e) => {
    e.preventDefault()
    setSubmittingRating(true)

    try {
      await api.post(`/volunteers/${id}/ratings`, newRating)
      setNewRating({
        rating_type: 'thumbs_up',
        skill_name: '',
        comment: ''
      })
      setShowAddRating(false)
      fetchVolunteerData() // Refresh data
    } catch (err) {
      console.error('Error adding rating:', err)
      alert('Failed to add rating. Please try again.')
    } finally {
      setSubmittingRating(false)
    }
  }

  const canRateVolunteer = () => {
    return hasAnyRole('team_lead', 'admin')
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  if (error || !volunteer) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-secondary-900 mb-4">Volunteer Not Found</h2>
          <p className="text-secondary-600 mb-4">{error || 'The requested volunteer could not be found.'}</p>
          <Link to="/volunteers" className="btn-primary">
            Back to Volunteers
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-secondary-50 py-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-8">
          <Link to="/volunteers" className="text-primary-600 hover:text-primary-800 text-sm font-medium mb-4">
            ← Back to Volunteers
          </Link>
          <h1 className="text-3xl font-bold text-secondary-900 mb-2">
            {volunteer.name || 'Unnamed Volunteer'} - Scorecard
          </h1>
          <p className="text-secondary-600">{volunteer.email}</p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Scorecard Summary */}
          <div className="lg:col-span-1">
            <div className="bg-white rounded-lg shadow-sm border border-secondary-200 p-6 mb-6">
              <h2 className="text-xl font-semibold text-secondary-900 mb-4">Overall Score</h2>
              {scorecard ? (
                <div className="text-center">
                  <div className="text-4xl font-bold text-primary-600 mb-2">
                    {scorecard.overall_score.toFixed(1)}
                  </div>
                  <div className="text-secondary-600 mb-4">
                    Based on {scorecard.total_ratings} ratings
                  </div>
                  <div className="grid grid-cols-3 gap-4 text-center">
                    <div>
                      <div className="text-2xl font-bold text-green-600">{scorecard.thumbs_up}</div>
                      <div className="text-sm text-secondary-600">👍</div>
                    </div>
                    <div>
                      <div className="text-2xl font-bold text-yellow-600">{scorecard.as_is}</div>
                      <div className="text-sm text-secondary-600">👌</div>
                    </div>
                    <div>
                      <div className="text-2xl font-bold text-red-600">{scorecard.thumbs_down}</div>
                      <div className="text-sm text-secondary-600">👎</div>
                    </div>
                  </div>
                </div>
              ) : (
                <div className="text-center text-secondary-600">
                  No ratings yet
                </div>
              )}
            </div>

            {/* Add Rating Form */}
            {canRateVolunteer() && (
              <div className="bg-white rounded-lg shadow-sm border border-secondary-200 p-6">
                {!showAddRating ? (
                  <button
                    onClick={() => setShowAddRating(true)}
                    className="w-full btn-primary"
                  >
                    Add Rating
                  </button>
                ) : (
                  <form onSubmit={handleAddRating} className="space-y-4">
                    <h3 className="text-lg font-semibold text-secondary-900">Add New Rating</h3>
                    
                    <div>
                      <label className="block text-sm font-medium text-secondary-900 mb-2">
                        Rating Type
                      </label>
                      <select
                        value={newRating.rating_type}
                        onChange={(e) => setNewRating(prev => ({ ...prev, rating_type: e.target.value }))}
                        className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                      >
                        <option value="thumbs_up">👍 Thumbs Up</option>
                        <option value="as_is">👌 As Is</option>
                        <option value="thumbs_down">👎 Thumbs Down</option>
                      </select>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-secondary-900 mb-2">
                        Skill (Optional)
                      </label>
                      <input
                        type="text"
                        value={newRating.skill_name}
                        onChange={(e) => setNewRating(prev => ({ ...prev, skill_name: e.target.value }))}
                        className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                        placeholder="Specific skill being rated"
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-secondary-900 mb-2">
                        Comment
                      </label>
                      <textarea
                        value={newRating.comment}
                        onChange={(e) => setNewRating(prev => ({ ...prev, comment: e.target.value }))}
                        rows={3}
                        className="w-full px-4 py-2 border border-secondary-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                        placeholder="Additional comments about the volunteer's performance"
                      />
                    </div>

                    <div className="flex gap-2">
                      <button
                        type="button"
                        onClick={() => setShowAddRating(false)}
                        className="flex-1 px-4 py-2 border border-secondary-300 text-secondary-700 rounded-lg hover:bg-secondary-50"
                      >
                        Cancel
                      </button>
                      <button
                        type="submit"
                        disabled={submittingRating}
                        className="flex-1 btn-primary"
                      >
                        {submittingRating ? 'Adding...' : 'Add Rating'}
                      </button>
                    </div>
                  </form>
                )}
              </div>
            )}
          </div>

          {/* Ratings History */}
          <div className="lg:col-span-2">
            <div className="bg-white rounded-lg shadow-sm border border-secondary-200 p-6">
              <h2 className="text-xl font-semibold text-secondary-900 mb-4">Rating History</h2>
              {ratings.length > 0 ? (
                <div className="space-y-4">
                  {ratings.map((rating) => (
                    <div key={rating.id} className="border border-secondary-200 rounded-lg p-4">
                      <div className="flex justify-between items-start mb-2">
                        <div className="flex items-center gap-2">
                          <span className="text-2xl">
                            {rating.rating_type === 'thumbs_up' ? '👍' :
                             rating.rating_type === 'thumbs_down' ? '👎' : '👌'}
                          </span>
                          <span className="font-medium text-secondary-900">
                            {rating.rating_type.replace('_', ' ').toUpperCase()}
                          </span>
                          {rating.skill_name && (
                            <span className="px-2 py-1 bg-primary-100 text-primary-800 text-xs rounded">
                              {rating.skill_name}
                            </span>
                          )}
                        </div>
                        <span className="text-sm text-secondary-500">
                          {new Date(rating.created_at).toLocaleDateString()}
                        </span>
                      </div>
                      {rating.comment && (
                        <p className="text-secondary-600 text-sm">{rating.comment}</p>
                      )}
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center py-8 text-secondary-600">
                  No ratings yet for this volunteer.
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
