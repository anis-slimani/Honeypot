package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"honey/internal/config"
	"honey/internal/database"
	"honey/internal/honeypot"
	"honey/internal/logger"
	httpservice "honey/internal/services/http"
	ftpservice "honey/internal/services/ftp"
	"honey/internal/web"
)

func main() {
	// Analyse des arguments de ligne de commande
	configPath := flag.String("config", "config.yaml", "Chemin vers le fichier de configuration")
	flag.Parse()

	// Chargement de la configuration YAML
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Échec du chargement de la configuration: %v", err)
	}

	// Initialisation du système de logging
	logger, err := logger.New(cfg.Logging)
	if err != nil {
		log.Fatalf("Échec de l'initialisation du logger: %v", err)
	}
	defer logger.Close()

	logger.Info("Démarrage du Honey SSH Honeypot...")

	// Initialisation de la base de données SQLite
	db, err := database.Initialize(cfg.Database)
	if err != nil {
		logger.Fatalf("Échec de l'initialisation de la base de données: %v", err)
	}
	defer database.Close()

	logger.Info("Base de données initialisée avec succès")

	// Création du contexte pour l'arrêt gracieux des services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialisation du gestionnaire d'alertes email
	alertManager := honeypot.NewAlertManager(cfg, logger)

	// Initialisation du serveur SSH honeypot
	honeypotServer := honeypot.New(cfg, logger, db)

	// Démarrage du serveur SSH en arrière-plan
	go func() {
		if err := honeypotServer.Start(ctx); err != nil {
			logger.Errorf("Erreur du serveur honeypot: %v", err)
		}
	}()

	// Démarrage du honeypot HTTP si activé
	if cfg.HTTP.Enabled {
		httpServer := httpservice.New(&cfg.HTTP, logger, db)
		httpServer.SetAlertManager(alertManager)
		go func() {
			if err := httpServer.Start(ctx); err != nil {
				logger.Errorf("Erreur du honeypot HTTP: %v", err)
			}
		}()
		if cfg.HTTP.TLS.Enabled {
			logger.Infof("Honeypot HTTP démarré sur %s:%d (HTTPS sur le port %d)", cfg.HTTP.Host, cfg.HTTP.Port, cfg.HTTP.TLS.Port)
		} else {
			logger.Infof("Honeypot HTTP démarré sur %s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
		}
	}

	// Démarrage du honeypot FTP si activé
	if cfg.FTP.Enabled {
		ftpServer := ftpservice.New(&cfg.FTP, logger, db)
		ftpServer.SetAlertManager(alertManager)
		go func() {
			if err := ftpServer.Start(ctx); err != nil {
				logger.Errorf("Erreur du honeypot FTP: %v", err)
			}
		}()
		logger.Infof("Honeypot FTP démarré sur %s:%d", cfg.FTP.Host, cfg.FTP.Port)
	}

	// Démarrage de l'interface web pour la visualisation
	if cfg.Web.Enabled {
		webServer := web.New(cfg.Web, logger, db)
		go func() {
			if err := webServer.Start(ctx); err != nil {
				logger.Errorf("Erreur du serveur web: %v", err)
			}
		}()
		logger.Infof("Interface web démarrée sur http://%s:%d", cfg.Web.Host, cfg.Web.Port)
	}

	logger.Infof("Serveur honeypot SSH démarré sur %s:%d", cfg.Server.Host, cfg.Server.Port)

	// Compter les services actifs
	servicesActifs := 1 // SSH toujours actif
	if cfg.HTTP.Enabled {
		servicesActifs++
	}
	if cfg.FTP.Enabled {
		servicesActifs++
	}
	if cfg.Web.Enabled {
		servicesActifs++
	}

	if servicesActifs > 1 {
		logger.Infof("Total de %d services actifs - Appuyez sur Ctrl+C pour arrêter tous les services", servicesActifs)
	} else {
		logger.Info("Appuyez sur Ctrl+C pour arrêter le serveur")
	}

	// Attendre le signal d'interruption
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Arrêt du honeypot...")
	cancel()

	// Donner du temps aux services pour s'arrêter gracieusement
	time.Sleep(2 * time.Second)
	logger.Info("Honeypot arrêté")
}
