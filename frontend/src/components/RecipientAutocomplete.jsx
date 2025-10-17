import React, { useState, useEffect, useRef } from 'react'
import { UserIcon, FolderIcon, MagnifyingGlassIcon } from '@heroicons/react/24/outline'
import { searchRecipients } from '../services/api'

const RecipientAutocomplete = ({ 
  value, 
  onChange, 
  placeholder = "Search for users or projects...",
  disabled = false 
}) => {
  const [inputValue, setInputValue] = useState('')
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [results, setResults] = useState({ users: [], projects: [] })
  const [selectedIndex, setSelectedIndex] = useState(-1)
  
  const inputRef = useRef(null)
  const dropdownRef = useRef(null)
  const debounceRef = useRef(null)

  // Initialize input value from selected recipient
  useEffect(() => {
    if (value) {
      setInputValue(value.name || value.title || '')
    } else {
      setInputValue('')
    }
  }, [value])

  // Debounced search
  useEffect(() => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current)
    }

    if (inputValue.length < 2) {
      setResults({ users: [], projects: [] })
      setIsOpen(false)
      return
    }

    debounceRef.current = setTimeout(async () => {
      setIsLoading(true)
      try {
        const response = await searchRecipients(inputValue)
        // Ensure we have the expected structure with fallbacks
        const data = response.data || {}
        setResults({
          users: Array.isArray(data.users) ? data.users : [],
          projects: Array.isArray(data.projects) ? data.projects : []
        })
        setIsOpen(true)
        setSelectedIndex(-1)
      } catch (error) {
        console.error('Failed to search recipients:', error)
        setResults({ users: [], projects: [] })
        setIsOpen(false)
      } finally {
        setIsLoading(false)
      }
    }, 300)
  }, [inputValue])

  // Handle clicks outside to close dropdown
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target) &&
          inputRef.current && !inputRef.current.contains(event.target)) {
        setIsOpen(false)
        setSelectedIndex(-1)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handleInputChange = (e) => {
    setInputValue(e.target.value)
    if (!e.target.value) {
      onChange(null)
    }
  }

  const handleInputFocus = () => {
    if (inputValue.length >= 2) {
      setIsOpen(true)
    } else if (inputValue.length > 0) {
      // Show a hint for minimum search length
      setIsOpen(true)
    }
  }

  const handleKeyDown = (e) => {
    if (!isOpen) return

    const allItems = [...results.users, ...results.projects]
    
    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault()
        setSelectedIndex(prev => 
          prev < allItems.length - 1 ? prev + 1 : prev
        )
        break
      case 'ArrowUp':
        e.preventDefault()
        setSelectedIndex(prev => prev > 0 ? prev - 1 : -1)
        break
      case 'Enter':
        e.preventDefault()
        if (selectedIndex >= 0 && selectedIndex < allItems.length) {
          selectItem(allItems[selectedIndex])
        }
        break
      case 'Escape':
        setIsOpen(false)
        setSelectedIndex(-1)
        break
    }
  }

  const selectItem = (item) => {
    onChange(item)
    setInputValue(item.name || item.title)
    setIsOpen(false)
    setSelectedIndex(-1)
  }

  const getDisplayName = (item) => {
    return item.name || item.title
  }

  const getSubtitle = (item) => {
    if (item.type === 'user') {
      return item.email
    }
    return null
  }

  const getIcon = (item) => {
    if (item.type === 'user') {
      return <UserIcon className="h-5 w-5 text-gray-400" />
    }
    return <FolderIcon className="h-5 w-5 text-gray-400" />
  }

  const allItems = [...results.users, ...results.projects]

  return (
    <div className="relative">
      <div className="relative">
        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
          <MagnifyingGlassIcon className="h-5 w-5 text-gray-400" />
        </div>
        <input
          ref={inputRef}
          type="text"
          value={inputValue}
          onChange={handleInputChange}
          onFocus={handleInputFocus}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          disabled={disabled}
          className="w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 disabled:bg-gray-50 disabled:cursor-not-allowed"
        />
        {isLoading && (
          <div className="absolute inset-y-0 right-0 pr-3 flex items-center">
            <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
          </div>
        )}
      </div>

      {/* Dropdown */}
      {isOpen && (
        <div
          ref={dropdownRef}
          className="absolute z-10 w-full mt-1 bg-white border border-gray-200 rounded-md shadow-lg max-h-60 overflow-y-auto"
        >
          {isLoading ? (
            <div className="px-3 py-3 text-center">
              <div className="flex items-center justify-center space-x-2">
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600"></div>
                <span className="text-sm text-gray-500">Searching...</span>
              </div>
            </div>
          ) : inputValue.length > 0 && inputValue.length < 2 ? (
            <div className="px-3 py-3 text-center">
              <div className="text-sm text-gray-500">
                Type at least 2 characters to search
              </div>
            </div>
          ) : allItems.length === 0 ? (
            <div className="px-3 py-3 text-center">
              <div className="text-sm text-gray-500 mb-1">
                No results found for "{inputValue}"
              </div>
              <div className="text-xs text-gray-400">
                Try a different search term or check spelling
              </div>
            </div>
          ) : (
            <>
              {/* Users Section */}
              {results.users.length > 0 && (
                <>
                  <div className="px-3 py-2 text-xs font-medium text-gray-500 bg-gray-50 border-b">
                    Users
                  </div>
                  {results.users.map((user, index) => (
                    <button
                      key={`user-${user.id}`}
                      onClick={() => selectItem(user)}
                      className={`w-full px-3 py-2 text-left hover:bg-gray-50 focus:bg-gray-50 focus:outline-none flex items-center space-x-3 ${
                        selectedIndex === index ? 'bg-gray-50' : ''
                      }`}
                    >
                      {getIcon(user)}
                      <div className="flex-1 min-w-0">
                        <div className="text-sm font-medium text-gray-900 truncate">
                          {getDisplayName(user)}
                        </div>
                        {getSubtitle(user) && (
                          <div className="text-sm text-gray-500 truncate">
                            {getSubtitle(user)}
                          </div>
                        )}
                      </div>
                    </button>
                  ))}
                </>
              )}

              {/* Projects Section */}
              {results.projects.length > 0 && (
                <>
                  {results.users.length > 0 && (
                    <div className="border-t border-gray-200"></div>
                  )}
                  <div className="px-3 py-2 text-xs font-medium text-gray-500 bg-gray-50 border-b">
                    Projects
                  </div>
                  {results.projects.map((project, index) => (
                    <button
                      key={`project-${project.id}`}
                      onClick={() => selectItem(project)}
                      className={`w-full px-3 py-2 text-left hover:bg-gray-50 focus:bg-gray-50 focus:outline-none flex items-center space-x-3 ${
                        selectedIndex === results.users.length + index ? 'bg-gray-50' : ''
                      }`}
                    >
                      {getIcon(project)}
                      <div className="flex-1 min-w-0">
                        <div className="text-sm font-medium text-gray-900 truncate">
                          {getDisplayName(project)}
                        </div>
                      </div>
                    </button>
                  ))}
                </>
              )}
            </>
          )}
        </div>
      )}
    </div>
  )
}

export default RecipientAutocomplete
