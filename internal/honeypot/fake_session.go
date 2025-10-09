package honeypot

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"honey/internal/config"
	"honey/internal/database"
	"honey/internal/logger"
	"honey/internal/models"
)

// FakeSession représente une session SSH factice
type FakeSession struct {
	channel    ssh.Channel
	requests   <-chan *ssh.Request
	config     *config.Config
	logger     logger.Logger
	username   string
	remoteAddr string
	connectionID int
}

// handle gère une session SSH factice
func (s *FakeSession) handle() {
	defer s.channel.Close()

	// Enregistrer la connexion réussie
	connection := &models.Connection{
		RemoteAddr:  s.remoteAddr,
		Username:    s.username,
		Success:     true,
		ConnectedAt: time.Now(),
	}

	if err := database.SaveConnection(connection); err != nil {
		s.logger.Errorf("Failed to save successful connection: %v", err)
	} else {
		s.connectionID = connection.ID
	}

	// Traiter les requêtes de la session
	go s.handleRequests()

	// Afficher le message de bienvenue
	s.sendMessage(s.config.Shell.WelcomeMessage + "\n")
	s.sendMessage("Last login: " + time.Now().Format("Mon Jan 2 15:04:05 2006") + " from " + s.remoteAddr + "\n")

	// Démarrer le shell factice
	s.runFakeShell()
}

// handleRequests traite les requêtes SSH
func (s *FakeSession) handleRequests() {
	for req := range s.requests {
		switch req.Type {
		case "shell":
			// Accepter la requête shell
			if req.WantReply {
				req.Reply(true, nil)
			}
		case "pty-req":
			// Accepter la requête PTY
			if req.WantReply {
				req.Reply(true, nil)
			}
		case "window-change":
			// Accepter le changement de taille de fenêtre
			if req.WantReply {
				req.Reply(true, nil)
			}
		case "exec":
			// Support des commandes non interactives: ssh host "cmd"
			var payload struct {
				Command string `ssh:"string"`
			}
			if err := ssh.Unmarshal(req.Payload, &payload); err != nil {
				if req.WantReply {
					req.Reply(false, nil)
				}
				continue
			}

			// Enregistrer et exécuter la commande factice
			cmd := &models.Command{
				ConnectionID: s.connectionID,
				Command:      strings.TrimSpace(payload.Command),
				ExecutedAt:   time.Now(),
			}
			if err := database.SaveCommand(cmd); err != nil {
				s.logger.Errorf("Failed to save command (exec): %v", err)
			}

			s.logger.Infof("Exec command from %s: %s", s.remoteAddr, payload.Command)
			response := s.executeFakeCommand(payload.Command)
			s.sendMessage(response)

			if req.WantReply {
				req.Reply(true, nil)
			}

			// Fermer le canal après l'exec comme le ferait un vrai serveur
			s.channel.CloseWrite()
			return
		default:
			// Refuser les autres requêtes
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}
}

// runFakeShell exécute le shell factice
func (s *FakeSession) runFakeShell() {
	reader := bufio.NewReader(s.channel)
	
	for {
		// Afficher le prompt
		s.sendMessage(s.config.Shell.Prompt)

		// Lire la commande (pas de timeout pour éviter les coupures)
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				s.logger.Errorf("Error reading command: %v", err)
			}
			break
		}

		// Nettoyer la commande
		command := strings.TrimSpace(line)
		if command == "" {
			continue
		}

		// Gérer les commandes de sortie
		if command == "exit" || command == "logout" {
			s.sendMessage("Goodbye!\n")
			break
		}

		// Enregistrer la commande
		cmd := &models.Command{
			ConnectionID: s.connectionID,
			Command:      command,
			ExecutedAt:   time.Now(),
		}

		if err := database.SaveCommand(cmd); err != nil {
			s.logger.Errorf("Failed to save command: %v", err)
		}

		s.logger.Infof("Command executed from %s: %s", s.remoteAddr, command)

		// Exécuter la commande factice
		response := s.executeFakeCommand(command)
		s.sendMessage(response)
	}
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
	case "ls":
		return s.fakeLs(args)
	case "pwd":
		return s.fakePwd()
	case "whoami":
		return s.fakeWhoami()
	case "ps":
		return s.fakePs()
	case "netstat":
		return s.fakeNetstat()
	case "ifconfig":
		return s.fakeIfconfig()
	case "cat":
		return s.fakeCat(args)
	case "grep":
		return s.fakeGrep(args)
	case "find":
		return s.fakeFind(args)
	case "wget":
		return s.fakeWget(args)
	case "curl":
		return s.fakeCurl(args)
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
	case "top":
		return s.fakeTop()
	case "history":
		return s.fakeHistory()
	case "exit", "logout":
		return "Goodbye!\n"
	default:
		return fmt.Sprintf("bash: %s: command not found\n", cmd)
	}
}

// fakeLs simule la commande ls
func (s *FakeSession) fakeLs(args []string) string {
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
}

// fakePwd simule la commande pwd
func (s *FakeSession) fakePwd() string {
	return "/home/user\n"
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
	switch filename {
	case "secret.txt":
		return "This is a secret file with sensitive data!\n"
	case "/etc/passwd":
		return "root:x:0:0:root:/root:/bin/bash\nuser:x:1000:1000:user:/home/user:/bin/bash\n"
	case "/etc/hosts":
		return "127.0.0.1 localhost\n192.168.1.100 honeypot\n"
	default:
		return fmt.Sprintf("cat: %s: No such file or directory\n", filename)
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
	return "Linux honeypot 5.4.0-74-generic #83-Ubuntu SMP Sat May 8 02:35:39 UTC 2021 x86_64 x86_64 x86_64 GNU/Linux\n"
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
	s.channel.Write([]byte(message))
}

