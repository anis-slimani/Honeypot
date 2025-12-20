package honeypot

import (
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
	"honey/internal/config"
	"honey/internal/database"
	"honey/internal/logger"
	"honey/internal/models"
)

// FakeSession représente une session SSH factice
type FakeSession struct {
	channel      ssh.Channel
	requests     <-chan *ssh.Request
	config       *config.Config
	logger       logger.Logger
	username     string
	remoteAddr   string
	connectionID int
	currentDir   string // Répertoire courant simulé
	alertManager *AlertManager
}

// handle gère une session SSH factice
func (s *FakeSession) handle() {
	defer s.channel.Close()

	// Initialiser le répertoire courant
	s.currentDir = "/home/user"

	// Enregistrer la connexion réussie
	connection := &models.Connection{
		RemoteAddr:  s.remoteAddr,
		Username:    s.username,
		Success:     true,
		ConnectedAt: time.Now(),
	}

	connID, err := database.SaveConnection(connection)
	if err != nil {
		s.logger.Errorf("Failed to save successful connection: %v", err)
	} else {
		s.connectionID = connID
	}

	// Traiter les requêtes SSH en arrière-plan
	go s.handleRequests()

	// Démarrer le shell interactif immédiatement
	s.runInteractiveShell()
}

// handleRequests traite les requêtes SSH (pty-req, shell, window-change, etc.)
func (s *FakeSession) handleRequests() {
	for req := range s.requests {
		s.logger.Debugf("SSH request '%s' from %s (WantReply=%v)", req.Type, s.remoteAddr, req.WantReply)
		
		switch req.Type {
		case "shell", "pty-req", "window-change", "env":
			// Accepter toutes ces requêtes
			if req.WantReply {
				req.Reply(true, nil)
			}
		case "exec":
			// Commande non-interactive : ssh host "cmd"
			s.handleExecRequest(req)
			return // Terminer après exec
		default:
			// Refuser les autres requêtes
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}
}

// handleExecRequest gère les commandes exec (non-interactives)
func (s *FakeSession) handleExecRequest(req *ssh.Request) {
	var payload struct {
		Command string `ssh:"string"`
	}
	
	if err := ssh.Unmarshal(req.Payload, &payload); err != nil {
		if req.WantReply {
			req.Reply(false, nil)
		}
		return
	}

	command := strings.TrimSpace(payload.Command)
	s.logger.Infof("Exec command from %s: %s", s.remoteAddr, command)

	// Enregistrer la commande
	cmd := &models.Command{
		ConnectionID: s.connectionID,
		Command:      command,
		ExecutedAt:   time.Now(),
	}
	if err := database.SaveCommand(cmd); err != nil {
		s.logger.Errorf("Failed to save command: %v", err)
	}

	// Exécuter et envoyer la réponse
	response := s.executeFakeCommand(command)
	if response != "" {
		s.channel.Write([]byte(response))
	}

	if req.WantReply {
		req.Reply(true, nil)
	}

	// Fermer le canal
	s.channel.CloseWrite()
}

// runInteractiveShell démarre un shell interactif en utilisant golang.org/x/term
func (s *FakeSession) runInteractiveShell() {
	// Créer un terminal virtuel avec golang.org/x/term
	terminal := term.NewTerminal(s.channel, s.config.Shell.Prompt)

	// Message de bienvenue
	welcome := fmt.Sprintf("%s\nLast login: %s from %s\n",
		s.config.Shell.WelcomeMessage,
		time.Now().Format("Mon Jan 2 15:04:05 2006"),
		s.remoteAddr)
	
	terminal.Write([]byte(welcome))

	s.logger.Infof("Interactive shell started for %s", s.remoteAddr)

	// Boucle de lecture des commandes
	for {
		// ReadLine() bloque jusqu'à ce que l'utilisateur appuie sur Entrée
		// Il gère automatiquement l'écho, backspace, Ctrl+C, etc.
		line, err := terminal.ReadLine()
		
		if err != nil {
			if err == io.EOF {
				s.logger.Infof("Client disconnected: %s", s.remoteAddr)
			} else {
				s.logger.Errorf("Error reading line from %s: %v", s.remoteAddr, err)
			}
			break
		}

		// Nettoyer la commande
		command := strings.TrimSpace(line)

		// Ignorer les lignes vides
		if command == "" {
			continue
		}

		s.logger.Infof("Command from %s: %s", s.remoteAddr, command)

		// Gérer les commandes de sortie
		if command == "exit" || command == "logout" {
			terminal.Write([]byte("Goodbye!\n"))
			s.logger.Infof("User %s logged out", s.remoteAddr)
			break
		}

		// Enregistrer la commande dans la base de données
		cmd := &models.Command{
			ConnectionID: s.connectionID,
			Command:      command,
			ExecutedAt:   time.Now(),
		}
		if err := database.SaveCommand(cmd); err != nil {
			s.logger.Errorf("Failed to save command: %v", err)
		}

		// Vérifier si c'est une commande dangereuse et déclencher une alerte
		if s.alertManager != nil {
			go s.alertManager.OnDangerousCommand(s.remoteAddr, s.username, command)
		}

		// Exécuter la commande factice
		response := s.executeFakeCommand(command)
		if response != "" {
			terminal.Write([]byte(response))
		}

		// Mettre à jour le prompt si cd a changé le répertoire
		parts := strings.Fields(command)
		if len(parts) > 0 && parts[0] == "cd" {
			// Le répertoire a changé, mettre à jour le prompt
			newPrompt := fmt.Sprintf("user@Tech.fr:%s$ ", s.currentDir)
			terminal.SetPrompt(newPrompt)
		}
	}

	s.logger.Infof("Interactive shell ended for %s", s.remoteAddr)
}

// executeFakeCommand exécute une commande factice
func (s *FakeSession) executeFakeCommand(command string) string {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return ""
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	// Commandes de base
	case "ls":
		return s.fakeLs(args)
	case "pwd":
		return s.fakePwd()
	case "cd":
		return s.fakeCd(args)
	case "whoami":
		return s.fakeWhoami()
	case "echo":
		return s.fakeEcho(args)
	case "cat":
		return s.fakeCat(args)
	case "grep":
		return s.fakeGrep(args)
	case "find":
		return s.fakeFind(args)
	case "history":
		return s.fakeHistory()
	
	// Gestion des fichiers
	case "mkdir":
		return s.fakeMkdir(args)
	case "touch":
		return s.fakeTouch(args)
	case "rm":
		return s.fakeRm(args)
	case "cp":
		return s.fakeCp(args)
	case "mv":
		return s.fakeMv(args)
	case "chmod":
		return s.fakeChmod(args)
	case "chown":
		return s.fakeChown(args)
	case "ln":
		return s.fakeLn(args)
	case "tar":
		return s.fakeTar(args)
	case "zip":
		return s.fakeZip(args)
	case "unzip":
		return s.fakeUnzip(args)
	case "gzip":
		return s.fakeGzip(args)
	case "gunzip":
		return s.fakeGunzip(args)
	
	// Commandes système critiques
	case "sudo":
		return s.fakeSudo(args)
	case "su":
		return s.fakeSu(args)
	case "passwd":
		return s.fakePasswd(args)
	case "useradd", "adduser":
		return s.fakeUseradd(args)
	case "userdel":
		return s.fakeUserdel(args)
	case "usermod":
		return s.fakeUsermod(args)
	case "groupadd":
		return s.fakeGroupadd(args)
	case "groupdel":
		return s.fakeGroupdel(args)
	
	// Processus et services
	case "ps":
		return s.fakePs()
	case "top":
		return s.fakeTop()
	case "htop":
		return s.fakeHtop()
	case "kill":
		return s.fakeKill(args)
	case "killall":
		return s.fakeKillall(args)
	case "pkill":
		return s.fakePkill(args)
	case "systemctl":
		return s.fakeSystemctl(args)
	case "service":
		return s.fakeService(args)
	case "reboot":
		return s.fakeReboot()
	case "shutdown":
		return s.fakeShutdown(args)
	case "init":
		return s.fakeInit(args)
	
	// Réseau
	case "netstat":
		return s.fakeNetstat()
	case "ifconfig", "ip":
		return s.fakeIfconfig()
	case "ping":
		return s.fakePing(args)
	case "traceroute":
		return s.fakeTraceroute(args)
	case "nslookup":
		return s.fakeNslookup(args)
	case "dig":
		return s.fakeDig(args)
	case "host":
		return s.fakeHost(args)
	case "route":
		return s.fakeRoute()
	case "iptables":
		return s.fakeIptables(args)
	case "ss":
		return s.fakeSs()
	case "nc", "netcat":
		return s.fakeNetcat(args)
	case "telnet":
		return s.fakeTelnet(args)
	case "ssh":
		return s.fakeSsh(args)
	case "scp":
		return s.fakeScp(args)
	case "rsync":
		return s.fakeRsync(args)
	
	// Téléchargement
	case "wget":
		return s.fakeWget(args)
	case "curl":
		return s.fakeCurl(args)
	case "ftp":
		return s.fakeFtp(args)
	
	// Informations système
	case "uname":
		return s.fakeUname()
	case "id":
		return s.fakeId()
	case "groups":
		return s.fakeGroups()
	case "uptime":
		return s.fakeUptime()
	case "df":
		return s.fakeDf()
	case "free":
		return s.fakeFree()
	case "hostname":
		return s.fakeHostname()
	case "hostnamectl":
		return s.fakeHostnamectl()
	case "lsb_release":
		return s.fakeLsbRelease()
	case "w", "who":
		return s.fakeWho()
	case "last":
		return s.fakeLast()
	case "lastlog":
		return s.fakeLastlog()
	case "env":
		return s.fakeEnv()
	case "printenv":
		return s.fakePrintenv()
	case "set":
		return s.fakeSet()
	case "export":
		return s.fakeExport(args)
	
	// Éditeurs
	case "vi", "vim":
		return s.fakeVim(args)
	case "nano":
		return s.fakeNano(args)
	case "emacs":
		return s.fakeEmacs(args)
	
	// Bases de données
	case "mysql":
		return s.fakeMysql(args)
	case "psql":
		return s.fakePsql(args)
	case "mongo":
		return s.fakeMongo(args)
	case "redis-cli":
		return s.fakeRedisCli(args)
	
	// Langages de programmation
	case "python", "python3":
		return s.fakePython(args)
	case "php":
		return s.fakePhp(args)
	case "perl":
		return s.fakePerl(args)
	case "ruby":
		return s.fakeRuby(args)
	case "node":
		return s.fakeNode(args)
	case "bash":
		return s.fakeBash(args)
	case "sh":
		return s.fakeSh(args)
	
	// Logs et monitoring
	case "tail":
		return s.fakeTail(args)
	case "head":
		return s.fakeHead(args)
	case "less":
		return s.fakeLess(args)
	case "more":
		return s.fakeMore(args)
	case "dmesg":
		return s.fakeDmesg()
	case "journalctl":
		return s.fakeJournalctl(args)
	
	// Cron et tâches
	case "crontab":
		return s.fakeCrontab(args)
	case "at":
		return s.fakeAt(args)
	
	// Autres commandes utiles
	case "clear":
		return "\033[H\033[2J"
	case "date":
		return s.fakeDate()
	case "cal":
		return s.fakeCal()
	case "man":
		return s.fakeMan(args)
	case "help":
		return s.fakeHelp()
	case "which":
		return s.fakeWhich(args)
	case "whereis":
		return s.fakeWhereis(args)
	case "locate":
		return s.fakeLocate(args)
	case "du":
		return s.fakeDu(args)
	case "wc":
		return s.fakeWc(args)
	case "sort":
		return s.fakeSort(args)
	case "uniq":
		return s.fakeUniq(args)
	case "cut":
		return s.fakeCut(args)
	case "awk":
		return s.fakeAwk(args)
	case "sed":
		return s.fakeSed(args)
	case "tr":
		return s.fakeTr(args)
	
	// Sortie
	case "exit", "logout":
		return "Goodbye!\n"
	
	default:
		return fmt.Sprintf("bash: %s: command not found\n", cmd)
	}
}

