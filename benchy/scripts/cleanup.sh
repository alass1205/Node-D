#!/bin/bash
echo "🧹 Cleaning up existing benchy containers..."

# Arrêter tous les containers benchy
docker ps | grep benchy | awk '{print $1}' | xargs -r docker stop

# Supprimer tous les containers benchy
docker ps -a | grep benchy | awk '{print $1}' | xargs -r docker rm

# Supprimer le réseau benchy
docker network rm benchy-network 2>/dev/null || true

echo "✅ Cleanup completed"
