package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"honey/internal/config"
	"honey/internal/database"
	"honey/internal/logger"
	"honey/internal/models"
)

// WebServer représente le serveur web
type WebServer struct {
	config *config.WebConfig
	logger logger.Logger
	db     *sql.DB
	server *http.Server
}

// New crée une nouvelle instance du serveur web
func New(cfg config.WebConfig, log logger.Logger, db *sql.DB) *WebServer {
	return &WebServer{
		config: &cfg,
		logger: log,
		db:     db,
	}
}

// Start démarre le serveur web
func (w *WebServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// Routes statiques
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(w.config.StaticPath))))

	// Routes API - SSH Honeypot
	mux.HandleFunc("/api/connections", w.handleConnections)
	mux.HandleFunc("/api/commands", w.handleCommands)
	mux.HandleFunc("/api/dangerous-commands", w.handleDangerousCommands)
	mux.HandleFunc("/api/statistics", w.handleStatistics)
	mux.HandleFunc("/api/alerts", w.handleAlerts)

	// Routes API - HTTP Honeypot
	mux.HandleFunc("/api/http/requests", w.handleHTTPRequests)
	mux.HandleFunc("/api/http/attacks", w.handleHTTPAttacks)
	mux.HandleFunc("/api/http/uploads", w.handleHTTPUploads)
	mux.HandleFunc("/api/http/credentials", w.handleHTTPCredentials)
	mux.HandleFunc("/api/http/scanners", w.handleHTTPScanners)
	mux.HandleFunc("/api/http/statistics", w.handleHTTPStatistics)

	// Routes API - FTP Honeypot
	mux.HandleFunc("/api/ftp/connections", w.handleFTPConnections)
	mux.HandleFunc("/api/ftp/commands", w.handleFTPCommands)
	mux.HandleFunc("/api/ftp/statistics", w.handleFTPStatistics)

	// Route principale
	mux.HandleFunc("/", w.handleIndex)

	// Créer le serveur HTTP
	w.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", w.config.Host, w.config.Port),
		Handler: mux,
	}

	// Démarrer le serveur
	go func() {
		<-ctx.Done()
		w.server.Shutdown(context.Background())
	}()

	w.logger.Infof("Web server starting on http://%s:%d", w.config.Host, w.config.Port)
	return w.server.ListenAndServe()
}

