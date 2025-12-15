package database

import (
	"database/sql"
	"fmt"
	"honey/internal/config"
	"honey/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// Initialize initialise la base de données
func Initialize(cfg config.DatabaseConfig) (*sql.DB, error) {
	var err error
	
	switch cfg.Type {
	case "sqlite":
		db, err = sql.Open("sqlite3", cfg.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	// Test de connexion
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Création des tables
	if err := createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

// createTables crée les tables nécessaires
func createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS connections (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			remote_addr TEXT NOT NULL,
			username TEXT,
			password TEXT,
			success BOOLEAN,
			connected_at DATETIME,
			disconnected_at DATETIME,
			duration INTEGER,
			user_agent TEXT,
			country TEXT,
			city TEXT,
			isp TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS commands (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			connection_id INTEGER NOT NULL,
			command TEXT NOT NULL,
			executed_at DATETIME,
			response TEXT,
			FOREIGN KEY (connection_id) REFERENCES connections (id)
		)`,
		`CREATE TABLE IF NOT EXISTS attack_patterns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			remote_addr TEXT NOT NULL,
			pattern_type TEXT,
			attempts INTEGER,
			time_window INTEGER,
			detected_at DATETIME,
			blocked BOOLEAN
		)`,
		`CREATE TABLE IF NOT EXISTS alerts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT,
			severity TEXT,
			message TEXT,
			remote_addr TEXT,
			details TEXT,
			created_at DATETIME,
			sent BOOLEAN,
			sent_at DATETIME
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

// Close ferme la connexion à la base de données
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// SaveConnection sauvegarde une connexion
func SaveConnection(conn *models.Connection) error {
	query := `INSERT INTO connections 
		(remote_addr, username, password, success, connected_at, disconnected_at, duration, user_agent, country, city, isp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err := db.Exec(query,
		conn.RemoteAddr, conn.Username, conn.Password, conn.Success,
		conn.ConnectedAt, conn.DisconnectedAt, conn.Duration,
		conn.UserAgent, conn.Country, conn.City, conn.ISP)
	
	return err
}

// SaveCommand sauvegarde une commande
func SaveCommand(cmd *models.Command) error {
	query := `INSERT INTO commands (connection_id, command, executed_at, response)
		VALUES (?, ?, ?, ?)`
	
	_, err := db.Exec(query, cmd.ConnectionID, cmd.Command, cmd.ExecutedAt, cmd.Response)
	return err
}

// SaveAttackPattern sauvegarde un pattern d'attaque
func SaveAttackPattern(pattern *models.AttackPattern) error {
	query := `INSERT INTO attack_patterns 
		(remote_addr, pattern_type, attempts, time_window, detected_at, blocked)
		VALUES (?, ?, ?, ?, ?, ?)`
	
	_, err := db.Exec(query,
		pattern.RemoteAddr, pattern.PatternType, pattern.Attempts,
		pattern.TimeWindow, pattern.DetectedAt, pattern.Blocked)
	
	return err
}

// SaveAlert sauvegarde une alerte
func SaveAlert(alert *models.Alert) error {
	query := `INSERT INTO alerts 
		(type, severity, message, remote_addr, details, created_at, sent, sent_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err := db.Exec(query,
		alert.Type, alert.Severity, alert.Message, alert.RemoteAddr,
		alert.Details, alert.CreatedAt, alert.Sent, alert.SentAt)
	
	return err
}

// GetConnections récupère les connexions récentes
func GetConnections(limit int) ([]models.Connection, error) {
	query := `SELECT id, remote_addr, username, password, success, connected_at, 
		disconnected_at, duration, user_agent, country, city, isp
		FROM connections ORDER BY connected_at DESC LIMIT ?`
	
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []models.Connection
	for rows.Next() {
		var conn models.Connection
		err := rows.Scan(&conn.ID, &conn.RemoteAddr, &conn.Username, &conn.Password,
			&conn.Success, &conn.ConnectedAt, &conn.DisconnectedAt, &conn.Duration,
			&conn.UserAgent, &conn.Country, &conn.City, &conn.ISP)
		if err != nil {
			return nil, err
		}
		connections = append(connections, conn)
	}

	return connections, nil
}

// GetCommands récupère les commandes d'une connexion
func GetCommands(connectionID int) ([]models.Command, error) {
	query := `SELECT id, connection_id, command, executed_at, response
		FROM commands WHERE connection_id = ? ORDER BY executed_at`
	
	rows, err := db.Query(query, connectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []models.Command
	for rows.Next() {
		var cmd models.Command
		err := rows.Scan(&cmd.ID, &cmd.ConnectionID, &cmd.Command, &cmd.ExecutedAt, &cmd.Response)
		if err != nil {
			return nil, err
		}
		commands = append(commands, cmd)
	}

	return commands, nil
}

// GetStatistics récupère les statistiques
func GetStatistics() (*models.Statistics, error) {
	stats := &models.Statistics{}

	// Total des connexions
	err := db.QueryRow("SELECT COUNT(*) FROM connections").Scan(&stats.TotalConnections)
	if err != nil {
		return nil, err
	}

	// Connexions réussies
	err = db.QueryRow("SELECT COUNT(*) FROM connections WHERE success = 1").Scan(&stats.SuccessfulLogins)
	if err != nil {
		return nil, err
	}

	// Connexions échouées
	err = db.QueryRow("SELECT COUNT(*) FROM connections WHERE success = 0").Scan(&stats.FailedLogins)
	if err != nil {
		return nil, err
	}

	// Total des commandes
	err = db.QueryRow("SELECT COUNT(*) FROM commands").Scan(&stats.TotalCommands)
	if err != nil {
		return nil, err
	}

	// Attaquants uniques
	err = db.QueryRow("SELECT COUNT(DISTINCT remote_addr) FROM connections").Scan(&stats.UniqueAttackers)
	if err != nil {
		return nil, err
	}

	// Attaques des dernières 24h
	query := "SELECT COUNT(*) FROM connections WHERE connected_at > datetime('now', '-1 day')"
	err = db.QueryRow(query).Scan(&stats.AttacksLast24h)
	if err != nil {
		return nil, err
	}

	// Attaques des 7 derniers jours
	query = "SELECT COUNT(*) FROM connections WHERE connected_at > datetime('now', '-7 days')"
	err = db.QueryRow(query).Scan(&stats.AttacksLast7d)
	if err != nil {
		return nil, err
	}

	// Attaques des 30 derniers jours
	query = "SELECT COUNT(*) FROM connections WHERE connected_at > datetime('now', '-30 days')"
	err = db.QueryRow(query).Scan(&stats.AttacksLast30d)
	if err != nil {
		return nil, err
	}

	// Top pays
	rows, err := db.Query(`
		SELECT country, COUNT(*) as count 
		FROM connections 
		WHERE country IS NOT NULL AND country != ''
		GROUP BY country 
		ORDER BY count DESC 
		LIMIT 10`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cs models.CountryStats
			rows.Scan(&cs.Country, &cs.Count)
			stats.TopCountries = append(stats.TopCountries, cs)
		}
	}

	// Top usernames
	rows, err = db.Query(`
		SELECT username, COUNT(*) as count 
		FROM connections 
		WHERE username IS NOT NULL AND username != ''
		GROUP BY username 
		ORDER BY count DESC 
		LIMIT 10`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var us models.UsernameStats
			rows.Scan(&us.Username, &us.Count)
			stats.TopUsernames = append(stats.TopUsernames, us)
		}
	}

	// Top passwords
	rows, err = db.Query(`
		SELECT password, COUNT(*) as count 
		FROM connections 
		WHERE password IS NOT NULL AND password != ''
		GROUP BY password 
		ORDER BY count DESC 
		LIMIT 10`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ps models.PasswordStats
			rows.Scan(&ps.Password, &ps.Count)
			stats.TopPasswords = append(stats.TopPasswords, ps)
		}
	}

	// Top commandes
	rows, err = db.Query(`
		SELECT command, COUNT(*) as count 
		FROM commands 
		GROUP BY command 
		ORDER BY count DESC 
		LIMIT 10`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cs models.CommandStats
			rows.Scan(&cs.Command, &cs.Count)
			stats.TopCommands = append(stats.TopCommands, cs)
		}
	}

	return stats, nil
}

// CheckBruteForce vérifie s'il y a une attaque par force brute
func CheckBruteForce(remoteAddr string, timeWindow int, threshold int) (bool, int, error) {
	query := `
		SELECT COUNT(*) 
		FROM connections 
		WHERE remote_addr = ? 
		AND connected_at > datetime('now', '-? seconds')
		AND success = 0`
	
	var attempts int
	err := db.QueryRow(query, remoteAddr, timeWindow).Scan(&attempts)
	if err != nil {
		return false, 0, err
	}

	return attempts >= threshold, attempts, nil
}

// CountFailedConnectionsInTimeWindow compte les connexions échouées dans une fenêtre de temps
func CountFailedConnectionsInTimeWindow(remoteAddr string, timeWindowSeconds int) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM connections 
		WHERE remote_addr = ? 
		AND connected_at > datetime('now', '-' || ? || ' seconds')
		AND success = 0`
	
	var count int
	err := db.QueryRow(query, remoteAddr, timeWindowSeconds).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// UpdateAlert met à jour une alerte
func UpdateAlert(alert *models.Alert) error {
	query := `UPDATE alerts 
			  SET sent = ?, sent_at = ? 
			  WHERE id = ?`
	
	_, err := db.Exec(query, alert.Sent, alert.SentAt, alert.ID)
	return err
}
