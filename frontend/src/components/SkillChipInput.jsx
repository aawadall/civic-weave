import { useState, useEffect, useRef } from 'react'
import { useToast } from '../contexts/ToastContext'
import api from '../services/api'
import { XMarkIcon, PlusIcon } from '@heroicons/react/24/outline'

export default function SkillChipInput({ 
  selectedSkills = [], 
  onChange, 
  placeholder = "Type skills or select from suggestions",
  maxSkills = 20,
  showDescription = false,
  disabled = false
}) {
  const [availableSkills, setAvailableSkills] = useState([])
  const [inputValue, setInputValue] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [showSuggestions, setShowSuggestions] = useState(false)
  const [filteredSuggestions, setFilteredSuggestions] = useState([])
  
  const inputRef = useRef(null)
  const suggestionsRef = useRef(null)
  const { showToast } = useToast()

  // Load available skills on mount
  useEffect(() => {
    fetchAvailableSkills()
  }, [])

  // Filter suggestions based on input
  useEffect(() => {
    if (inputValue.trim()) {
      const filtered = availableSkills.filter(skill =>
        skill.skill_name.toLowerCase().includes(inputValue.toLowerCase()) &&
        !selectedSkills.includes(skill.skill_name)
      ).slice(0, 10) // Limit to 10 suggestions
      setFilteredSuggestions(filtered)
      setShowSuggestions(true)
    } else {
      setFilteredSuggestions([])
      setShowSuggestions(false)
    }
  }, [inputValue, availableSkills, selectedSkills])

  // Handle clicks outside to close suggestions
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (suggestionsRef.current && !suggestionsRef.current.contains(event.target) &&
          inputRef.current && !inputRef.current.contains(event.target)) {
        setShowSuggestions(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const fetchAvailableSkills = async () => {
    try {
      setIsLoading(true)
      const response = await api.get('/skills/taxonomy')
      setAvailableSkills(response.data.skills || [])
    } catch (error) {
      showToast('Failed to load skills', 'error')
      console.error('Failed to fetch skills:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const handleInputChange = (e) => {
    const value = e.target.value
    setInputValue(value)
  }

  const handleInputKeyDown = (e) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      addSkillFromInput()
    } else if (e.key === 'Backspace' && !inputValue && selectedSkills.length > 0) {
      // Remove last skill if input is empty and backspace is pressed
      removeSkill(selectedSkills[selectedSkills.length - 1])
    }
  }

  const addSkillFromInput = () => {
    const skillName = inputValue.trim()
    if (!skillName) return

    // Parse comma-separated skills
    const skillsToAdd = skillName.split(',').map(s => s.trim()).filter(s => s.length > 0)
    
    for (const skill of skillsToAdd) {
      addSkill(skill)
    }
    
    setInputValue('')
    setShowSuggestions(false)
  }

  const addSkill = (skillName) => {
    if (selectedSkills.length >= maxSkills) {
      showToast(`Maximum ${maxSkills} skills allowed`, 'error')
      return
    }

    if (selectedSkills.includes(skillName)) {
      return // Skill already selected
    }

    const newSkills = [...selectedSkills, skillName]
    onChange(newSkills)
  }

  const removeSkill = (skillName) => {
    const newSkills = selectedSkills.filter(skill => skill !== skillName)
    onChange(newSkills)
  }

  const handleSuggestionClick = (skill) => {
    addSkill(skill.skill_name)
    setInputValue('')
    setShowSuggestions(false)
  }

  const getPopularSkills = () => {
    // Return first 12 skills as "popular" for quick selection
    return availableSkills.slice(0, 12)
  }

  return (
    <div className="space-y-4">
      {/* Selected Skills Display */}
      {selectedSkills.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {selectedSkills.map((skill, index) => (
            <div
              key={`${skill}-${index}`}
              className="inline-flex items-center gap-1 px-3 py-1 bg-primary-100 text-primary-800 rounded-full text-sm font-medium"
            >
              <span>{skill}</span>
              {!disabled && (
                <button
                  onClick={() => removeSkill(skill)}
                  className="ml-1 hover:bg-primary-200 rounded-full p-0.5"
                  type="button"
                >
                  <XMarkIcon className="h-3 w-3" />
                </button>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Input Field */}
      {!disabled && selectedSkills.length < maxSkills && (
        <div className="relative">
          <div className="flex items-center space-x-2">
            <input
              ref={inputRef}
              type="text"
              value={inputValue}
              onChange={handleInputChange}
              onKeyDown={handleInputKeyDown}
              onFocus={() => setShowSuggestions(!!inputValue)}
              placeholder={placeholder}
              className="flex-1 px-3 py-2 border border-secondary-300 rounded-lg focus:ring-primary-500 focus:border-primary-500"
              disabled={isLoading}
            />
            <button
              onClick={addSkillFromInput}
              disabled={!inputValue.trim() || isLoading}
              className="px-3 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
              type="button"
            >
              <PlusIcon className="h-4 w-4" />
            </button>
          </div>

          {/* Suggestions Dropdown */}
          {showSuggestions && filteredSuggestions.length > 0 && (
            <div
              ref={suggestionsRef}
              className="absolute z-10 w-full mt-1 bg-white border border-secondary-200 rounded-lg shadow-lg max-h-48 overflow-y-auto"
            >
              {filteredSuggestions.map((skill) => (
                <button
                  key={skill.id}
                  onClick={() => handleSuggestionClick(skill)}
                  className="w-full px-3 py-2 text-left hover:bg-secondary-50 focus:bg-secondary-50 focus:outline-none"
                  type="button"
                >
                  <span className="text-secondary-900">{skill.skill_name}</span>
                </button>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Popular Skills Grid */}
      {!disabled && selectedSkills.length === 0 && availableSkills.length > 0 && (
        <div>
          <h4 className="text-sm font-medium text-secondary-700 mb-2">Popular Skills</h4>
          <div className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-6 gap-2">
            {getPopularSkills().map((skill) => (
              <button
                key={skill.id}
                onClick={() => addSkill(skill.skill_name)}
                disabled={selectedSkills.includes(skill.skill_name)}
                className="px-3 py-2 text-xs bg-secondary-100 text-secondary-700 rounded-lg hover:bg-secondary-200 disabled:opacity-50 disabled:cursor-not-allowed text-center"
                type="button"
              >
                {skill.skill_name}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Helper Text */}
      <div className="text-xs text-secondary-500">
        {selectedSkills.length === 0 && (
          <span>Start typing to search or select from popular skills above</span>
        )}
        {selectedSkills.length > 0 && selectedSkills.length < maxSkills && (
          <span>{selectedSkills.length}/{maxSkills} skills selected. Type to add more or use commas to add multiple.</span>
        )}
        {selectedSkills.length >= maxSkills && (
          <span className="text-amber-600">Maximum skills reached. Remove some to add more.</span>
        )}
      </div>

      {/* Description Field (Optional) */}
      {showDescription && selectedSkills.length > 0 && (
        <div>
          <label className="block text-sm font-medium text-secondary-700 mb-2">
            Additional Details (Optional)
          </label>
          <textarea
            placeholder="Add any specific details about your experience with these skills..."
            className="w-full px-3 py-2 border border-secondary-300 rounded-lg focus:ring-primary-500 focus:border-primary-500 resize-none"
            rows={3}
            maxLength={500}
          />
        </div>
      )}
    </div>
  )
}