// fakeLs simule la commande ls
func (s *FakeSession) fakeLs(args []string) string {
	// Afficher le contenu en fonction du répertoire courant
	switch s.currentDir {
	case "/home/user":
		output := "total 48\n"
		output += "drwxr-xr-x  2 user user 4096 Jan 15 10:30 .\n"
		output += "drwxr-xr-x  3 root root 4096 Jan 15 10:30 ..\n"
		output += "-rw-r--r--  1 user user  220 Jan 15 10:30 .bash_logout\n"
		output += "-rw-r--r--  1 user user 3771 Jan 15 10:30 .bashrc\n"
		output += "-rw-r--r--  1 user user  807 Jan 15 10:30 .profile\n"
		output += "drwxr-xr-x  2 user user 4096 Jan 15 10:30 Documents\n"
		output += "drwxr-xr-x  2 user user 4096 Jan 15 10:30 Downloads\n"
		output += "drwxr-xr-x  2 user user 4096 Jan 15 10:30 Pictures\n"
		output += "-rw-r--r--  1 user user   25 Jan 15 10:30 secret.txt\n"
		output += "-rwxr-xr-x  1 user user 8192 Jan 15 10:30 script.sh\n"
		return output
	case "/home/user/Documents":
		output := "total 16\n"
		output += "drwxr-xr-x  2 user user 4096 Jan 15 10:30 .\n"
		output += "drwxr-xr-x  5 user user 4096 Jan 15 10:30 ..\n"
		output += "-rw-r--r--  1 user user 2048 Jan 15 10:30 notes.txt\n"
		output += "-rw-r--r--  1 user user 4096 Jan 15 10:30 report.pdf\n"
		return output
	case "/home/user/Downloads":
		output := "total 12\n"
		output += "drwxr-xr-x  2 user user 4096 Jan 15 10:30 .\n"
		output += "drwxr-xr-x  5 user user 4096 Jan 15 10:30 ..\n"
		output += "-rw-r--r--  1 user user 1024 Jan 15 10:30 file.zip\n"
		return output
	case "/home/user/Pictures":
		output := "total 8\n"
		output += "drwxr-xr-x  2 user user 4096 Jan 15 10:30 .\n"
		output += "drwxr-xr-x  5 user user 4096 Jan 15 10:30 ..\n"
		output += "-rw-r--r--  1 user user 2048 Jan 15 10:30 photo.jpg\n"
		return output
	case "/":
		output := "total 84\n"
		output += "drwxr-xr-x  17 root root  4096 Jan 15 10:30 .\n"
		output += "drwxr-xr-x  17 root root  4096 Jan 15 10:30 ..\n"
		output += "drwxr-xr-x   2 root root  4096 Jan 15 10:30 bin\n"
		output += "drwxr-xr-x   3 root root  4096 Jan 15 10:30 boot\n"
		output += "drwxr-xr-x  16 root root  3840 Jan 15 10:30 dev\n"
		output += "drwxr-xr-x  94 root root  4096 Jan 15 10:30 etc\n"
		output += "drwxr-xr-x   3 root root  4096 Jan 15 10:30 home\n"
		output += "drwxr-xr-x  15 root root  4096 Jan 15 10:30 lib\n"
		output += "drwxr-xr-x   2 root root  4096 Jan 15 10:30 media\n"
		output += "drwxr-xr-x   2 root root  4096 Jan 15 10:30 mnt\n"
		output += "drwxr-xr-x   2 root root  4096 Jan 15 10:30 opt\n"
		output += "drwxr-xr-x   2 root root  4096 Jan 15 10:30 root\n"
		output += "drwxr-xr-x   6 root root  4096 Jan 15 10:30 run\n"
		output += "drwxr-xr-x   2 root root  4096 Jan 15 10:30 sbin\n"
		output += "drwxr-xr-x   2 root root  4096 Jan 15 10:30 tmp\n"
		output += "drwxr-xr-x  10 root root  4096 Jan 15 10:30 usr\n"
		output += "drwxr-xr-x  12 root root  4096 Jan 15 10:30 var\n"
		return output
	case "/etc":
		output := "total 20\n"
		output += "drwxr-xr-x  2 root root 4096 Jan 15 10:30 .\n"
		output += "drwxr-xr-x 17 root root 4096 Jan 15 10:30 ..\n"
		output += "-rw-r--r--  1 root root  220 Jan 15 10:30 passwd\n"
		output += "-rw-r--r--  1 root root  100 Jan 15 10:30 hosts\n"
		output += "-rw-r--r--  1 root root  150 Jan 15 10:30 hostname\n"
		return output
	case "/tmp":
		output := "total 4\n"
		output += "drwxrwxrwt  2 root root 4096 Jan 15 10:30 .\n"
		output += "drwxr-xr-x 17 root root 4096 Jan 15 10:30 ..\n"
		return output
	default:
		return "total 0\n"
	}
}

