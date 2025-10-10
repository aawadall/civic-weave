import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useToast } from '../../contexts/ToastContext'
import { ArrowLeftIcon } from '@heroicons/react/24/outline'
import api from '../../services/api'

const SKILL_OPTIONS = [
  'Event Planning',
  'Marketing',
  'Social Media',
  'Content Creation',
  'Graphic Design',
  'Photography',
  'Videography',
  'Writing',
  'Public Speaking',
  'Community Outreach',
  'Fundraising',
  'Project Management',
  'Data Analysis',
  'Research',
  'Teaching/Training',
  'Translation',
  'Legal',
  'Accounting',
  'Technology',
  'Web Development',
  'Mobile App Development',
  'Cybersecurity',
  'Healthcare',
  'Mental Health',
  'Elderly Care',
  'Childcare',
  'Environmental',
  'Sustainability',
  'Gardening',
  'Construction',
  'Handyman',
  'Transportation',
  'Cooking',
  'Catering',
  'Music',
  'Art',
  'Theater',
  'Sports',
  'Fitness',
  'Other'
]

export default function CreateInitiativePage() {
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    requiredSkills: [],
    locationAddress: '',
    startDate: '',
    endDate: '',
    status: 'draft'
  })
  const [loading, setLoading] = useState(false)
  const [errors, setErrors] = useState({})

  const { showToast } = useToast()
  const navigate = useNavigate()

  const handleInputChange = (e) => {
    const { name, value } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: value
    }))
    
    // Clear error when user starts typing
    if (errors[name]) {
      setErrors(prev => ({ ...prev, [name]: '' }))
    }
  }

  const handleSkillToggle = (skill) => {
    setFormData(prev => ({
      ...prev,
      requiredSkills: prev.requiredSkills.includes(skill)
        ? prev.requiredSkills.filter(s => s !== skill)
        : [...prev.requiredSkills, skill]
    }))
  }

  const validateForm = () => {
    const newErrors = {}

    if (!formData.title.trim()) {
      newErrors.title = 'Title is required'
    }

    if (!formData.description.trim()) {
      newErrors.description = 'Description is required'
    }

    if (formData.startDate && formData.endDate && formData.startDate > formData.endDate) {
      newErrors.endDate = 'End date must be after start date'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    
    if (!validateForm()) {
      return
    }

    setLoading(true)
    try {
      await api.post('/initiatives', formData)
      showToast('Initiative created successfully!', 'success')
      navigate('/admin')
    } catch (error) {
      showToast(error.response?.data?.error || 'Failed to create initiative', 'error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-8">
        <button
          onClick={() => navigate('/admin')}
          className="inline-flex items-center text-primary-600 hover:text-primary-500 mb-4"
        >
          <ArrowLeftIcon className="h-4 w-4 mr-1" />
          Back to Admin Dashboard
        </button>
        
        <h1 className="text-3xl font-bold text-secondary-900">
          Create New Initiative
        </h1>
        <p className="text-secondary-600 mt-2">
          Add a new volunteer opportunity to the platform
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="card">
          <h3 className="text-lg font-semibold text-secondary-900 mb-4">
            Basic Information
          </h3>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Title */}
            <div className="md:col-span-2">
              <label htmlFor="title" className="form-label">
                Initiative Title *
              </label>
              <input
                id="title"
                name="title"
                type="text"
                required
                className={`input-field ${errors.title ? 'border-red-500' : ''}`}
                placeholder="Enter initiative title"
                value={formData.title}
                onChange={handleInputChange}
              />
              {errors.title && <p className="mt-1 text-sm text-red-600">{errors.title}</p>}
            </div>

            {/* Description */}
            <div className="md:col-span-2">
              <label htmlFor="description" className="form-label">
                Description *
              </label>
              <textarea
                id="description"
                name="description"
                rows={4}
                required
                className={`input-field ${errors.description ? 'border-red-500' : ''}`}
                placeholder="Describe the initiative, what volunteers will do, and the impact they'll make"
                value={formData.description}
                onChange={handleInputChange}
              />
              {errors.description && <p className="mt-1 text-sm text-red-600">{errors.description}</p>}
            </div>

            {/* Location */}
            <div>
              <label htmlFor="locationAddress" className="form-label">
                Location
              </label>
              <input
                id="locationAddress"
                name="locationAddress"
                type="text"
                className="input-field"
                placeholder="City, State/Province (e.g., Toronto, ON)"
                value={formData.locationAddress}
                onChange={handleInputChange}
              />
              <p className="mt-1 text-xs text-secondary-500">
                This helps match volunteers by location
              </p>
            </div>

            {/* Status */}
            <div>
              <label htmlFor="status" className="form-label">
                Status
              </label>
              <select
                id="status"
                name="status"
                className="input-field"
                value={formData.status}
                onChange={handleInputChange}
              >
                <option value="draft">Draft</option>
                <option value="active">Active</option>
                <option value="closed">Closed</option>
              </select>
            </div>

            {/* Start Date */}
            <div>
              <label htmlFor="startDate" className="form-label">
                Start Date
              </label>
              <input
                id="startDate"
                name="startDate"
                type="date"
                className="input-field"
                value={formData.startDate}
                onChange={handleInputChange}
              />
            </div>

            {/* End Date */}
            <div>
              <label htmlFor="endDate" className="form-label">
                End Date
              </label>
              <input
                id="endDate"
                name="endDate"
                type="date"
                className={`input-field ${errors.endDate ? 'border-red-500' : ''}`}
                value={formData.endDate}
                onChange={handleInputChange}
              />
              {errors.endDate && <p className="mt-1 text-sm text-red-600">{errors.endDate}</p>}
            </div>
          </div>
        </div>

        <div className="card">
          <h3 className="text-lg font-semibold text-secondary-900 mb-4">
            Required Skills
          </h3>
          
          <div className="grid grid-cols-2 md:grid-cols-3 gap-2 max-h-48 overflow-y-auto border border-secondary-300 rounded-lg p-3">
            {SKILL_OPTIONS.map(skill => (
              <label key={skill} className="flex items-center space-x-2 text-sm">
                <input
                  type="checkbox"
                  checked={formData.requiredSkills.includes(skill)}
                  onChange={() => handleSkillToggle(skill)}
                  className="rounded border-secondary-300 text-primary-600 focus:ring-primary-500"
                />
                <span>{skill}</span>
              </label>
            ))}
          </div>
          <p className="mt-2 text-xs text-secondary-500">
            Select skills that volunteers need for this initiative
          </p>
        </div>

        <div className="flex justify-end space-x-4">
          <button
            type="button"
            onClick={() => navigate('/admin')}
            className="btn-secondary"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={loading}
            className="btn-primary"
          >
            {loading ? 'Creating...' : 'Create Initiative'}
          </button>
        </div>
      </form>
    </div>
  )
}
