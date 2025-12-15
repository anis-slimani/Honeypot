#!/bin/bash
# Script pour vérifier le contenu de la base de données

echo "=========================================="
echo "📊 VÉRIFICATION DE LA BASE DE DONNÉES"
echo "=========================================="
echo ""

if [ ! -f "honeypot.db" ]; then
    echo "❌ Base de données honeypot.db introuvable !"
    exit 1
fi

echo "✅ Base de données trouvée"
echo ""

echo "📋 CONNEXIONS :"
sqlite3 honeypot.db "SELECT id, remote_addr, username, success, datetime(connected_at, 'localtime') FROM connections ORDER BY id DESC LIMIT 10;"
echo ""

echo "⌨️  COMMANDES :"
sqlite3 honeypot.db "SELECT c.id, c.connection_id, c.command, cn.remote_addr, cn.username, datetime(c.executed_at, 'localtime') FROM commands c LEFT JOIN connections cn ON c.connection_id = cn.id ORDER BY c.id DESC LIMIT 20;"
echo ""

echo "📊 STATISTIQUES :"
echo "Total connexions: $(sqlite3 honeypot.db 'SELECT COUNT(*) FROM connections;')"
echo "Total commandes: $(sqlite3 honeypot.db 'SELECT COUNT(*) FROM commands;')"
echo ""