// fakePwd simule la commande pwd
func (s *FakeSession) fakePwd() string {
	return s.currentDir + "\n"
}

// fakeCd simule la commande cd
func (s *FakeSession) fakeCd(args []string) string {
	if len(args) == 0 {
		s.currentDir = "/home/user"
		return ""
	}
	
	target := args[0]
	
	// Gérer les chemins spéciaux
	if target == "~" || target == "~/" {
		s.currentDir = "/home/user"
		return ""
	}
	
	if target == ".." {
		// Remonter d'un niveau
		if s.currentDir != "/" {
			parts := strings.Split(strings.Trim(s.currentDir, "/"), "/")
			if len(parts) > 1 {
				s.currentDir = "/" + strings.Join(parts[:len(parts)-1], "/")
			} else {
				s.currentDir = "/"
			}
		}
		return ""
	}
	
	if target == "." {
		return ""
	}
	
	// Chemins absolus
	if strings.HasPrefix(target, "/") {
		// Simuler quelques dossiers existants
		validDirs := []string{"/", "/home", "/home/user", "/home/user/Documents", 
			"/home/user/Downloads", "/home/user/Pictures", "/etc", "/var", "/tmp", "/root"}
		for _, dir := range validDirs {
			if target == dir {
				s.currentDir = target
				return ""
			}
		}
		return fmt.Sprintf("bash: cd: %s: No such file or directory\n", target)
	}
	
	// Chemins relatifs
	newPath := s.currentDir + "/" + target
	newPath = strings.ReplaceAll(newPath, "//", "/")
	
	// Simuler quelques sous-dossiers
	validDirs := map[string][]string{
		"/home/user": {"Documents", "Downloads", "Pictures"},
		"/":          {"home", "etc", "var", "tmp", "root", "usr", "bin"},
		"/home":      {"user"},
	}
	
	if dirs, ok := validDirs[s.currentDir]; ok {
		for _, dir := range dirs {
			if target == dir {
				s.currentDir = newPath
				return ""
			}
		}
	}
	
	return fmt.Sprintf("bash: cd: %s: No such file or directory\n", target)
}

// fakeMkdir simule la commande mkdir
func (s *FakeSession) fakeMkdir(args []string) string {
	if len(args) == 0 {
		return "mkdir: missing operand\n"
	}
	// Simuler la création réussie
	return ""
}

// fakeTouch simule la commande touch
func (s *FakeSession) fakeTouch(args []string) string {
	if len(args) == 0 {
		return "touch: missing file operand\n"
	}
	// Simuler la création réussie
	return ""
}

// fakeRm simule la commande rm
func (s *FakeSession) fakeRm(args []string) string {
	if len(args) == 0 {
		return "rm: missing operand\n"
	}
	// Simuler la suppression réussie
	return ""
}

// fakeCp simule la commande cp
func (s *FakeSession) fakeCp(args []string) string {
	if len(args) < 2 {
		return "cp: missing file operand\n"
	}
	// Simuler la copie réussie
	return ""
}

// fakeMv simule la commande mv
func (s *FakeSession) fakeMv(args []string) string {
	if len(args) < 2 {
		return "mv: missing file operand\n"
	}
	// Simuler le déplacement réussi
	return ""
}

// fakeEcho simule la commande echo
func (s *FakeSession) fakeEcho(args []string) string {
	return strings.Join(args, " ") + "\n"
}

// fakeWhoami simule la commande whoami
func (s *FakeSession) fakeWhoami() string {
	return s.username + "\n"
}

