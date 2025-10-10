import { useState, useEffect } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import { useToast } from '../../contexts/ToastContext';
import api from '../../services/api';
import { 
  UserCircleIcon, 
  ShieldCheckIcon, 
  CogIcon,
  ChartBarIcon,
  UsersIcon,
  ClipboardDocumentListIcon,
  AdjustmentsHorizontalIcon,
  EyeIcon,
  EyeSlashIcon,
  PencilIcon,
  CheckIcon,
  XMarkIcon
} from '@heroicons/react/24/outline';

export default function AdminProfilePage() {
  const { user, fetchUserProfile } = useAuth();
  const { showToast } = useToast();
  const [loading, setLoading] = useState(true);
  const [profileData, setProfileData] = useState({
    name: '',
    email: '',
    phone: '',
    locationAddress: '',
    preferences: {
      emailNotifications: true,
      skillReviewReminders: true,
      weeklyReports: false,
      instantAlerts: true
    }
  });
  const [systemStats, setSystemStats] = useState({
    totalVolunteers: 0,
    activeInitiatives: 0,
    pendingApplications: 0,
    totalSkillClaims: 0,
    lowWeightClaims: 0,
    recentActivity: []
  });
  const [editingField, setEditingField] = useState(null);
  const [tempValue, setTempValue] = useState('');

  useEffect(() => {
    if (user) {
      fetchProfileData();
      fetchSystemStats();
    }
  }, [user]);

  const fetchProfileData = async () => {
    try {
      setLoading(true);
      // Get admin profile data
      const response = await api.get('/admin/profile');
      setProfileData({
        name: response.data.name || '',
        email: response.data.email || '',
        phone: response.data.phone || '',
        locationAddress: response.data.location_address || '',
        preferences: response.data.preferences || {
          emailNotifications: true,
          skillReviewReminders: true,
          weeklyReports: false,
          instantAlerts: true
        }
      });
    } catch (error) {
      showToast('Failed to load profile data.', 'error');
    } finally {
      setLoading(false);
    }
  };

  const fetchSystemStats = async () => {
    try {
      const response = await api.get('/admin/stats');
      setSystemStats(response.data);
    } catch (error) {
      console.error('Failed to fetch system stats:', error);
    }
  };

  const handleEditField = (field, currentValue) => {
    setEditingField(field);
    setTempValue(currentValue);
  };

  const handleSaveField = async (field) => {
    try {
      const updateData = { [field]: tempValue };
      await api.put('/admin/profile', updateData);
      
      setProfileData(prev => ({
        ...prev,
        [field]: tempValue
      }));
      
      setEditingField(null);
      setTempValue('');
      showToast('Profile updated successfully!', 'success');
    } catch (error) {
      showToast(error.response?.data?.error || 'Failed to update profile.', 'error');
    }
  };

  const handleCancelEdit = () => {
    setEditingField(null);
    setTempValue('');
  };

  const handlePreferenceChange = async (preference, value) => {
    try {
      const newPreferences = { ...profileData.preferences, [preference]: value };
      await api.put('/admin/profile', { preferences: newPreferences });
      
      setProfileData(prev => ({
        ...prev,
        preferences: newPreferences
      }));
      
      showToast('Preferences updated successfully!', 'success');
    } catch (error) {
      showToast(error.response?.data?.error || 'Failed to update preferences.', 'error');
    }
  };

  const handleKeyPress = (e, field) => {
    if (e.key === 'Enter') {
      handleSaveField(field);
    } else if (e.key === 'Escape') {
      handleCancelEdit();
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto"></div>
          <p className="mt-4 text-secondary-600">Loading admin profile...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center space-x-3">
          <ShieldCheckIcon className="h-8 w-8 text-primary-600" />
          <div>
            <h1 className="text-3xl font-bold text-secondary-900">Admin Profile</h1>
            <p className="text-secondary-600 mt-1">
              Manage your administrator account and system preferences
            </p>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Left Column - Profile Information */}
        <div className="lg:col-span-2 space-y-6">
          {/* Personal Information */}
          <div className="bg-white shadow-md rounded-lg p-6">
            <div className="flex items-center space-x-2 mb-4">
              <UserCircleIcon className="h-5 w-5 text-secondary-500" />
              <h2 className="text-xl font-semibold text-secondary-800">Personal Information</h2>
            </div>
            
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-secondary-700 mb-1">
                  Full Name
                </label>
                {editingField === 'name' ? (
                  <div className="flex items-center space-x-2">
                    <input
                      type="text"
                      value={tempValue}
                      onChange={(e) => setTempValue(e.target.value)}
                      onKeyPress={(e) => handleKeyPress(e, 'name')}
                      className="input-field flex-grow"
                      autoFocus
                    />
                    <button
                      onClick={() => handleSaveField('name')}
                      className="text-green-600 hover:text-green-800"
                    >
                      <CheckIcon className="h-4 w-4" />
                    </button>
                    <button
                      onClick={handleCancelEdit}
                      className="text-red-600 hover:text-red-800"
                    >
                      <XMarkIcon className="h-4 w-4" />
                    </button>
                  </div>
                ) : (
                  <div className="flex items-center justify-between">
                    <span className="text-secondary-900">{profileData.name || 'Not set'}</span>
                    <button
                      onClick={() => handleEditField('name', profileData.name)}
                      className="text-secondary-400 hover:text-secondary-600"
                    >
                      <PencilIcon className="h-4 w-4" />
                    </button>
                  </div>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-secondary-700 mb-1">
                  Email Address
                </label>
                <div className="flex items-center justify-between">
                  <span className="text-secondary-900">{profileData.email}</span>
                  <span className="text-xs text-secondary-500 bg-secondary-100 px-2 py-1 rounded">
                    Primary
                  </span>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-secondary-700 mb-1">
                  Phone Number
                </label>
                {editingField === 'phone' ? (
                  <div className="flex items-center space-x-2">
                    <input
                      type="tel"
                      value={tempValue}
                      onChange={(e) => setTempValue(e.target.value)}
                      onKeyPress={(e) => handleKeyPress(e, 'phone')}
                      className="input-field flex-grow"
                      autoFocus
                    />
                    <button
                      onClick={() => handleSaveField('phone')}
                      className="text-green-600 hover:text-green-800"
                    >
                      <CheckIcon className="h-4 w-4" />
                    </button>
                    <button
                      onClick={handleCancelEdit}
                      className="text-red-600 hover:text-red-800"
                    >
                      <XMarkIcon className="h-4 w-4" />
                    </button>
                  </div>
                ) : (
                  <div className="flex items-center justify-between">
                    <span className="text-secondary-900">{profileData.phone || 'Not set'}</span>
                    <button
                      onClick={() => handleEditField('phone', profileData.phone)}
                      className="text-secondary-400 hover:text-secondary-600"
                    >
                      <PencilIcon className="h-4 w-4" />
                    </button>
                  </div>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-secondary-700 mb-1">
                  Location
                </label>
                {editingField === 'locationAddress' ? (
                  <div className="flex items-center space-x-2">
                    <input
                      type="text"
                      value={tempValue}
                      onChange={(e) => setTempValue(e.target.value)}
                      onKeyPress={(e) => handleKeyPress(e, 'locationAddress')}
                      className="input-field flex-grow"
                      autoFocus
                    />
                    <button
                      onClick={() => handleSaveField('locationAddress')}
                      className="text-green-600 hover:text-green-800"
                    >
                      <CheckIcon className="h-4 w-4" />
                    </button>
                    <button
                      onClick={handleCancelEdit}
                      className="text-red-600 hover:text-red-800"
                    >
                      <XMarkIcon className="h-4 w-4" />
                    </button>
                  </div>
                ) : (
                  <div className="flex items-center justify-between">
                    <span className="text-secondary-900">{profileData.locationAddress || 'Not set'}</span>
                    <button
                      onClick={() => handleEditField('locationAddress', profileData.locationAddress)}
                      className="text-secondary-400 hover:text-secondary-600"
                    >
                      <PencilIcon className="h-4 w-4" />
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* System Preferences */}
          <div className="bg-white shadow-md rounded-lg p-6">
            <div className="flex items-center space-x-2 mb-4">
              <CogIcon className="h-5 w-5 text-secondary-500" />
              <h2 className="text-xl font-semibold text-secondary-800">System Preferences</h2>
            </div>
            
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-sm font-medium text-secondary-900">Email Notifications</h3>
                  <p className="text-xs text-secondary-500">Receive email updates about system activities</p>
                </div>
                <button
                  onClick={() => handlePreferenceChange('emailNotifications', !profileData.preferences.emailNotifications)}
                  className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                    profileData.preferences.emailNotifications ? 'bg-primary-600' : 'bg-secondary-200'
                  }`}
                >
                  <span
                    className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                      profileData.preferences.emailNotifications ? 'translate-x-6' : 'translate-x-1'
                    }`}
                  />
                </button>
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-sm font-medium text-secondary-900">Skill Review Reminders</h3>
                  <p className="text-xs text-secondary-500">Get reminded to review skill claims and weights</p>
                </div>
                <button
                  onClick={() => handlePreferenceChange('skillReviewReminders', !profileData.preferences.skillReviewReminders)}
                  className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                    profileData.preferences.skillReviewReminders ? 'bg-primary-600' : 'bg-secondary-200'
                  }`}
                >
                  <span
                    className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                      profileData.preferences.skillReviewReminders ? 'translate-x-6' : 'translate-x-1'
                    }`}
                  />
                </button>
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-sm font-medium text-secondary-900">Weekly Reports</h3>
                  <p className="text-xs text-secondary-500">Receive weekly system performance reports</p>
                </div>
                <button
                  onClick={() => handlePreferenceChange('weeklyReports', !profileData.preferences.weeklyReports)}
                  className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                    profileData.preferences.weeklyReports ? 'bg-primary-600' : 'bg-secondary-200'
                  }`}
                >
                  <span
                    className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                      profileData.preferences.weeklyReports ? 'translate-x-6' : 'translate-x-1'
                    }`}
                  />
                </button>
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-sm font-medium text-secondary-900">Instant Alerts</h3>
                  <p className="text-xs text-secondary-500">Get immediate notifications for critical issues</p>
                </div>
                <button
                  onClick={() => handlePreferenceChange('instantAlerts', !profileData.preferences.instantAlerts)}
                  className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                    profileData.preferences.instantAlerts ? 'bg-primary-600' : 'bg-secondary-200'
                  }`}
                >
                  <span
                    className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                      profileData.preferences.instantAlerts ? 'translate-x-6' : 'translate-x-1'
                    }`}
                  />
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Right Column - System Overview */}
        <div className="space-y-6">
          {/* System Statistics */}
          <div className="bg-white shadow-md rounded-lg p-6">
            <div className="flex items-center space-x-2 mb-4">
              <ChartBarIcon className="h-5 w-5 text-secondary-500" />
              <h2 className="text-xl font-semibold text-secondary-800">System Overview</h2>
            </div>
            
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <UsersIcon className="h-4 w-4 text-blue-600" />
                  <span className="text-sm text-secondary-700">Total Volunteers</span>
                </div>
                <span className="text-lg font-semibold text-secondary-900">
                  {systemStats.totalVolunteers.toLocaleString()}
                </span>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <ClipboardDocumentListIcon className="h-4 w-4 text-green-600" />
                  <span className="text-sm text-secondary-700">Active Initiatives</span>
                </div>
                <span className="text-lg font-semibold text-secondary-900">
                  {systemStats.activeInitiatives.toLocaleString()}
                </span>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <AdjustmentsHorizontalIcon className="h-4 w-4 text-purple-600" />
                  <span className="text-sm text-secondary-700">Skill Claims</span>
                </div>
                <span className="text-lg font-semibold text-secondary-900">
                  {systemStats.totalSkillClaims.toLocaleString()}
                </span>
              </div>

              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <EyeSlashIcon className="h-4 w-4 text-red-600" />
                  <span className="text-sm text-secondary-700">Low Weight Claims</span>
                </div>
                <span className="text-lg font-semibold text-secondary-900">
                  {systemStats.lowWeightClaims.toLocaleString()}
                </span>
              </div>
            </div>
          </div>

          {/* Quick Actions */}
          <div className="bg-white shadow-md rounded-lg p-6">
            <h2 className="text-xl font-semibold text-secondary-800 mb-4">Quick Actions</h2>
            
            <div className="space-y-3">
              <a
                href="/admin/skills"
                className="w-full btn-primary text-center block"
              >
                Manage Skill Claims
              </a>
              <a
                href="/admin/initiatives"
                className="w-full btn-secondary text-center block"
              >
                Manage Initiatives
              </a>
              <a
                href="/admin/applications"
                className="w-full btn-secondary text-center block"
              >
                Review Applications
              </a>
              <button
                onClick={fetchSystemStats}
                className="w-full btn-secondary text-center"
              >
                Refresh Statistics
              </button>
            </div>
          </div>

          {/* Recent Activity */}
          <div className="bg-white shadow-md rounded-lg p-6">
            <h2 className="text-xl font-semibold text-secondary-800 mb-4">Recent Activity</h2>
            
            {systemStats.recentActivity.length === 0 ? (
              <p className="text-sm text-secondary-500">No recent activity</p>
            ) : (
              <div className="space-y-2">
                {systemStats.recentActivity.map((activity, index) => (
                  <div key={index} className="text-sm">
                    <p className="text-secondary-900">{activity.description}</p>
                    <p className="text-xs text-secondary-500">{activity.timestamp}</p>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
