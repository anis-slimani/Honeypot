# Guide de Conversion PDF

Ce guide explique comment convertir n'importe quel fichier Markdown en PDF professionnel.

---

## 🎯 Convertisseur MD → PDF

Le projet inclut un convertisseur Markdown vers PDF professionnel qui génère des documents avec :

- ✅ **Mise en page professionnelle** (marges, en-têtes, numérotation de pages)
- ✅ **Tableaux bien formatés** avec alternance de couleurs
- ✅ **Code colorisé** avec fond gris et bordures
- ✅ **Titres hiérarchisés** avec couleurs et bordures
- ✅ **Liens cliquables** en bleu
- ✅ **Évite les coupures de page** mal placées

---

## 📦 Installation

Les outils nécessaires sont installés automatiquement lors de la première utilisation.

**Dépendances :**
- Python 3.x (déjà installé)
- `markdown` (installation automatique)
- `weasyprint` (installation automatique)

---

## 🚀 Utilisation rapide

### Convertir le cahier des charges (défaut)

```bash
cd /home/kali/Honeypot
./scripts/md2pdf.sh
```

Cela génère : `docs/CAHIER_DES_CHARGES_FINAL.pdf`

### Convertir un fichier spécifique

```bash
# Convertir README.md en README.pdf
./scripts/md2pdf.sh README.md

# Convertir avec nom personnalisé
./scripts/md2pdf.sh docs/SECURITY_DOCUMENTATION.md securite.pdf

# Convertir n'importe quel document
./scripts/md2pdf.sh docs/QUICK_START.md docs/QUICK_START.pdf
```

### Afficher l'aide

```bash
./scripts/md2pdf.sh --help
```

---

## 🔧 Utilisation avancée (Python direct)

Si vous préférez utiliser directement le script Python :

```bash
# Activer l'environnement virtuel
source .venv/bin/activate

# Convertir un fichier
python scripts/convert_to_pdf.py docs/FICHIER.md sortie.pdf

# Désactiver l'environnement
deactivate
```

---

## 📄 Fichiers créés

| Fichier | Description |
|---------|-------------|
| `scripts/md2pdf.sh` | Script shell principal (facile à utiliser) |
| `scripts/convert_to_pdf.py` | Script Python de conversion |
| `.venv/` | Environnement virtuel Python (créé automatiquement) |

---

## 🎨 Style du PDF

Le PDF généré inclut :

### En-tête et numérotation
- Numérotation de pages en haut à droite : "Page X / Y"
- Marges professionnelles : 2.5cm haut/bas, 2cm gauche/droite

### Titres
- **H1** : Bleu foncé (#2c3e50), 24pt, bordure bleue
- **H2** : Gris foncé (#34495e), 18pt, bordure grise
- **H3** : Gris (#555), 14pt
- **H4** : Gris clair (#666), 12pt

### Tableaux
- En-têtes bleus (#3498db) avec texte blanc
- Lignes alternées (gris clair / blanc)
- Bordures subtiles

### Code
- Fond gris clair (#f4f4f4)
- Police monospace (DejaVu Sans Mono)
- Bordure bleue à gauche pour les blocs de code

### Listes et paragraphes
- Texte justifié
- Interligne 1.6 pour meilleure lisibilité
- Marges cohérentes

---

## 🐛 Dépannage

### Erreur : "Environnement virtuel non trouvé"

Le script le crée automatiquement. Si problème :

```bash
cd /home/kali/Honeypot
python3 -m venv .venv
.venv/bin/pip install markdown weasyprint
```

### Erreur : "Module 'markdown' not found"

```bash
.venv/bin/pip install markdown weasyprint
```

### Le PDF est vide ou mal formaté

Vérifiez que votre fichier Markdown est valide :
- Encodage UTF-8
- Syntaxe Markdown correcte
- Pas de caractères spéciaux non supportés

### Permission refusée

```bash
chmod +x scripts/md2pdf.sh
```

---

## 📝 Exemples d'utilisation

### 1. Documentation complète du projet

```bash
# Créer des PDFs pour toute la documentation importante
./scripts/md2pdf.sh docs/CAHIER_DES_CHARGES_FINAL.md
./scripts/md2pdf.sh docs/SECURITY_DOCUMENTATION.md
./scripts/md2pdf.sh docs/QUICK_START.md
./scripts/md2pdf.sh README.md
```

### 2. Rapport de projet

```bash
# Convertir le résumé des modifications
./scripts/md2pdf.sh docs/RESUME_MODIFICATIONS.md rapport_projet.pdf
```

### 3. Guide de présentation

```bash
# Créer un PDF pour la présentation
./scripts/md2pdf.sh docs/PRACTICAL_SECURITY_EXAMPLES.md demo_securite.pdf
```

---

## 💡 Conseils

**Pour une meilleure qualité PDF :**

1. **Utilisez des titres hiérarchiques** : # H1, ## H2, ### H3
2. **Incluez des tableaux** : ils sont automatiquement stylisés
3. **Utilisez des listes** : à puces ou numérotées
4. **Ajoutez du code** : avec triple backticks \`\`\`
5. **Séparez les sections** : avec des lignes horizontales `---`

**Évitez :**
- Images très lourdes (le PDF sera volumineux)
- Tableaux trop larges (risque de débordement)
- Trop de niveaux de titres (max H4 recommandé)

---

## 🔄 Conversion en masse

Pour convertir plusieurs fichiers à la fois :

```bash
# Créer un script de conversion en masse
for file in docs/*.md; do
    echo "Conversion de $file..."
    ./scripts/md2pdf.sh "$file"
done
```

---

## 📋 Format supporté

### Markdown standard
- Titres (`#`, `##`, `###`)
- Listes (à puces et numérotées)
- Liens `[texte](url)`
- Images `![alt](url)`
- Gras `**texte**`
- Italique `*texte*`
- Code inline `` `code` ``
- Blocs de code ` ```langue ... ``` `
- Tableaux
- Citations `> texte`
- Lignes horizontales `---`

### Extensions activées
- `tables` - Support complet des tableaux
- `fenced_code` - Blocs de code avec backticks
- `nl2br` - Sauts de ligne automatiques
- `sane_lists` - Listes améliorées

---

## ✅ Vérification du PDF

Après conversion, vérifiez :

```bash
# Taille du fichier
ls -lh docs/CAHIER_DES_CHARGES_FINAL.pdf

# Type de fichier
file docs/CAHIER_DES_CHARGES_FINAL.pdf

# Ouvrir le PDF (si navigateur disponible)
xdg-open docs/CAHIER_DES_CHARGES_FINAL.pdf
```

---

## 📚 Ressources

- **WeasyPrint Documentation** : https://doc.courtbouillon.org/weasyprint/
- **Python Markdown** : https://python-markdown.github.io/
- **Syntaxe Markdown** : https://www.markdownguide.org/

---

**Créé le :** 2025-12-20
**Dernière mise à jour :** 2025-12-20