// fakePs simule la commande ps
func (s *FakeSession) fakePs() string {
	output := "    PID TTY          TIME CMD\n"
	output += "   1234 pts/0    00:00:00 bash\n"
	output += "   1235 pts/0    00:00:00 ps\n"
	output += "   1001 ?        00:00:01 systemd\n"
	output += "   1002 ?        00:00:00 sshd\n"
	return output
}

// fakeNetstat simule la commande netstat
func (s *FakeSession) fakeNetstat() string {
	output := "Active Internet connections (w/o servers)\n"
	output += "Proto Recv-Q Send-Q Local Address           Foreign Address         State\n"
	output += "tcp        0      0 192.168.1.100:22        192.168.1.50:12345      ESTABLISHED\n"
	output += "tcp        0      0 192.168.1.100:80        0.0.0.0:*               LISTEN\n"
	output += "tcp        0      0 127.0.0.1:3306          0.0.0.0:*               LISTEN\n"
	return output
}

// fakeIfconfig simule la commande ifconfig
func (s *FakeSession) fakeIfconfig() string {
	output := "eth0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500\n"
	output += "        inet 192.168.1.100  netmask 255.255.255.0  broadcast 192.168.1.255\n"
	output += "        inet6 fe80::a00:27ff:fe4e:66a1  prefixlen 64  scopeid 0x20<link>\n"
	output += "        ether 08:00:27:4e:66:a1  txqueuelen 1000  (Ethernet)\n"
	output += "        RX packets 1234  bytes 123456 (123.4 KB)\n"
	output += "        RX errors 0  dropped 0  overruns 0  frame 0\n"
	output += "        TX packets 567  bytes 78901 (78.9 KB)\n"
	output += "        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0\n"
	return output
}

// fakeCat simule la commande cat
func (s *FakeSession) fakeCat(args []string) string {
	if len(args) == 0 {
		return "cat: missing file operand\n"
	}
	
	filename := args[0]
	
	// Gérer les chemins relatifs et absolus
	if !strings.HasPrefix(filename, "/") {
		filename = s.currentDir + "/" + filename
		filename = strings.ReplaceAll(filename, "//", "/")
	}
	
	switch filename {
	case "/home/user/secret.txt":
		return "This is a secret file with sensitive data!\n"
	case "/home/user/script.sh":
		return "#!/bin/bash\n# Backup script\ntar -czf backup.tar.gz /home/user/*\necho 'Backup completed'\n"
	case "/home/user/Documents/notes.txt":
		return "Important notes:\n- Server maintenance on Sunday\n- Password: admin123\n- Database backup location: /var/backups\n"
	case "/home/user/.bashrc":
		return "# .bashrc\nexport PATH=$PATH:/usr/local/bin\nalias ll='ls -la'\n"
	case "/etc/passwd":
		return "root:x:0:0:root:/root:/bin/bash\nuser:x:1000:1000:user:/home/user:/bin/bash\nadmin:x:1001:1001:admin:/home/admin:/bin/bash\n"
	case "/etc/hosts":
		return "127.0.0.1 localhost\n192.168.1.100 Tech.fr\n192.168.1.1 gateway\n"
	case "/etc/hostname":
		return "Tech.fr\n"
	default:
		return fmt.Sprintf("cat: %s: No such file or directory\n", args[0])
	}
}

// fakeGrep simule la commande grep
func (s *FakeSession) fakeGrep(args []string) string {
	if len(args) < 2 {
		return "grep: missing pattern or file\n"
	}
	return "grep: no matches found\n"
}

// fakeFind simule la commande find
func (s *FakeSession) fakeFind(args []string) string {
	return "./Documents\n./Downloads\n./Pictures\n./secret.txt\n./script.sh\n"
}

// fakeWget simule la commande wget
func (s *FakeSession) fakeWget(args []string) string {
	return "wget: command not found (simulated)\n"
}

// fakeCurl simule la commande curl
func (s *FakeSession) fakeCurl(args []string) string {
	return "curl: command not found (simulated)\n"
}

// fakeUname simule la commande uname
func (s *FakeSession) fakeUname() string {
	return "Linux Tech.fr 5.4.0-74-generic #83-Ubuntu SMP Sat May 8 02:35:39 UTC 2021 x86_64 x86_64 x86_64 GNU/Linux\n"
}

// fakeId simule la commande id
func (s *FakeSession) fakeId() string {
	return fmt.Sprintf("uid=1000(%s) gid=1000(%s) groups=1000(%s),4(adm),24(cdrom),27(sudo),30(dip),46(plugdev),120(lpadmin),131(lxd),132(sambashare)\n", 
		s.username, s.username, s.username)
}

// fakeGroups simule la commande groups
func (s *FakeSession) fakeGroups() string {
	return fmt.Sprintf("%s adm cdrom sudo dip plugdev lpadmin lxd sambashare\n", s.username)
}

// fakeUptime simule la commande uptime
func (s *FakeSession) fakeUptime() string {
	return " 10:30:45 up 2 days,  3:15,  1 user,  load average: 0.08, 0.02, 0.01\n"
}

// fakeDf simule la commande df
func (s *FakeSession) fakeDf() string {
	output := "Filesystem     1K-blocks    Used Available Use% Mounted on\n"
	output += "/dev/sda1       20971520 8388608  12582912  40% /\n"
	output += "tmpfs             1024000       0   1024000   0% /dev/shm\n"
	output += "/dev/sda2       52428800 1048576  51380224   2% /home\n"
	return output
}

// fakeFree simule la commande free
func (s *FakeSession) fakeFree() string {
	output := "              total        used        free      shared  buff/cache   available\n"
	output += "Mem:        2048000      512000      1024000       10240       512000     1536000\n"
	output += "Swap:       1048576           0     1048576\n"
	return output
}

// fakeTop simule la commande top
func (s *FakeSession) fakeTop() string {
	output := "top - 10:30:45 up 2 days,  3:15,  1 user,  load average: 0.08, 0.02, 0.01\n"
	output += "Tasks: 123 total,   1 running, 122 sleeping,   0 stopped,   0 zombie\n"
	output += "%Cpu(s):  0.1 us,  0.0 sy,  0.0 ni, 99.9 id,  0.0 wa,  0.0 hi,  0.0 si,  0.0 st\n"
	output += "MiB Mem :   2048.0 total,   1024.0 free,    512.0 used,    512.0 buff/cache\n"
	output += "MiB Swap:   1024.0 total,   1024.0 free,      0.0 used.   1536.0 avail Mem\n"
	output += "\n"
	output += "    PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND\n"
	output += "   1001 root      20   0  123456   1234    567 S   0.1   0.1   0:01.23 systemd\n"
	output += "   1002 root      20   0   56789    890    123 S   0.0   0.0   0:00.12 sshd\n"
	return output
}

// fakeHistory simule la commande history
func (s *FakeSession) fakeHistory() string {
	output := "    1  ls\n"
	output += "    2  pwd\n"
	output += "    3  whoami\n"
	output += "    4  ps\n"
	output += "    5  history\n"
	return output
}

