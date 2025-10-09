package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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

	// Routes API
	mux.HandleFunc("/api/connections", w.handleConnections)
	mux.HandleFunc("/api/commands", w.handleCommands)
	mux.HandleFunc("/api/statistics", w.handleStatistics)
	mux.HandleFunc("/api/alerts", w.handleAlerts)

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
	// Pour simplifier, on récupère toutes les commandes
	// Dans une vraie implémentation, on aurait une fonction GetCommands avec limite
	query := `SELECT id, connection_id, command, executed_at, response 
			  FROM commands ORDER BY executed_at DESC LIMIT 50`
	
	rows, err := w.db.Query(query)
	if err != nil {
		http.Error(wr, "Failed to get commands", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var commands []models.Command
	for rows.Next() {
		var cmd models.Command
		err := rows.Scan(&cmd.ID, &cmd.ConnectionID, &cmd.Command, &cmd.ExecutedAt, &cmd.Response)
		if err != nil {
			continue
		}
		commands = append(commands, cmd)
	}

	wr.Header().Set("Content-Type", "application/json")
	json.NewEncoder(wr).Encode(commands)
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