// handleIndex gère la page principale
func (w *WebServer) handleIndex(wr http.ResponseWriter, r *http.Request) {
	// Try to serve the dashboard template file
	http.ServeFile(wr, r, w.config.TemplatesPath+"/dashboard.html")
	return

	// Fallback HTML if template not found
	html := `<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Honey SSH Honeypot</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            border-radius: 10px;
            margin-bottom: 20px;
            text-align: center;
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .stat-card {
            background: white;
            padding: 20px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            text-align: center;
        }
        .stat-number {
            font-size: 2em;
            font-weight: bold;
            color: #667eea;
        }
        .stat-label {
            color: #666;
            margin-top: 5px;
        }
        .section {
            background: white;
            padding: 20px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            margin-bottom: 20px;
        }
        .section h2 {
            margin-top: 0;
            color: #333;
            border-bottom: 2px solid #667eea;
            padding-bottom: 10px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 15px;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: #f8f9fa;
            font-weight: 600;
        }
        .status-success {
            color: #28a745;
            font-weight: bold;
        }
        .status-failed {
            color: #dc3545;
            font-weight: bold;
        }
        .refresh-btn {
            background: #667eea;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 5px;
            cursor: pointer;
            margin: 10px 0;
        }
        .refresh-btn:hover {
            background: #5a6fd8;
        }
        .loading {
            text-align: center;
            color: #666;
            font-style: italic;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🍯 Honey SSH Honeypot</h1>
            <p>Surveillance des tentatives d'intrusion en temps réel</p>
        </div>

        <div class="stats-grid" id="statsGrid">
            <div class="stat-card">
                <div class="stat-number" id="totalConnections">-</div>
                <div class="stat-label">Connexions Total</div>
            </div>
            <div class="stat-card">
                <div class="stat-number" id="successfulLogins">-</div>
                <div class="stat-label">Connexions Réussies</div>
            </div>
            <div class="stat-card">
                <div class="stat-number" id="failedLogins">-</div>
                <div class="stat-label">Tentatives Échouées</div>
            </div>
            <div class="stat-card">
                <div class="stat-number" id="uniqueAttackers">-</div>
                <div class="stat-label">Attaquants Uniques</div>
            </div>
        </div>

        <div class="section">
            <h2>📊 Statistiques Détaillées</h2>
            <button class="refresh-btn" onclick="loadStatistics()">Actualiser</button>
            <div id="detailedStats" class="loading">Chargement...</div>
        </div>

        <div class="section">
            <h2>🔗 Connexions Récentes</h2>
            <button class="refresh-btn" onclick="loadConnections()">Actualiser</button>
            <div id="connectionsTable" class="loading">Chargement...</div>
        </div>

        <div class="section">
            <h2>⚡ Commandes Exécutées</h2>
            <button class="refresh-btn" onclick="loadCommands()">Actualiser</button>
            <div id="commandsTable" class="loading">Chargement...</div>
        </div>

        <div class="section">
            <h2>🚨 Alertes</h2>
            <button class="refresh-btn" onclick="loadAlerts()">Actualiser</button>
            <div id="alertsTable" class="loading">Chargement...</div>
        </div>
    </div>

    <script>
        // Charger les données au démarrage
        window.onload = function() {
            loadStatistics();
            loadConnections();
            loadCommands();
            loadAlerts();
            
            // Actualiser automatiquement toutes les 30 secondes
            setInterval(function() {
                loadStatistics();
                loadConnections();
                loadCommands();
                loadAlerts();
            }, 30000);
        };

        function loadStatistics() {
            fetch('/api/statistics')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('totalConnections').textContent = data.total_connections;
                    document.getElementById('successfulLogins').textContent = data.successful_logins;
                    document.getElementById('failedLogins').textContent = data.failed_logins;
                    document.getElementById('uniqueAttackers').textContent = data.unique_attackers;
                    
                    // Statistiques détaillées
                    let detailedHtml = '<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px;">';
                    detailedHtml += '<div><strong>Commandes Total:</strong> ' + data.total_commands + '</div>';
                    detailedHtml += '<div><strong>Attaques 24h:</strong> ' + data.attacks_last_24h + '</div>';
                    detailedHtml += '<div><strong>Attaques 7j:</strong> ' + data.attacks_last_7d + '</div>';
                    detailedHtml += '<div><strong>Attaques 30j:</strong> ' + data.attacks_last_30d + '</div>';
                    detailedHtml += '</div>';
                    
                    if (data.top_countries && data.top_countries.length > 0) {
                        detailedHtml += '<h3>Top Pays</h3><ul>';
                        data.top_countries.slice(0, 5).forEach(country => {
                            detailedHtml += '<li>' + country.country + ': ' + country.count + '</li>';
                        });
                        detailedHtml += '</ul>';
                    }
                    
                    if (data.top_usernames && data.top_usernames.length > 0) {
                        detailedHtml += '<h3>Top Utilisateurs</h3><ul>';
                        data.top_usernames.slice(0, 5).forEach(user => {
                            detailedHtml += '<li>' + user.username + ': ' + user.count + '</li>';
                        });
                        detailedHtml += '</ul>';
                    }
                    
                    document.getElementById('detailedStats').innerHTML = detailedHtml;
                })
                .catch(error => {
                    console.error('Erreur lors du chargement des statistiques:', error);
                    document.getElementById('detailedStats').innerHTML = '<div style="color: red;">Erreur de chargement</div>';
                });
        }

        function loadConnections() {
            fetch('/api/connections?limit=20')
                .then(response => response.json())
                .then(data => {
                    let html = '<table><thead><tr><th>Adresse IP</th><th>Utilisateur</th><th>Mot de Passe</th><th>Statut</th><th>Date</th><th>Durée</th></tr></thead><tbody>';
                    data.forEach(conn => {
                        const status = conn.success ? '<span class="status-success">✓ Réussi</span>' : '<span class="status-failed">✗ Échoué</span>';
                        const date = new Date(conn.connected_at).toLocaleString('fr-FR');
                        const duration = conn.duration ? conn.duration + 's' : '-';
                        html += '<tr>';
                        html += '<td>' + conn.remote_addr + '</td>';
                        html += '<td>' + (conn.username || '-') + '</td>';
                        html += '<td>' + (conn.password || '-') + '</td>';
                        html += '<td>' + status + '</td>';
                        html += '<td>' + date + '</td>';
                        html += '<td>' + duration + '</td>';
                        html += '</tr>';
                    });
                    html += '</tbody></table>';
                    document.getElementById('connectionsTable').innerHTML = html;
                })
                .catch(error => {
                    console.error('Erreur lors du chargement des connexions:', error);
                    document.getElementById('connectionsTable').innerHTML = '<div style="color: red;">Erreur de chargement</div>';
                });
        }

        function loadCommands() {
            fetch('/api/commands?limit=50')
                .then(response => response.json())
                .then(data => {
                    let html = '<table><thead><tr><th>Connexion ID</th><th>Commande</th><th>Date</th></tr></thead><tbody>';
                    data.forEach(cmd => {
                        const date = new Date(cmd.executed_at).toLocaleString('fr-FR');
                        html += '<tr>';
                        html += '<td>' + cmd.connection_id + '</td>';
                        html += '<td><code>' + cmd.command + '</code></td>';
                        html += '<td>' + date + '</td>';
                        html += '</tr>';
                    });
                    html += '</tbody></table>';
                    document.getElementById('commandsTable').innerHTML = html;
                })
                .catch(error => {
                    console.error('Erreur lors du chargement des commandes:', error);
                    document.getElementById('commandsTable').innerHTML = '<div style="color: red;">Erreur de chargement</div>';
                });
        }

        function loadAlerts() {
            fetch('/api/alerts?limit=20')
                .then(response => response.json())
                .then(data => {
                    let html = '<table><thead><tr><th>Type</th><th>Sévérité</th><th>Message</th><th>Adresse IP</th><th>Date</th></tr></thead><tbody>';
                    data.forEach(alert => {
                        const severityClass = alert.severity === 'high' ? 'status-failed' : 
                                            alert.severity === 'medium' ? 'status-warning' : 'status-success';
                        const date = new Date(alert.created_at).toLocaleString('fr-FR');
                        html += '<tr>';
                        html += '<td>' + alert.type + '</td>';
                        html += '<td><span class="' + severityClass + '">' + alert.severity + '</span></td>';
                        html += '<td>' + alert.message + '</td>';
                        html += '<td>' + alert.remote_addr + '</td>';
                        html += '<td>' + date + '</td>';
                        html += '</tr>';
                    });
                    html += '</tbody></table>';
                    document.getElementById('alertsTable').innerHTML = html;
                })
                .catch(error => {
                    console.error('Erreur lors du chargement des alertes:', error);
                    document.getElementById('alertsTable').innerHTML = '<div style="color: red;">Erreur de chargement</div>';
                });
        }
    </script>
</body>
</html>`

	wr.Header().Set("Content-Type", "text/html; charset=utf-8")
	wr.Write([]byte(html))
}