// sendMessage envoie un message au client
func (s *FakeSession) sendMessage(message string) {
	if len(message) > 0 {
		_, err := s.channel.Write([]byte(message))
		if err != nil {
			// Ne pas logger les erreurs EOF (c'est normal quand le canal est fermé)
			if err != io.EOF && !strings.Contains(err.Error(), "EOF") {
				s.logger.Errorf("Error writing to channel: %v", err)
			}
		}
	}
}

// ========== NOUVELLES COMMANDES SYSTÈME CRITIQUES ==========

// fakeSudo simule la commande sudo
func (s *FakeSession) fakeSudo(args []string) string {
	if len(args) == 0 {
		return "usage: sudo -h | -K | -k | -V\nusage: sudo -v [-AknS] [-g group] [-h host] [-p prompt] [-u user]\nusage: sudo -l [-AknS] [-g group] [-h host] [-p prompt] [-U user] [-u user] [command]\nusage: sudo [-AbEHknPS] [-r role] [-t type] [-C num] [-g group] [-h host] [-p prompt] [-T timeout] [-u user] [VAR=value] [-i|-s] [<command>]\n"
	}
	
	// Si c'est "sudo su", simuler l'élévation vers root
	if len(args) >= 1 && args[0] == "su" {
		return "[sudo] password for " + s.username + ": \nroot@Tech.fr:/home/user# "
	}
	
	// Simuler l'exécution de la commande avec sudo
	return "[sudo] password for " + s.username + ": \nSorry, try again.\n[sudo] password for " + s.username + ": \nsudo: 3 incorrect password attempts\n"
}

// fakeSu simule la commande su
func (s *FakeSession) fakeSu(args []string) string {
	if len(args) == 0 {
		return "Password: \nsu: Authentication failure\n"
	}
	
	if args[0] == "root" || args[0] == "-" {
		return "Password: \nsu: Authentication failure\n"
	}
	
	return "Password: \nsu: Authentication failure\n"
}

// fakePasswd simule la commande passwd
func (s *FakeSession) fakePasswd(args []string) string {
	return "Changing password for " + s.username + ".\nCurrent password: \npasswd: Authentication token manipulation error\npasswd: password unchanged\n"
}

// fakeUseradd simule la commande useradd
func (s *FakeSession) fakeUseradd(args []string) string {
	if len(args) == 0 {
		return "useradd: missing operand\n"
	}
	return "useradd: Permission denied\n"
}

// fakeUserdel simule la commande userdel
func (s *FakeSession) fakeUserdel(args []string) string {
	if len(args) == 0 {
		return "userdel: missing operand\n"
	}
	return "userdel: Permission denied\n"
}

// fakeUsermod simule la commande usermod
func (s *FakeSession) fakeUsermod(args []string) string {
	if len(args) == 0 {
		return "usermod: missing operand\n"
	}
	return "usermod: Permission denied\n"
}

// fakeGroupadd simule la commande groupadd
func (s *FakeSession) fakeGroupadd(args []string) string {
	if len(args) == 0 {
		return "groupadd: missing operand\n"
	}
	return "groupadd: Permission denied\n"
}

// fakeGroupdel simule la commande groupdel
func (s *FakeSession) fakeGroupdel(args []string) string {
	if len(args) == 0 {
		return "groupdel: missing operand\n"
	}
	return "groupdel: Permission denied\n"
}

// fakeChmod simule la commande chmod
func (s *FakeSession) fakeChmod(args []string) string {
	if len(args) < 2 {
		return "chmod: missing operand\n"
	}
	return ""
}

// fakeChown simule la commande chown
func (s *FakeSession) fakeChown(args []string) string {
	if len(args) < 2 {
		return "chown: missing operand\n"
	}
	return "chown: changing ownership of '" + args[1] + "': Operation not permitted\n"
}

// fakeLn simule la commande ln
func (s *FakeSession) fakeLn(args []string) string {
	if len(args) < 2 {
		return "ln: missing file operand\n"
	}
	return ""
}

// fakeTar simule la commande tar
func (s *FakeSession) fakeTar(args []string) string {
	if len(args) == 0 {
		return "tar: You must specify one of the '-Acdtrux', '--delete' or '--test-label' options\n"
	}
	return "tar: Removing leading `/' from member names\n"
}

// fakeZip simule la commande zip
func (s *FakeSession) fakeZip(args []string) string {
	if len(args) < 2 {
		return "zip error: Missing file operands\n"
	}
	return "  adding: " + args[1] + " (stored 0%)\n"
}

// fakeUnzip simule la commande unzip
func (s *FakeSession) fakeUnzip(args []string) string {
	if len(args) == 0 {
		return "UnZip 6.00: missing file operand\n"
	}
	return "Archive:  " + args[0] + "\n  inflating: file.txt\n"
}

// fakeGzip simule la commande gzip
func (s *FakeSession) fakeGzip(args []string) string {
	if len(args) == 0 {
		return "gzip: missing operand\n"
	}
	return ""
}

// fakeGunzip simule la commande gunzip
func (s *FakeSession) fakeGunzip(args []string) string {
	if len(args) == 0 {
		return "gunzip: missing operand\n"
	}
	return ""
}

// ========== COMMANDES DE PROCESSUS ET SERVICES ==========

// fakeHtop simule la commande htop
func (s *FakeSession) fakeHtop() string {
	return "htop: command not found (use 'top' instead)\n"
}

// fakeKill simule la commande kill
func (s *FakeSession) fakeKill(args []string) string {
	if len(args) == 0 {
		return "kill: usage: kill [-s sigspec | -n signum | -sigspec] pid | jobspec ... or kill -l [sigspec]\n"
	}
	return ""
}

// fakeKillall simule la commande killall
func (s *FakeSession) fakeKillall(args []string) string {
	if len(args) == 0 {
		return "killall: no process name specified\n"
	}
	return args[0] + ": no process found\n"
}

// fakePkill simule la commande pkill
func (s *FakeSession) fakePkill(args []string) string {
	if len(args) == 0 {
		return "pkill: no matching criteria specified\n"
	}
	return ""
}

// fakeSystemctl simule la commande systemctl
func (s *FakeSession) fakeSystemctl(args []string) string {
	if len(args) == 0 {
		return "● Tech.fr\n    State: running\n     Jobs: 0 queued\n   Failed: 0 units\n"
	}
	
	action := args[0]
	switch action {
	case "status":
		if len(args) < 2 {
			return "● Tech.fr\n    State: running\n"
		}
		service := args[1]
		return "● " + service + "\n   Loaded: loaded (/lib/systemd/system/" + service + "; enabled)\n   Active: active (running)\n"
	case "start", "stop", "restart", "reload":
		return "Failed to " + action + " service: Access denied\n"
	case "enable", "disable":
		return "Failed to execute operation: Access denied\n"
	case "list-units":
		return "UNIT                           LOAD   ACTIVE SUB     DESCRIPTION\nsshd.service                   loaded active running OpenSSH server\nnginx.service                  loaded active running nginx web server\nmysql.service                  loaded active running MySQL Server\n"
	default:
		return "Unknown operation '" + action + "'\n"
	}
}

