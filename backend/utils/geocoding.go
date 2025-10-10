package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// GeocodingService handles geocoding operations
type GeocodingService struct {
	NominatimBaseURL string
	HTTPClient       *http.Client
}

// GeocodingResult represents a geocoding result from Nominatim
type GeocodingResult struct {
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
	PlaceID     int64  `json:"place_id"`
}

// NewGeocodingService creates a new geocoding service
func NewGeocodingService(baseURL string) *GeocodingService {
	return &GeocodingService{
		NominatimBaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GeocodeAddress converts an address to latitude and longitude coordinates
func (s *GeocodingService) GeocodeAddress(address string) (lat, lng float64, displayName string, err error) {
	if address == "" {
		return 0, 0, "", fmt.Errorf("address cannot be empty")
	}

	// Prepare the request URL
	requestURL := fmt.Sprintf("%s/search?format=json&q=%s&limit=1&addressdetails=1",
		s.NominatimBaseURL, url.QueryEscape(address))

	// Make the request
	resp, err := s.HTTPClient.Get(requestURL)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to make geocoding request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return 0, 0, "", fmt.Errorf("geocoding service returned status %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var results []GeocodingResult
	if err := json.Unmarshal(body, &results); err != nil {
		return 0, 0, "", fmt.Errorf("failed to parse geocoding response: %w", err)
	}

	// Check if we got any results
	if len(results) == 0 {
		return 0, 0, "", fmt.Errorf("no geocoding results found for address: %s", address)
	}

	result := results[0]

	// Parse latitude and longitude
	lat, err = strconv.ParseFloat(result.Lat, 64)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to parse latitude: %w", err)
	}

	lng, err = strconv.ParseFloat(result.Lon, 64)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to parse longitude: %w", err)
	}

	return lat, lng, result.DisplayName, nil
}

// ReverseGeocode converts latitude and longitude coordinates to an address
func (s *GeocodingService) ReverseGeocode(lat, lng float64) (address string, err error) {
	// Prepare the request URL
	requestURL := fmt.Sprintf("%s/reverse?format=json&lat=%f&lon=%f&addressdetails=1",
		s.NominatimBaseURL, lat, lng)

	// Make the request
	resp, err := s.HTTPClient.Get(requestURL)
	if err != nil {
		return "", fmt.Errorf("failed to make reverse geocoding request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("reverse geocoding service returned status %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var result struct {
		DisplayName string `json:"display_name"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse reverse geocoding response: %w", err)
	}

	return result.DisplayName, nil
}

// CalculateDistance calculates the distance between two points using the Haversine formula
// Returns distance in kilometers
func CalculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371 // Earth's radius in kilometers

	// Convert degrees to radians
	lat1Rad := lat1 * 3.14159265359 / 180
	lng1Rad := lng1 * 3.14159265359 / 180
	lat2Rad := lat2 * 3.14159265359 / 180
	lng2Rad := lng2 * 3.14159265359 / 180

	// Haversine formula
	dlat := lat2Rad - lat1Rad
	dlng := lng2Rad - lng1Rad

	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dlng/2)*math.Sin(dlng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