// handleConnections gère les requêtes pour les connexions
func (w *WebServer) handleConnections(wr http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	connections, err := database.GetConnections(limit)
	if err != nil {
		http.Error(wr, "Failed to get connections", http.StatusInternalServerError)
		return
	}

	wr.Header().Set("Content-Type", "application/json")
	json.NewEncoder(wr).Encode(connections)
}

// handleCommands gère les requêtes pour les commandes
func (w *WebServer) handleCommands(wr http.ResponseWriter, r *http.Request) {
	// Jointure avec connections pour récupérer l'IP et le username
	query := `SELECT 
				c.id, 
				c.connection_id, 
				c.command, 
				c.executed_at, 
				c.response,
				COALESCE(cn.remote_addr, '') as remote_addr,
				COALESCE(cn.username, '') as username
			  FROM commands c
			  LEFT JOIN connections cn ON c.connection_id = cn.id
			  ORDER BY c.executed_at DESC LIMIT 100`
	
	rows, err := w.db.Query(query)
	if err != nil {
		w.logger.Errorf("Failed to query commands: %v", err)
		http.Error(wr, "Failed to get commands", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type CommandWithUser struct {
		models.Command
		RemoteAddr string `json:"remote_addr"`
		Username   string `json:"username"`
	}

	var commands []CommandWithUser
	for rows.Next() {
		var id, connectionID int
		var cmdText, response, remoteAddr, username sql.NullString
		var executedAt sql.NullTime
		
		err := rows.Scan(&id, &connectionID, &cmdText, &executedAt, 
						&response, &remoteAddr, &username)
		if err != nil {
			w.logger.Errorf("Failed to scan command: %v", err)
			continue
		}
		
		// Créer la structure CommandWithUser
		cmd := CommandWithUser{
			Command: models.Command{
				ID:           id,
				ConnectionID: connectionID,
				Command:      cmdText.String,
				ExecutedAt:   executedAt.Time,
				Response:     response.String,
			},
			RemoteAddr: remoteAddr.String,
			Username:   username.String,
		}
		
		commands = append(commands, cmd)
	}

	wr.Header().Set("Content-Type", "application/json")
	json.NewEncoder(wr).Encode(commands)
}

// handleDangerousCommands gère les requêtes pour les commandes dangereuses uniquement
func (w *WebServer) handleDangerousCommands(wr http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	// Patterns de commandes dangereuses
	dangerousPatterns := []string{
		"wget", "curl", "nc", "netcat", "/bin/bash -i", "python -c", "perl -e",
		"bash -i", "sh -i", "base64 -d", "history -c", "rm -rf", "dd if=",
		"mkfs", "chmod 777", "passwd", "/dev/sda", "ps aux", "netstat",
		"ifconfig", "who", "last", "cat /etc/passwd", "cat /etc/shadow",
		"find / -name", "sudo",
	}

	// Construire la requête SQL avec des conditions OR pour chaque pattern
	query := `SELECT
				c.id,
				c.connection_id,
				c.command,
				c.executed_at,
				c.response,
				COALESCE(cn.remote_addr, '') as remote_addr,
				COALESCE(cn.username, '') as username
			  FROM commands c
			  LEFT JOIN connections cn ON c.connection_id = cn.id
			  WHERE (`

	conditions := []string{}
	args := []interface{}{}
	for _, pattern := range dangerousPatterns {
		conditions = append(conditions, "LOWER(c.command) LIKE ?")
		args = append(args, "%"+strings.ToLower(pattern)+"%")
	}

	query += strings.Join(conditions, " OR ")
	query += `) ORDER BY c.executed_at DESC LIMIT ?`
	args = append(args, limit)

	rows, err := w.db.Query(query, args...)
	if err != nil {
		w.logger.Errorf("Failed to query dangerous commands: %v", err)
		http.Error(wr, "Failed to get dangerous commands", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type DangerousCommand struct {
		ID           int       `json:"id"`
		ConnectionID int       `json:"connection_id"`
		Command      string    `json:"command"`
		ExecutedAt   time.Time `json:"executed_at"`
		Response     string    `json:"response"`
		RemoteAddr   string    `json:"remote_addr"`
		Username     string    `json:"username"`
		DangerLevel  string    `json:"danger_level"`
		DangerIcon   string    `json:"danger_icon"`
	}

	var commands []DangerousCommand
	for rows.Next() {
		var id, connectionID int
		var cmdText, response, remoteAddr, username sql.NullString
		var executedAt sql.NullTime

		err := rows.Scan(&id, &connectionID, &cmdText, &executedAt,
						&response, &remoteAddr, &username)
		if err != nil {
			w.logger.Errorf("Failed to scan dangerous command: %v", err)
			continue
		}

		// Analyser le niveau de danger
		dangerLevel, dangerIcon := analyzeDangerLevel(cmdText.String)

		cmd := DangerousCommand{
			ID:           id,
			ConnectionID: connectionID,
			Command:      cmdText.String,
			ExecutedAt:   executedAt.Time,
			Response:     response.String,
			RemoteAddr:   remoteAddr.String,
			Username:     username.String,
			DangerLevel:  dangerLevel,
			DangerIcon:   dangerIcon,
		}

		commands = append(commands, cmd)
	}

	wr.Header().Set("Content-Type", "application/json")
	json.NewEncoder(wr).Encode(commands)
}

// analyzeDangerLevel analyse le niveau de danger d'une commande
func analyzeDangerLevel(command string) (string, string) {
	cmd := strings.ToLower(command)

	// Commandes critiques
	if strings.Contains(cmd, "rm -rf") || strings.Contains(cmd, "dd if=") ||
		strings.Contains(cmd, "mkfs") || strings.Contains(cmd, "chmod 777") ||
		strings.Contains(cmd, "passwd") || strings.Contains(cmd, "/dev/sda") {
		return "critical", "🔴"
	}

	// Commandes hautement suspectes
	if strings.Contains(cmd, "wget") || strings.Contains(cmd, "curl") ||
		strings.Contains(cmd, "nc -") || strings.Contains(cmd, "netcat") ||
		strings.Contains(cmd, "/bin/bash -i") || strings.Contains(cmd, "python -c") ||
		strings.Contains(cmd, "perl -e") || strings.Contains(cmd, "bash -i") ||
		strings.Contains(cmd, "sh -i") || strings.Contains(cmd, "base64 -d") ||
		strings.Contains(cmd, "history -c") {
		return "high", "🟠"
	}

	// Commandes moyennement suspectes
	if strings.Contains(cmd, "ps aux") || strings.Contains(cmd, "netstat") ||
		strings.Contains(cmd, "ifconfig") || strings.Contains(cmd, "who") ||
		strings.Contains(cmd, "last") || strings.Contains(cmd, "cat /etc/passwd") ||
		strings.Contains(cmd, "cat /etc/shadow") || strings.Contains(cmd, "find / -name") ||
		strings.Contains(cmd, "sudo ") {
		return "medium", "🟡"
	}

	return "low", "🔵"
}

// handleStatistics gère les requêtes pour les statistiques
func (w *WebServer) handleStatistics(wr http.ResponseWriter, r *http.Request) {
	stats, err := database.GetStatistics()
	if err != nil {
		http.Error(wr, "Failed to get statistics", http.StatusInternalServerError)
		return
	}

	wr.Header().Set("Content-Type", "application/json")
	json.NewEncoder(wr).Encode(stats)
}

// handleAlerts gère les requêtes pour les alertes
func (w *WebServer) handleAlerts(wr http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	query := `SELECT id, type, severity, message, remote_addr, details, created_at, sent, sent_at 
			  FROM alerts ORDER BY created_at DESC LIMIT ?`
	
	rows, err := w.db.Query(query, limit)
	if err != nil {
		http.Error(wr, "Failed to get alerts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var alerts []models.Alert
	for rows.Next() {
		var alert models.Alert
		err := rows.Scan(&alert.ID, &alert.Type, &alert.Severity, &alert.Message, 
			&alert.RemoteAddr, &alert.Details, &alert.CreatedAt, &alert.Sent, &alert.SentAt)
		if err != nil {
			continue
		}
		alerts = append(alerts, alert)
	}

	wr.Header().Set("Content-Type", "application/json")
	json.NewEncoder(wr).Encode(alerts)
}

// handleHTTPRequests gère les requêtes HTTP honeypot
func (w *WebServer) handleHTTPRequests(wr http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	requests, err := database.GetHTTPRequests(limit)
	if err != nil {
		w.logger.Errorf("Failed to get HTTP requests: %v", err)
		http.Error(wr, "Failed to get HTTP requests", http.StatusInternalServerError)
		return
	}

	wr.Header().Set("Content-Type", "application/json")
	json.NewEncoder(wr).Encode(requests)
}

// handleHTTPAttacks gère les attaques HTTP détectées
func (w *WebServer) handleHTTPAttacks(wr http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	attacks, err := database.GetHTTPAttacks(limit)
	if err != nil {
		w.logger.Errorf("Failed to get HTTP attacks: %v", err)
		http.Error(wr, "Failed to get HTTP attacks", http.StatusInternalServerError)
		return
	}

	wr.Header().Set("Content-Type", "application/json")
	json.NewEncoder(wr).Encode(attacks)
}

// handleHTTPUploads gère les fichiers uploadés
func (w *WebServer) handleHTTPUploads(wr http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	query := `SELECT id, request_id, filename, original_filename, size, content_type,
			  md5_hash, sha256_hash, file_path, is_malicious, virustotal_result, timestamp
			  FROM uploaded_files ORDER BY timestamp DESC LIMIT ?`

	rows, err := w.db.Query(query, limit)
	if err != nil {
		w.logger.Errorf("Failed to query uploaded files: %v", err)
		http.Error(wr, "Failed to get uploaded files", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var uploads []models.UploadedFile
	for rows.Next() {
		var upload models.UploadedFile
		var vtResult sql.NullString
		err := rows.Scan(&upload.ID, &upload.RequestID, &upload.Filename, &upload.OriginalFilename,
			&upload.Size, &upload.ContentType, &upload.MD5Hash, &upload.SHA256Hash,
			&upload.FilePath, &upload.IsMalicious, &vtResult, &upload.Timestamp)
		if err != nil {
			w.logger.Errorf("Failed to scan uploaded file: %v", err)
			continue
		}
		upload.VirusTotalResult = vtResult.String
		uploads = append(uploads, upload)
	}

	wr.Header().Set("Content-Type", "application/json")
	json.NewEncoder(wr).Encode(uploads)
}

// handleHTTPCredentials gère les credentials capturés
func (w *WebServer) handleHTTPCredentials(wr http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	query := `SELECT id, request_id, application, username, password, success, timestamp
			  FROM http_credentials ORDER BY timestamp DESC LIMIT ?`

	rows, err := w.db.Query(query, limit)
	if err != nil {
		w.logger.Errorf("Failed to query HTTP credentials: %v", err)
		http.Error(wr, "Failed to get HTTP credentials", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var credentials []models.HTTPCredential
	for rows.Next() {
		var cred models.HTTPCredential
		err := rows.Scan(&cred.ID, &cred.RequestID, &cred.Application, &cred.Username,
			&cred.Password, &cred.Success, &cred.Timestamp)
		if err != nil {
			w.logger.Errorf("Failed to scan HTTP credential: %v", err)
			continue
		}
		credentials = append(credentials, cred)
	}

	wr.Header().Set("Content-Type", "application/json")
	json.NewEncoder(wr).Encode(credentials)
}

// handleHTTPScanners gère les scanners détectés
func (w *WebServer) handleHTTPScanners(wr http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	query := `SELECT id, remote_addr, scanner_type, version, request_count,
			  requests_per_second, detected_at, last_seen
			  FROM scanner_detections ORDER BY last_seen DESC LIMIT ?`

	rows, err := w.db.Query(query, limit)
	if err != nil {
		w.logger.Errorf("Failed to query scanners: %v", err)
		http.Error(wr, "Failed to get scanners", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var scanners []models.ScannerDetection
	for rows.Next() {
		var scanner models.ScannerDetection
		var version sql.NullString
		err := rows.Scan(&scanner.ID, &scanner.RemoteAddr, &scanner.ScannerType, &version,
			&scanner.RequestCount, &scanner.RequestsPerSecond, &scanner.DetectedAt, &scanner.LastSeen)
		if err != nil {
			w.logger.Errorf("Failed to scan scanner: %v", err)
			continue
		}
		scanner.Version = version.String
		scanners = append(scanners, scanner)
	}

	wr.Header().Set("Content-Type", "application/json")
	json.NewEncoder(wr).Encode(scanners)
}

// handleHTTPStatistics gère les statistiques HTTP
func (w *WebServer) handleHTTPStatistics(wr http.ResponseWriter, r *http.Request) {
	stats := models.HTTPStatistics{}

	// Total requests
	w.db.QueryRow("SELECT COUNT(*) FROM http_requests").Scan(&stats.TotalRequests)

	// Total attacks
	w.db.QueryRow("SELECT COUNT(*) FROM http_attacks").Scan(&stats.TotalAttacks)

	// Total uploads
	w.db.QueryRow("SELECT COUNT(*) FROM uploaded_files").Scan(&stats.TotalUploads)

	// Malicious uploads
	w.db.QueryRow("SELECT COUNT(*) FROM uploaded_files WHERE is_malicious = 1").Scan(&stats.MaliciousUploads)

	// Top attack types
	rows, err := w.db.Query(`SELECT attack_type, COUNT(*) as count FROM http_attacks
							 GROUP BY attack_type ORDER BY count DESC LIMIT 10`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ats models.AttackTypeStats
			rows.Scan(&ats.AttackType, &ats.Count)
			stats.TopAttackTypes = append(stats.TopAttackTypes, ats)
		}
	}

	// Top paths
	rows, err = w.db.Query(`SELECT path, COUNT(*) as count FROM http_requests
							GROUP BY path ORDER BY count DESC LIMIT 10`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ps models.PathStats
			rows.Scan(&ps.Path, &ps.Count)
			stats.TopPaths = append(stats.TopPaths, ps)
		}
	}

	// Top user agents
	rows, err = w.db.Query(`SELECT user_agent, COUNT(*) as count FROM http_requests
							WHERE user_agent IS NOT NULL AND user_agent != ''
							GROUP BY user_agent ORDER BY count DESC LIMIT 10`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var uas models.UserAgentStats
			rows.Scan(&uas.UserAgent, &uas.Count)
			stats.TopUserAgents = append(stats.TopUserAgents, uas)
		}
	}

	wr.Header().Set("Content-Type", "application/json")
	json.NewEncoder(wr).Encode(stats)
}

// handleFTPConnections gère les requêtes pour les connexions FTP
func (w *WebServer) handleFTPConnections(wr http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	query := `SELECT id, remote_addr, username, password, authenticated, connected_at,
			  disconnected_at, duration, current_dir, login_attempts, country, city
			  FROM ftp_connections ORDER BY connected_at DESC LIMIT ?`

	rows, err := w.db.Query(query, limit)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var connections []map[string]interface{}
	for rows.Next() {
		var (
			id             int
			remoteAddr     string
			username       sql.NullString
			password       sql.NullString
			authenticated  bool
			connectedAt    time.Time
			disconnectedAt sql.NullTime
			duration       sql.NullInt64
			currentDir     string
			loginAttempts  int
			country        sql.NullString
			city           sql.NullString
		)

		err := rows.Scan(&id, &remoteAddr, &username, &password, &authenticated,
			&connectedAt, &disconnectedAt, &duration, &currentDir, &loginAttempts,
			&country, &city)
		if err != nil {
			continue
		}

		conn := map[string]interface{}{
			"id":             id,
			"remote_addr":    remoteAddr,
			"username":       username.String,
			"password":       password.String,
			"authenticated":  authenticated,
			"connected_at":   connectedAt,
			"current_dir":    currentDir,
			"login_attempts": loginAttempts,
			"country":        country.String,
			"city":           city.String,
		}

		if disconnectedAt.Valid {
			conn["disconnected_at"] = disconnectedAt.Time
		}
		if duration.Valid {
			conn["duration"] = duration.Int64
		}

		connections = append(connections, conn)
	}

	wr.Header().Set("Content-Type", "application/json")
	json.NewEncoder(wr).Encode(connections)
}

// handleFTPCommands gère les requêtes pour les commandes FTP
func (w *WebServer) handleFTPCommands(wr http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	query := `SELECT c.id, c.connection_id, c.command, c.executed_at, c.response,
			  COALESCE(fc.remote_addr, '') as remote_addr,
			  COALESCE(fc.username, '') as username
			  FROM ftp_commands c
			  LEFT JOIN ftp_connections fc ON c.connection_id = fc.id
			  ORDER BY c.executed_at DESC LIMIT ?`

	rows, err := w.db.Query(query, limit)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var commands []map[string]interface{}
	for rows.Next() {
		var (
			id           int
			connectionID int
			command      string
			executedAt   time.Time
			response     sql.NullString
			remoteAddr   string
			username     string
		)

		err := rows.Scan(&id, &connectionID, &command, &executedAt, &response,
			&remoteAddr, &username)
		if err != nil {
			continue
		}

		cmd := map[string]interface{}{
			"id":            id,
			"connection_id": connectionID,
			"command":       command,
			"executed_at":   executedAt,
			"remote_addr":   remoteAddr,
			"username":      username,
		}

		if response.Valid {
			cmd["response"] = response.String
		}

		commands = append(commands, cmd)
	}

	wr.Header().Set("Content-Type", "application/json")
	json.NewEncoder(wr).Encode(commands)
}

// handleFTPStatistics gère les statistiques FTP
func (w *WebServer) handleFTPStatistics(wr http.ResponseWriter, r *http.Request) {
	stats := make(map[string]interface{})

	// Total connexions FTP
	var totalConnections int
	err := w.db.QueryRow("SELECT COUNT(*) FROM ftp_connections").Scan(&totalConnections)
	if err == nil {
		stats["total_connections"] = totalConnections
	}

	// Connexions authentifiées
	var authenticatedLogins int
	err = w.db.QueryRow("SELECT COUNT(*) FROM ftp_connections WHERE authenticated = 1").Scan(&authenticatedLogins)
	if err == nil {
		stats["authenticated_logins"] = authenticatedLogins
	}

	// Tentatives échouées
	var failedLogins int
	err = w.db.QueryRow("SELECT COUNT(*) FROM ftp_connections WHERE authenticated = 0 AND username IS NOT NULL").Scan(&failedLogins)
	if err == nil {
		stats["failed_logins"] = failedLogins
	}

	// Total commandes FTP
	var totalCommands int
	err = w.db.QueryRow("SELECT COUNT(*) FROM ftp_commands").Scan(&totalCommands)
	if err == nil {
		stats["total_commands"] = totalCommands
	}

	// Attaquants uniques
	var uniqueAttackers int
	err = w.db.QueryRow("SELECT COUNT(DISTINCT remote_addr) FROM ftp_connections").Scan(&uniqueAttackers)
	if err == nil {
		stats["unique_attackers"] = uniqueAttackers
	}

	// Connexions dernières 24h
	var last24h int
	err = w.db.QueryRow("SELECT COUNT(*) FROM ftp_connections WHERE connected_at > datetime('now', '-1 day')").Scan(&last24h)
	if err == nil {
		stats["connections_last_24h"] = last24h
	}

	// Top usernames
	rows, err := w.db.Query(`SELECT username, COUNT(*) as count FROM ftp_connections
							 WHERE username IS NOT NULL AND username != ''
							 GROUP BY username ORDER BY count DESC LIMIT 10`)
	if err == nil {
		defer rows.Close()
		var topUsernames []map[string]interface{}
		for rows.Next() {
			var username string
			var count int
			rows.Scan(&username, &count)
			topUsernames = append(topUsernames, map[string]interface{}{
				"username": username,
				"count":    count,
			})
		}
		stats["top_usernames"] = topUsernames
	}

	// Top passwords
	rows, err = w.db.Query(`SELECT password, COUNT(*) as count FROM ftp_connections
							 WHERE password IS NOT NULL AND password != ''
							 GROUP BY password ORDER BY count DESC LIMIT 10`)
	if err == nil {
		defer rows.Close()
		var topPasswords []map[string]interface{}
		for rows.Next() {
			var password string
			var count int
			rows.Scan(&password, &count)
			topPasswords = append(topPasswords, map[string]interface{}{
				"password": password,
				"count":    count,
			})
		}
		stats["top_passwords"] = topPasswords
	}

	// Top commands
	rows, err = w.db.Query(`SELECT command, COUNT(*) as count FROM ftp_commands
							 GROUP BY command ORDER BY count DESC LIMIT 10`)
	if err == nil {
		defer rows.Close()
		var topCommands []map[string]interface{}
		for rows.Next() {
			var command string
			var count int
			rows.Scan(&command, &count)
			topCommands = append(topCommands, map[string]interface{}{
				"command": command,
				"count":   count,
			})
		}
		stats["top_commands"] = topCommands
	}

	// Top pays
	rows, err = w.db.Query(`SELECT country, COUNT(*) as count FROM ftp_connections
							 WHERE country IS NOT NULL AND country != ''
							 GROUP BY country ORDER BY count DESC LIMIT 10`)
	if err == nil {
		defer rows.Close()
		var topCountries []map[string]interface{}
		for rows.Next() {
			var country string
			var count int
			rows.Scan(&country, &count)
			topCountries = append(topCountries, map[string]interface{}{
				"country": country,
				"count":   count,
			})
		}
		stats["top_countries"] = topCountries
	}

	wr.Header().Set("Content-Type", "application/json")
	json.NewEncoder(wr).Encode(stats)
}