// fakeService simule la commande service
func (s *FakeSession) fakeService(args []string) string {
	if len(args) < 2 {
		return "Usage: service < option > | --status-all | [ service_name [ command | --full-restart ] ]\n"
	}
	
	service := args[0]
	action := args[1]
	
	switch action {
	case "status":
		return service + " is running\n"
	case "start", "stop", "restart":
		return "Failed to " + action + " " + service + ": Permission denied\n"
	default:
		return "Unrecognized service\n"
	}
}

// fakeReboot simule la commande reboot
func (s *FakeSession) fakeReboot() string {
	return "Failed to reboot system: Access denied\n"
}

// fakeShutdown simule la commande shutdown
func (s *FakeSession) fakeShutdown(args []string) string {
	return "Failed to shutdown system: Access denied\n"
}

// fakeInit simule la commande init
func (s *FakeSession) fakeInit(args []string) string {
	return "init: Need to be root\n"
}

// ========== COMMANDES RÉSEAU ==========

// fakePing simule la commande ping
func (s *FakeSession) fakePing(args []string) string {
	if len(args) == 0 {
		return "ping: usage error: Destination address required\n"
	}
	
	host := args[0]
	return "PING " + host + " (93.184.216.34) 56(84) bytes of data.\n64 bytes from " + host + ": icmp_seq=1 ttl=56 time=11.2 ms\n64 bytes from " + host + ": icmp_seq=2 ttl=56 time=10.8 ms\n^C\n--- " + host + " ping statistics ---\n2 packets transmitted, 2 received, 0% packet loss, time 1001ms\nrtt min/avg/max/mdev = 10.800/11.000/11.200/0.200 ms\n"
}

// fakeTraceroute simule la commande traceroute
func (s *FakeSession) fakeTraceroute(args []string) string {
	if len(args) == 0 {
		return "traceroute: missing host operand\n"
	}
	
	host := args[0]
	return "traceroute to " + host + " (93.184.216.34), 30 hops max, 60 byte packets\n 1  192.168.1.1 (192.168.1.1)  1.234 ms  1.123 ms  1.012 ms\n 2  10.0.0.1 (10.0.0.1)  5.678 ms  5.567 ms  5.456 ms\n 3  * * *\n"
}

// fakeNslookup simule la commande nslookup
func (s *FakeSession) fakeNslookup(args []string) string {
	if len(args) == 0 {
		return "nslookup: missing host operand\n"
	}
	
	host := args[0]
	return "Server:\t\t192.168.1.1\nAddress:\t192.168.1.1#53\n\nNon-authoritative answer:\nName:\t" + host + "\nAddress: 93.184.216.34\n"
}

// fakeDig simule la commande dig
func (s *FakeSession) fakeDig(args []string) string {
	if len(args) == 0 {
		return "dig: missing host operand\n"
	}
	
	host := args[0]
	return "; <<>> DiG 9.16.1-Ubuntu <<>> " + host + "\n;; global options: +cmd\n;; Got answer:\n;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 12345\n;; ANSWER SECTION:\n" + host + ".\t\t3600\tIN\tA\t93.184.216.34\n"
}

// fakeHost simule la commande host
func (s *FakeSession) fakeHost(args []string) string {
	if len(args) == 0 {
		return "host: missing host operand\n"
	}
	
	host := args[0]
	return host + " has address 93.184.216.34\n"
}

// fakeRoute simule la commande route
func (s *FakeSession) fakeRoute() string {
	return "Kernel IP routing table\nDestination     Gateway         Genmask         Flags Metric Ref    Use Iface\ndefault         192.168.1.1     0.0.0.0         UG    100    0        0 eth0\n192.168.1.0     0.0.0.0         255.255.255.0   U     100    0        0 eth0\n"
}

// fakeIptables simule la commande iptables
func (s *FakeSession) fakeIptables(args []string) string {
	return "iptables: Permission denied (you must be root)\n"
}

// fakeSs simule la commande ss
func (s *FakeSession) fakeSs() string {
	return "Netid  State      Recv-Q Send-Q Local Address:Port               Peer Address:Port\ntcp    ESTAB      0      0      192.168.1.100:22                192.168.1.50:12345\ntcp    LISTEN     0      128    0.0.0.0:80                      0.0.0.0:*\ntcp    LISTEN     0      128    127.0.0.1:3306                  0.0.0.0:*\n"
}

// fakeNetcat simule la commande netcat
func (s *FakeSession) fakeNetcat(args []string) string {
	if len(args) < 2 {
		return "nc: missing destination\n"
	}
	return "Connection refused\n"
}

// fakeTelnet simule la commande telnet
func (s *FakeSession) fakeTelnet(args []string) string {
	if len(args) == 0 {
		return "telnet: missing host operand\n"
	}
	return "Trying " + args[0] + "...\ntelnet: Unable to connect to remote host: Connection refused\n"
}

// fakeSsh simule la commande ssh
func (s *FakeSession) fakeSsh(args []string) string {
	if len(args) == 0 {
		return "usage: ssh [-46AaCfGgKkMNnqsTtVvXxYy] [-B bind_interface]\n"
	}
	return "ssh: connect to host " + args[0] + " port 22: Connection refused\n"
}

// fakeScp simule la commande scp
func (s *FakeSession) fakeScp(args []string) string {
	if len(args) < 2 {
		return "usage: scp [-346BCpqrTv] [-c cipher] [-F ssh_config] [-i identity_file]\n"
	}
	return "scp: Connection refused\n"
}

// fakeRsync simule la commande rsync
func (s *FakeSession) fakeRsync(args []string) string {
	if len(args) < 2 {
		return "rsync: missing destination\n"
	}
	return "rsync: failed to connect to " + args[len(args)-1] + ": Connection refused\n"
}

// fakeFtp simule la commande ftp
func (s *FakeSession) fakeFtp(args []string) string {
	if len(args) == 0 {
		return "ftp: missing host operand\n"
	}
	return "ftp: connect: Connection refused\n"
}

// ========== INFORMATIONS SYSTÈME ==========

// fakeHostname simule la commande hostname
func (s *FakeSession) fakeHostname() string {
	return "Tech.fr\n"
}

// fakeHostnamectl simule la commande hostnamectl
func (s *FakeSession) fakeHostnamectl() string {
	return "   Static hostname: Tech.fr\n         Icon name: computer-vm\n           Chassis: vm\n        Machine ID: 1234567890abcdef1234567890abcdef\n           Boot ID: abcdef1234567890abcdef1234567890\n    Virtualization: oracle\n  Operating System: Ubuntu 20.04.3 LTS\n            Kernel: Linux 5.4.0-74-generic\n      Architecture: x86-64\n"
}

