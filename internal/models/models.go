package models

import (
	"time"
)

// Connection représente une connexion SSH
type Connection struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	RemoteAddr   string    `json:"remote_addr" gorm:"not null"`
	Username     string    `json:"username"`
	Password     string    `json:"password"`
	Success      bool      `json:"success"`
	ConnectedAt  time.Time `json:"connected_at"`
	DisconnectedAt *time.Time `json:"disconnected_at,omitempty"`
	Duration     int64     `json:"duration"` // en secondes
	UserAgent    string    `json:"user_agent,omitempty"`
	Country      string    `json:"country,omitempty"`
	City         string    `json:"city,omitempty"`
	ISP          string    `json:"isp,omitempty"`
}

// Command représente une commande exécutée dans le shell factice
type Command struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	ConnectionID int       `json:"connection_id" gorm:"not null"`
	Command      string    `json:"command" gorm:"not null"`
	ExecutedAt   time.Time `json:"executed_at"`
	Response     string    `json:"response,omitempty"`
}

// AttackPattern représente un pattern d'attaque détecté
type AttackPattern struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	RemoteAddr  string    `json:"remote_addr" gorm:"not null"`
	PatternType string    `json:"pattern_type"` // "brute_force", "dictionary", "credential_stuffing"
	Attempts    int       `json:"attempts"`
	TimeWindow  int       `json:"time_window"` // en secondes
	DetectedAt  time.Time `json:"detected_at"`
	Blocked     bool      `json:"blocked"`
}

// Alert représente une alerte générée
type Alert struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Type        string    `json:"type"` // "high_risk", "brute_force", "suspicious_activity"
	Severity    string    `json:"severity"` // "low", "medium", "high", "critical"
	Message     string    `json:"message"`
	RemoteAddr  string    `json:"remote_addr"`
	Details     string    `json:"details,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Sent        bool      `json:"sent"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
}

// Statistics représente les statistiques du honeypot
type Statistics struct {
	TotalConnections    int     `json:"total_connections"`
	SuccessfulLogins    int     `json:"successful_logins"`
	FailedLogins        int     `json:"failed_logins"`
	TotalCommands       int     `json:"total_commands"`
	UniqueAttackers     int     `json:"unique_attackers"`
	TopCountries        []CountryStats `json:"top_countries"`
	TopUsernames        []UsernameStats `json:"top_usernames"`
	TopPasswords        []PasswordStats `json:"top_passwords"`
	TopCommands         []CommandStats `json:"top_commands"`
	AttacksLast24h      int     `json:"attacks_last_24h"`
	AttacksLast7d       int     `json:"attacks_last_7d"`
	AttacksLast30d      int     `json:"attacks_last_30d"`
}

// CountryStats statistiques par pays
type CountryStats struct {
	Country string `json:"country"`
	Count   int    `json:"count"`
}

// UsernameStats statistiques par nom d'utilisateur
type UsernameStats struct {
	Username string `json:"username"`
	Count    int    `json:"count"`
}

// PasswordStats statistiques par mot de passe
type PasswordStats struct {
	Password string `json:"password"`
	Count    int    `json:"count"`
}

// CommandStats statistiques par commande
type CommandStats struct {
	Command string `json:"command"`
	Count   int    `json:"count"`
}

// GeolocationData données de géolocalisation
type GeolocationData struct {
	IP      string `json:"ip"`
	Country string `json:"country"`
	City    string `json:"city"`
	ISP     string `json:"isp"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}
