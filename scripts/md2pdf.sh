#!/bin/bash
# Script de conversion Markdown vers PDF
# Usage: ./md2pdf.sh [fichier.md] [sortie.pdf]

# Activer le mode strict
set -e

# Couleurs pour les messages
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Fonction d'aide
show_help() {
    echo "Usage: $0 [fichier.md] [sortie.pdf]"
    echo ""
    echo "Convertit un fichier Markdown en PDF professionnel"
    echo ""
    echo "Arguments:"
    echo "  fichier.md   - Fichier Markdown source (défaut: CAHIER_DES_CHARGES_FINAL.md)"
    echo "  sortie.pdf   - Fichier PDF de sortie (défaut: même nom avec .pdf)"
    echo ""
    echo "Exemples:"
    echo "  $0                                    # Convertir le cahier des charges"
    echo "  $0 README.md                          # Convertir README.md en README.pdf"
    echo "  $0 docs/SECURITY.md security.pdf      # Convertir avec nom personnalisé"
}

# Vérifier si on demande l'aide
if [[ "$1" == "-h" ]] || [[ "$1" == "--help" ]]; then
    show_help
    exit 0
fi

# Répertoire du script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Vérifier si l'environnement virtuel existe
if [ ! -d "$PROJECT_ROOT/.venv" ]; then
    echo -e "${RED}❌ Environnement virtuel Python non trouvé${NC}"
    echo -e "${BLUE}📦 Création de l'environnement virtuel...${NC}"
    cd "$PROJECT_ROOT"
    python3 -m venv .venv

    echo -e "${BLUE}📦 Installation des dépendances...${NC}"
    .venv/bin/pip install -q markdown weasyprint
    echo -e "${GREEN}✅ Environnement créé et configuré${NC}"
fi

# Activer l'environnement virtuel
source "$PROJECT_ROOT/.venv/bin/activate"

# Arguments
INPUT_FILE="${1:-$PROJECT_ROOT/docs/CAHIER_DES_CHARGES_FINAL.md}"
OUTPUT_FILE="${2}"

# Si pas de fichier de sortie spécifié, utiliser le même nom avec .pdf
if [ -z "$OUTPUT_FILE" ]; then
    OUTPUT_FILE="${INPUT_FILE%.md}.pdf"
fi

# Vérifier que le fichier source existe
if [ ! -f "$INPUT_FILE" ]; then
    echo -e "${RED}❌ Erreur: Fichier source non trouvé: $INPUT_FILE${NC}"
    exit 1
fi

echo -e "${BLUE}📄 Conversion de Markdown vers PDF${NC}"
echo -e "${BLUE}   Source: $INPUT_FILE${NC}"
echo -e "${BLUE}   Sortie: $OUTPUT_FILE${NC}"
echo ""

# Lancer la conversion
python3 "$SCRIPT_DIR/convert_to_pdf.py" "$INPUT_FILE" "$OUTPUT_FILE"

# Vérifier que le PDF a été créé
if [ -f "$OUTPUT_FILE" ]; then
    SIZE=$(du -h "$OUTPUT_FILE" | cut -f1)
    echo ""
    echo -e "${GREEN}✅ Conversion réussie !${NC}"
    echo -e "${GREEN}   Fichier: $OUTPUT_FILE${NC}"
    echo -e "${GREEN}   Taille: $SIZE${NC}"
else
    echo -e "${RED}❌ Erreur: Le fichier PDF n'a pas été créé${NC}"
    exit 1
fi