// fakeLsbRelease simule la commande lsb_release
func (s *FakeSession) fakeLsbRelease() string {
	return "Distributor ID:\tUbuntu\nDescription:\tUbuntu 20.04.3 LTS\nRelease:\t20.04\nCodename:\tfocal\n"
}

// fakeWho simule la commande who
func (s *FakeSession) fakeWho() string {
	return s.username + "  pts/0        " + time.Now().Format("2006-01-02 15:04") + " (" + s.remoteAddr + ")\n"
}

// fakeLast simule la commande last
func (s *FakeSession) fakeLast() string {
	return s.username + "  pts/0        " + s.remoteAddr + "   " + time.Now().Format("Mon Jan 2 15:04") + " still logged in\nadmin    pts/1        192.168.1.50     Mon Dec 18 14:23 - 16:45  (02:22)\nroot     pts/0        192.168.1.100    Sun Dec 17 09:12 - 18:30  (09:18)\n"
}

// fakeLastlog simule la commande lastlog
func (s *FakeSession) fakeLastlog() string {
	return "Username         Port     From             Latest\nroot             pts/0    192.168.1.100    Sun Dec 17 09:12:45 +0000 2025\nadmin            pts/1    192.168.1.50     Mon Dec 18 14:23:12 +0000 2025\n" + s.username + "            pts/0    " + s.remoteAddr + "    " + time.Now().Format("Mon Jan 2 15:04:05 -0700 2006") + "\n"
}

// fakeEnv simule la commande env
func (s *FakeSession) fakeEnv() string {
	return "USER=" + s.username + "\nHOME=/home/" + s.username + "\nSHELL=/bin/bash\nPATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\nPWD=" + s.currentDir + "\nLANG=en_US.UTF-8\nSSH_CONNECTION=" + s.remoteAddr + " 192.168.1.100 22\n"
}

// fakePrintenv simule la commande printenv
func (s *FakeSession) fakePrintenv() string {
	return s.fakeEnv()
}

// fakeSet simule la commande set
func (s *FakeSession) fakeSet() string {
	return "BASH=/bin/bash\nBASHOPTS=checkwinsize:cmdhist:complete_fullquote:expand_aliases\nBASH_VERSION='5.0.17(1)-release'\nHOME=/home/" + s.username + "\nHOSTNAME=Tech.fr\nPATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\nPWD=" + s.currentDir + "\nSHELL=/bin/bash\nUSER=" + s.username + "\n"
}

// fakeExport simule la commande export
func (s *FakeSession) fakeExport(args []string) string {
	if len(args) == 0 {
		return s.fakeEnv()
	}
	return ""
}

// ========== ÉDITEURS ==========

// fakeVim simule la commande vim
func (s *FakeSession) fakeVim(args []string) string {
	return "Vim: Warning: Output is not to a terminal\nVim: Warning: Input is not from a terminal\n"
}

// fakeNano simule la commande nano
func (s *FakeSession) fakeNano(args []string) string {
	return "Error opening terminal: unknown.\n"
}

// fakeEmacs simule la commande emacs
func (s *FakeSession) fakeEmacs(args []string) string {
	return "emacs: standard input is not a tty\n"
}

// ========== BASES DE DONNÉES ==========

// fakeMysql simule la commande mysql
func (s *FakeSession) fakeMysql(args []string) string {
	return "ERROR 1045 (28000): Access denied for user '" + s.username + "'@'localhost' (using password: NO)\n"
}

// fakePsql simule la commande psql
func (s *FakeSession) fakePsql(args []string) string {
	return "psql: error: connection to server on socket \"/var/run/postgresql/.s.PGSQL.5432\" failed: No such file or directory\n"
}

// fakeMongo simule la commande mongo
func (s *FakeSession) fakeMongo(args []string) string {
	return "MongoDB shell version v4.4.6\nconnecting to: mongodb://127.0.0.1:27017/\nError: couldn't connect to server 127.0.0.1:27017\n"
}

// fakeRedisCli simule la commande redis-cli
func (s *FakeSession) fakeRedisCli(args []string) string {
	return "Could not connect to Redis at 127.0.0.1:6379: Connection refused\n"
}

// ========== LANGAGES DE PROGRAMMATION ==========

// fakePython simule la commande python
func (s *FakeSession) fakePython(args []string) string {
	if len(args) == 0 {
		return "Python 3.8.10 (default, Nov 14 2022, 12:59:47)\nType \"help\", \"copyright\", \"credits\" or \"license\" for more information.\n>>> \n"
	}
	return "python: can't open file '" + args[0] + "': [Errno 2] No such file or directory\n"
}

// fakePhp simule la commande php
func (s *FakeSession) fakePhp(args []string) string {
	if len(args) == 0 {
		return "Interactive shell\n\nphp > \n"
	}
	return "Could not open input file: " + args[0] + "\n"
}

// fakePerl simule la commande perl
func (s *FakeSession) fakePerl(args []string) string {
	if len(args) == 0 {
		return "Can't open perl script \"\": No such file or directory\n"
	}
	return "Can't open perl script \"" + args[0] + "\": No such file or directory\n"
}

// fakeRuby simule la commande ruby
func (s *FakeSession) fakeRuby(args []string) string {
	if len(args) == 0 {
		return "ruby: no Ruby script given\n"
	}
	return "ruby: No such file or directory -- " + args[0] + " (LoadError)\n"
}

// fakeNode simule la commande node
func (s *FakeSession) fakeNode(args []string) string {
	if len(args) == 0 {
		return "Welcome to Node.js v14.17.0.\nType \".help\" for more information.\n> \n"
	}
	return "Error: Cannot find module '" + args[0] + "'\n"
}

// fakeBash simule la commande bash
func (s *FakeSession) fakeBash(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return "bash: " + args[0] + ": No such file or directory\n"
}

// fakeSh simule la commande sh
func (s *FakeSession) fakeSh(args []string) string {
	return s.fakeBash(args)
}

// ========== LOGS ET MONITORING ==========

// fakeTail simule la commande tail
func (s *FakeSession) fakeTail(args []string) string {
	if len(args) == 0 {
		return "tail: missing file operand\n"
	}
	
	filename := args[len(args)-1]
	if filename == "/var/log/auth.log" {
		return "Dec 20 10:30:45 Tech.fr sshd[1234]: Accepted password for admin from 192.168.1.50 port 12345 ssh2\nDec 20 10:31:12 Tech.fr sshd[1235]: Failed password for invalid user test from 192.168.1.51 port 54321 ssh2\nDec 20 10:32:00 Tech.fr sshd[1236]: Connection closed by 192.168.1.52 port 11111 [preauth]\n"
	}
	
	return "tail: cannot open '" + filename + "' for reading: No such file or directory\n"
}

