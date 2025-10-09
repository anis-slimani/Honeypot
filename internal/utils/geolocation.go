package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GeolocationService interface pour la géolocalisation
type GeolocationService interface {
	GetLocation(ip string) (*GeolocationData, error)
}

// GeolocationData représente les données de géolocalisation
type GeolocationData struct {
	IP      string  `json:"ip"`
	Country string  `json:"country"`
	City    string  `json:"city"`
	ISP     string  `json:"isp"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

// IPAPIService service de géolocalisation utilisant ip-api.com
type IPAPIService struct {
	client *http.Client
}

// NewGeolocationService crée un nouveau service de géolocalisation
func NewGeolocationService() GeolocationService {
	return &IPAPIService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetLocation récupère la localisation d'une IP
func (s *IPAPIService) GetLocation(ip string) (*GeolocationData, error) {
	// Vérifier si c'est une IP locale
	if s.isLocalIP(ip) {
		return &GeolocationData{
			IP:      ip,
			Country: "Local",
			City:    "Local Network",
			ISP:     "Local Network",
			Lat:     0,
			Lon:     0,
		}, nil
	}

	// URL de l'API ip-api.com (gratuite, sans clé API)
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get geolocation data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geolocation API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parser la réponse JSON
	var apiResponse struct {
		Status      string  `json:"status"`
		Country     string  `json:"country"`
		CountryCode string  `json:"countryCode"`
		Region      string  `json:"region"`
		RegionName  string  `json:"regionName"`
		City        string  `json:"city"`
		Zip         string  `json:"zip"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
		Timezone    string  `json:"timezone"`
		ISP         string  `json:"isp"`
		Org         string  `json:"org"`
		AS          string  `json:"as"`
		Query       string  `json:"query"`
		Message     string  `json:"message"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if apiResponse.Status != "success" {
		return nil, fmt.Errorf("geolocation API error: %s", apiResponse.Message)
	}

	return &GeolocationData{
		IP:      apiResponse.Query,
		Country: apiResponse.Country,
		City:    apiResponse.City,
		ISP:     apiResponse.ISP,
		Lat:     apiResponse.Lat,
		Lon:     apiResponse.Lon,
	}, nil
}

// isLocalIP vérifie si une IP est locale
func (s *IPAPIService) isLocalIP(ip string) bool {
	// IPs locales communes
	localIPs := []string{
		"127.0.0.1",
		"::1",
		"localhost",
	}

	for _, localIP := range localIPs {
		if ip == localIP {
			return true
		}
	}

	// Vérifier les plages d'IPs privées
	// 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16
	if s.isPrivateIP(ip) {
		return true
	}

	return false
}

// isPrivateIP vérifie si une IP est dans une plage privée
func (s *IPAPIService) isPrivateIP(ip string) bool {
	// Cette fonction est simplifiée
	// Dans une vraie implémentation, on parserait l'IP et on vérifierait les plages
	return false
}

// MockGeolocationService service de géolocalisation factice pour les tests
type MockGeolocationService struct{}

// NewMockGeolocationService crée un service de géolocalisation factice
func NewMockGeolocationService() GeolocationService {
	return &MockGeolocationService{}
}

// GetLocation retourne des données factices
func (m *MockGeolocationService) GetLocation(ip string) (*GeolocationData, error) {
	// Données factices pour les tests
	return &GeolocationData{
		IP:      ip,
		Country: "Test Country",
		City:    "Test City",
		ISP:     "Test ISP",
		Lat:     48.8566,
		Lon:     2.3522,
	}, nil
}
