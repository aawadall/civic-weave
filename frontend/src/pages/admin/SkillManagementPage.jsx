import { useState, useEffect } from 'react';
import { useToast } from '../../contexts/ToastContext';
import api from '../../services/api';
import { 
  PencilIcon, 
  CheckIcon, 
  XMarkIcon,
  EyeIcon,
  EyeSlashIcon,
  FunnelIcon,
  AdjustmentsHorizontalIcon
} from '@heroicons/react/24/outline';

export default function SkillManagementPage() {
  const { showToast } = useToast();
  const [loading, setLoading] = useState(true);
  const [skillClaims, setSkillClaims] = useState([]);
  const [volunteers, setVolunteers] = useState({});
  const [editingClaim, setEditingClaim] = useState(null);
  const [tempWeight, setTempWeight] = useState(0);
  const [filters, setFilters] = useState({
    minWeight: 0.1,
    maxWeight: 1.0,
    showInactive: false,
    searchTerm: ''
  });

  useEffect(() => {
    fetchSkillClaims();
  }, []);

  const fetchSkillClaims = async () => {
    try {
      setLoading(true);
      const response = await api.get('/admin/skill-claims');
      setSkillClaims(response.data);
      
      // Fetch volunteer details for each claim
      const volunteerIds = [...new Set(response.data.map(claim => claim.volunteer_id))];
      const volunteerPromises = volunteerIds.map(async (volunteerId) => {
        try {
          const volunteerResponse = await api.get(`/volunteers/${volunteerId}`);
          return { id: volunteerId, ...volunteerResponse.data };
        } catch (error) {
          console.error(`Failed to fetch volunteer ${volunteerId}:`, error);
          return { id: volunteerId, name: 'Unknown Volunteer' };
        }
      });
      
      const volunteerData = await Promise.all(volunteerPromises);
      const volunteerMap = {};
      volunteerData.forEach(volunteer => {
        volunteerMap[volunteer.id] = volunteer;
      });
      setVolunteers(volunteerMap);
    } catch (error) {
      showToast('Failed to load skill claims.', 'error');
    } finally {
      setLoading(false);
    }
  };

  const handleEditWeight = (claim) => {
    setEditingClaim(claim.id);
    setTempWeight(claim.claim_weight);
  };

  const handleSaveWeight = async (claimId) => {
    try {
      await api.patch(`/admin/skill-claims/${claimId}/weight`, {
        claim_weight: tempWeight
      });
      
      // Update local state
      setSkillClaims(prev => prev.map(claim => 
        claim.id === claimId 
          ? { ...claim, claim_weight: tempWeight, updated_at: new Date().toISOString() }
          : claim
      ));
      
      setEditingClaim(null);
      showToast('Skill weight updated successfully!', 'success');
    } catch (error) {
      showToast(error.response?.data?.error || 'Failed to update skill weight.', 'error');
    }
  };

  const handleCancelEdit = () => {
    setEditingClaim(null);
    setTempWeight(0);
  };

  const toggleClaimVisibility = async (claimId, currentStatus) => {
    try {
      await api.patch(`/admin/skill-claims/${claimId}/visibility`, {
        is_active: !currentStatus
      });
      
      // Update local state
      setSkillClaims(prev => prev.map(claim => 
        claim.id === claimId 
          ? { ...claim, is_active: !currentStatus }
          : claim
      ));
      
      showToast(`Skill claim ${!currentStatus ? 'activated' : 'deactivated'} successfully!`, 'success');
    } catch (error) {
      showToast(error.response?.data?.error || 'Failed to update skill visibility.', 'error');
    }
  };

  const getWeightColor = (weight) => {
    if (weight >= 0.8) return 'text-green-600 bg-green-100';
    if (weight >= 0.6) return 'text-yellow-600 bg-yellow-100';
    if (weight >= 0.4) return 'text-orange-600 bg-orange-100';
    return 'text-red-600 bg-red-100';
  };

  const getWeightLabel = (weight) => {
    if (weight >= 0.8) return 'High';
    if (weight >= 0.6) return 'Good';
    if (weight >= 0.4) return 'Fair';
    return 'Low';
  };

  const filteredClaims = skillClaims.filter(claim => {
    const volunteer = volunteers[claim.volunteer_id];
    const volunteerName = volunteer?.name || 'Unknown';
    
    // Weight filter
    if (claim.claim_weight < filters.minWeight || claim.claim_weight > filters.maxWeight) {
      return false;
    }
    
    // Active/Inactive filter
    if (!filters.showInactive && !claim.is_active) {
      return false;
    }
    
    // Search filter
    if (filters.searchTerm) {
      const searchLower = filters.searchTerm.toLowerCase();
      return claim.skill_name.toLowerCase().includes(searchLower) ||
             volunteerName.toLowerCase().includes(searchLower);
    }
    
    return true;
  });

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto"></div>
          <p className="mt-4 text-secondary-600">Loading skill claims...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-secondary-900">Skill Claims Management</h1>
        <p className="mt-2 text-secondary-600">
          Manage volunteer skill claims and adjust weights for better matching accuracy.
        </p>
      </div>

      {/* Filters */}
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <div className="flex items-center space-x-4 mb-4">
          <FunnelIcon className="h-5 w-5 text-secondary-500" />
          <h3 className="text-lg font-medium text-secondary-900">Filters</h3>
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div>
            <label className="block text-sm font-medium text-secondary-700 mb-1">
              Search
            </label>
            <input
              type="text"
              value={filters.searchTerm}
              onChange={(e) => setFilters(prev => ({ ...prev, searchTerm: e.target.value }))}
              placeholder="Search by skill or volunteer name..."
              className="input-field"
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-secondary-700 mb-1">
              Min Weight
            </label>
            <input
              type="number"
              min="0.1"
              max="1.0"
              step="0.1"
              value={filters.minWeight}
              onChange={(e) => setFilters(prev => ({ ...prev, minWeight: parseFloat(e.target.value) }))}
              className="input-field"
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-secondary-700 mb-1">
              Max Weight
            </label>
            <input
              type="number"
              min="0.1"
              max="1.0"
              step="0.1"
              value={filters.maxWeight}
              onChange={(e) => setFilters(prev => ({ ...prev, maxWeight: parseFloat(e.target.value) }))}
              className="input-field"
            />
          </div>
          
          <div className="flex items-end">
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={filters.showInactive}
                onChange={(e) => setFilters(prev => ({ ...prev, showInactive: e.target.checked }))}
                className="rounded border-secondary-300 text-primary-600 focus:ring-primary-500"
              />
              <span className="ml-2 text-sm text-secondary-700">Show inactive claims</span>
            </label>
          </div>
        </div>
      </div>

      {/* Statistics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-white shadow rounded-lg p-4">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <AdjustmentsHorizontalIcon className="h-8 w-8 text-blue-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-secondary-500">Total Claims</p>
              <p className="text-2xl font-semibold text-secondary-900">{skillClaims.length}</p>
            </div>
          </div>
        </div>
        
        <div className="bg-white shadow rounded-lg p-4">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <EyeIcon className="h-8 w-8 text-green-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-secondary-500">Active Claims</p>
              <p className="text-2xl font-semibold text-secondary-900">
                {skillClaims.filter(claim => claim.is_active).length}
              </p>
            </div>
          </div>
        </div>
        
        <div className="bg-white shadow rounded-lg p-4">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <EyeSlashIcon className="h-8 w-8 text-red-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-secondary-500">Inactive Claims</p>
              <p className="text-2xl font-semibold text-secondary-900">
                {skillClaims.filter(claim => !claim.is_active).length}
              </p>
            </div>
          </div>
        </div>
        
        <div className="bg-white shadow rounded-lg p-4">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <div className="h-8 w-8 bg-purple-100 rounded-full flex items-center justify-center">
                <span className="text-purple-600 font-bold text-sm">AVG</span>
              </div>
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-secondary-500">Avg Weight</p>
              <p className="text-2xl font-semibold text-secondary-900">
                {skillClaims.length > 0 
                  ? (skillClaims.reduce((sum, claim) => sum + claim.claim_weight, 0) / skillClaims.length).toFixed(2)
                  : '0.00'
                }
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Skill Claims Table */}
      <div className="bg-white shadow overflow-hidden sm:rounded-lg">
        <div className="px-4 py-5 sm:px-6">
          <h3 className="text-lg leading-6 font-medium text-secondary-900">
            Skill Claims ({filteredClaims.length} of {skillClaims.length})
          </h3>
          <p className="mt-1 max-w-2xl text-sm text-secondary-500">
            Manage individual skill claims and their weights for matching accuracy.
          </p>
        </div>
        
        {filteredClaims.length === 0 ? (
          <div className="text-center py-12">
            <AdjustmentsHorizontalIcon className="mx-auto h-12 w-12 text-secondary-400" />
            <h3 className="mt-2 text-sm font-medium text-secondary-900">No skill claims found</h3>
            <p className="mt-1 text-sm text-secondary-500">
              Try adjusting your filters to see more results.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-secondary-200">
              <thead className="bg-secondary-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-secondary-500 uppercase tracking-wider">
                    Volunteer
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-secondary-500 uppercase tracking-wider">
                    Skill Claim
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-secondary-500 uppercase tracking-wider">
                    Proficiency
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-secondary-500 uppercase tracking-wider">
                    Weight
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-secondary-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-secondary-500 uppercase tracking-wider">
                    Created
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-secondary-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-secondary-200">
                {filteredClaims.map((claim) => {
                  const volunteer = volunteers[claim.volunteer_id];
                  return (
                    <tr key={claim.id} className={claim.is_active ? '' : 'bg-secondary-50'}>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-secondary-900">
                          {volunteer?.name || 'Unknown Volunteer'}
                        </div>
                        <div className="text-sm text-secondary-500">
                          {volunteer?.email || 'No email'}
                        </div>
                      </td>
                      
                      <td className="px-6 py-4">
                        <div className="text-sm text-secondary-900 max-w-xs truncate">
                          {claim.skill_name}
                        </div>
                      </td>
                      
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                          Level {claim.proficiency_level}
                        </span>
                      </td>
                      
                      <td className="px-6 py-4 whitespace-nowrap">
                        {editingClaim === claim.id ? (
                          <div className="flex items-center space-x-2">
                            <input
                              type="number"
                              min="0.1"
                              max="1.0"
                              step="0.05"
                              value={tempWeight}
                              onChange={(e) => setTempWeight(parseFloat(e.target.value))}
                              className="w-20 px-2 py-1 border border-secondary-300 rounded text-sm"
                            />
                            <button
                              onClick={() => handleSaveWeight(claim.id)}
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
                          <div className="flex items-center space-x-2">
                            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getWeightColor(claim.claim_weight)}`}>
                              {claim.claim_weight.toFixed(2)} ({getWeightLabel(claim.claim_weight)})
                            </span>
                            <button
                              onClick={() => handleEditWeight(claim)}
                              className="text-secondary-400 hover:text-secondary-600"
                            >
                              <PencilIcon className="h-4 w-4" />
                            </button>
                          </div>
                        )}
                      </td>
                      
                      <td className="px-6 py-4 whitespace-nowrap">
                        <button
                          onClick={() => toggleClaimVisibility(claim.id, claim.is_active)}
                          className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            claim.is_active 
                              ? 'bg-green-100 text-green-800' 
                              : 'bg-red-100 text-red-800'
                          }`}
                        >
                          {claim.is_active ? (
                            <>
                              <EyeIcon className="h-3 w-3 mr-1" />
                              Active
                            </>
                          ) : (
                            <>
                              <EyeSlashIcon className="h-3 w-3 mr-1" />
                              Inactive
                            </>
                          )}
                        </button>
                      </td>
                      
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-secondary-500">
                        {new Date(claim.created_at).toLocaleDateString()}
                      </td>
                      
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                        <div className="flex space-x-2">
                          <button
                            onClick={() => handleEditWeight(claim)}
                            className="text-primary-600 hover:text-primary-900"
                            disabled={editingClaim === claim.id}
                          >
                            <PencilIcon className="h-4 w-4" />
                          </button>
                          <button
                            onClick={() => toggleClaimVisibility(claim.id, claim.is_active)}
                            className={claim.is_active ? 'text-red-600 hover:text-red-900' : 'text-green-600 hover:text-green-900'}
                          >
                            {claim.is_active ? <EyeSlashIcon className="h-4 w-4" /> : <EyeIcon className="h-4 w-4" />}
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Help Text */}
      <div className="mt-6 bg-blue-50 border border-blue-200 rounded-lg p-4">
        <div className="flex">
          <div className="flex-shrink-0">
            <AdjustmentsHorizontalIcon className="h-5 w-5 text-blue-400" />
          </div>
          <div className="ml-3">
            <h3 className="text-sm font-medium text-blue-800">Weight Guidelines</h3>
            <div className="mt-2 text-sm text-blue-700">
              <ul className="list-disc list-inside space-y-1">
                <li><strong>0.8-1.0 (High):</strong> Verified expert skills with strong task performance</li>
                <li><strong>0.6-0.8 (Good):</strong> Demonstrated skills with positive task outcomes</li>
                <li><strong>0.4-0.6 (Fair):</strong> Basic skills with average performance</li>
                <li><strong>0.1-0.4 (Low):</strong> Unverified or poor-performing skills</li>
              </ul>
              <p className="mt-2">
                Weights are automatically adjusted based on task completion performance, but can be manually overridden by admins.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