// fakeHead simule la commande head
func (s *FakeSession) fakeHead(args []string) string {
	if len(args) == 0 {
		return "head: missing file operand\n"
	}
	return "head: cannot open '" + args[0] + "' for reading: No such file or directory\n"
}

// fakeLess simule la commande less
func (s *FakeSession) fakeLess(args []string) string {
	if len(args) == 0 {
		return "Missing filename (\"less --help\" for help)\n"
	}
	return args[0] + ": No such file or directory\n"
}

// fakeMore simule la commande more
func (s *FakeSession) fakeMore(args []string) string {
	return s.fakeLess(args)
}

// fakeDmesg simule la commande dmesg
func (s *FakeSession) fakeDmesg() string {
	return "[    0.000000] Linux version 5.4.0-74-generic (buildd@lcy01-amd64-030)\n[    0.000000] Command line: BOOT_IMAGE=/boot/vmlinuz-5.4.0-74-generic root=UUID=1234-5678\n[    0.001234] Kernel command line: BOOT_IMAGE=/boot/vmlinuz-5.4.0-74-generic\n[    1.234567] smpboot: CPU0: Intel(R) Core(TM) i5-8250U CPU @ 1.60GHz\n[    2.345678] eth0: link up, 1000Mbps, full-duplex\n"
}

// fakeJournalctl simule la commande journalctl
func (s *FakeSession) fakeJournalctl(args []string) string {
	return "-- Logs begin at Mon 2025-12-18 09:00:00 UTC, end at Fri 2025-12-20 10:30:00 UTC. --\nDec 20 10:25:00 Tech.fr systemd[1]: Started Session 1 of user admin.\nDec 20 10:26:15 Tech.fr sshd[1234]: Accepted password for admin from 192.168.1.50\nDec 20 10:30:00 Tech.fr systemd[1]: Started Daily apt download activities.\n"
}

// ========== CRON ET TÂCHES ==========

// fakeCrontab simule la commande crontab
func (s *FakeSession) fakeCrontab(args []string) string {
	if len(args) == 0 {
		return "usage: crontab [-u user] file\n       crontab [-u user] [ -e | -l | -r ]\n"
	}
	
	if args[0] == "-l" {
		return "# m h  dom mon dow   command\n0 2 * * * /usr/bin/backup.sh\n30 3 * * 0 /usr/bin/cleanup.sh\n"
	}
	
	if args[0] == "-e" {
		return "no crontab for " + s.username + " - using an empty one\ncrontab: no changes made to crontab\n"
	}
	
	return ""
}

// fakeAt simule la commande at
func (s *FakeSession) fakeAt(args []string) string {
	return "Cannot find atd daemon\n"
}

// ========== AUTRES COMMANDES ==========

// fakeDate simule la commande date
func (s *FakeSession) fakeDate() string {
	return time.Now().Format("Mon Jan 2 15:04:05 MST 2006") + "\n"
}

// fakeCal simule la commande cal
func (s *FakeSession) fakeCal() string {
	return "    December 2025\nSu Mo Tu We Th Fr Sa\n    1  2  3  4  5  6\n 7  8  9 10 11 12 13\n14 15 16 17 18 19 20\n21 22 23 24 25 26 27\n28 29 30 31\n"
}

// fakeMan simule la commande man
func (s *FakeSession) fakeMan(args []string) string {
	if len(args) == 0 {
		return "What manual page do you want?\n"
	}
	return "No manual entry for " + args[0] + "\n"
}

// fakeHelp simule la commande help
func (s *FakeSession) fakeHelp() string {
	return "GNU bash, version 5.0.17(1)-release (x86_64-pc-linux-gnu)\nThese shell commands are defined internally.  Type `help' to see this list.\n\ncd [-L|[-P [-e]] [-@]] [dir]\nls [options] [file...]\npwd [-LP]\ncat [file...]\necho [arg...]\nexit [n]\n"
}

// fakeWhich simule la commande which
func (s *FakeSession) fakeWhich(args []string) string {
	if len(args) == 0 {
		return "which: missing operand\n"
	}
	
	cmd := args[0]
	commonCmds := map[string]string{
		"ls": "/bin/ls", "cat": "/bin/cat", "grep": "/bin/grep",
		"ps": "/bin/ps", "top": "/usr/bin/top", "sudo": "/usr/bin/sudo",
		"ssh": "/usr/bin/ssh", "wget": "/usr/bin/wget", "curl": "/usr/bin/curl",
		"python": "/usr/bin/python3", "php": "/usr/bin/php", "mysql": "/usr/bin/mysql",
	}
	
	if path, ok := commonCmds[cmd]; ok {
		return path + "\n"
	}
	
	return ""
}

// fakeWhereis simule la commande whereis
func (s *FakeSession) fakeWhereis(args []string) string {
	if len(args) == 0 {
		return "whereis: missing operand\n"
	}
	
	cmd := args[0]
	return cmd + ": /usr/bin/" + cmd + " /usr/share/man/man1/" + cmd + ".1.gz\n"
}

// fakeLocate simule la commande locate
func (s *FakeSession) fakeLocate(args []string) string {
	if len(args) == 0 {
		return "locate: missing operand\n"
	}
	return "/home/user/Documents/" + args[0] + "\n/home/user/Downloads/" + args[0] + "\n"
}

// fakeDu simule la commande du
func (s *FakeSession) fakeDu(args []string) string {
	return "4\t./Documents\n8\t./Downloads\n12\t./Pictures\n48\t.\n"
}

// fakeWc simule la commande wc
func (s *FakeSession) fakeWc(args []string) string {
	if len(args) == 0 {
		return "wc: missing operand\n"
	}
	return "  10  50  250 " + args[0] + "\n"
}

// fakeSort simule la commande sort
func (s *FakeSession) fakeSort(args []string) string {
	if len(args) == 0 {
		return "sort: missing operand\n"
	}
	return "sort: cannot read: " + args[0] + ": No such file or directory\n"
}

// fakeUniq simule la commande uniq
func (s *FakeSession) fakeUniq(args []string) string {
	if len(args) == 0 {
		return "uniq: missing operand\n"
	}
	return "uniq: " + args[0] + ": No such file or directory\n"
}

// fakeCut simule la commande cut
func (s *FakeSession) fakeCut(args []string) string {
	return "cut: you must specify a list of bytes, characters, or fields\n"
}

// fakeAwk simule la commande awk
func (s *FakeSession) fakeAwk(args []string) string {
	if len(args) == 0 {
		return "awk: cmd. line:1: fatal: cannot open file `' for reading (No such file or directory)\n"
	}
	return ""
}

// fakeSed simule la commande sed
func (s *FakeSession) fakeSed(args []string) string {
	if len(args) == 0 {
		return "sed: no input files\n"
	}
	return ""
}

// fakeTr simule la commande tr
func (s *FakeSession) fakeTr(args []string) string {
	if len(args) < 2 {
		return "tr: missing operand\n"
	}
	return ""
}

