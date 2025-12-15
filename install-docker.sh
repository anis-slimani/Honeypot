#!/bin/bash

echo "════════════════════════════════════════════════════════"
echo "🐳 INSTALLATION DE DOCKER"
echo "════════════════════════════════════════════════════════"
echo ""

# Vérifier si Docker est déjà installé
if command -v docker &> /dev/null; then
    echo "✅ Docker est déjà installé : $(docker --version)"
    echo ""
    read -p "Voulez-vous réinstaller Docker ? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 0
    fi
fi

echo "📦 Mise à jour du système..."
sudo apt update

echo ""
echo "📦 Installation des prérequis..."
sudo apt install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg \
    lsb-release

echo ""
echo "🔑 Ajout de la clé GPG Docker..."
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

echo ""
echo "📝 Ajout du dépôt Docker..."
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

echo ""
echo "📦 Installation de Docker et Docker Compose..."
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin docker-compose

echo ""
echo "🚀 Démarrage de Docker..."
sudo systemctl start docker
sudo systemctl enable docker

echo ""
echo "👤 Ajout de l'utilisateur au groupe docker..."
sudo usermod -aG docker $USER

echo ""
echo "✅ Installation terminée !"
echo ""
echo "📋 Versions installées :"
docker --version
docker-compose --version

echo ""
echo "⚠️  IMPORTANT : Vous devez vous DÉCONNECTER et vous RECONNECTER"
echo "    pour que les changements de groupe prennent effet."
echo ""
echo "Après reconnexion, testez avec : docker run hello-world"
echo ""

