#!/bin/bash
# Redis Cache Monitoring Script

echo "=========================================="
echo "   Redis Cache Monitor"
echo "=========================================="
echo ""

# Check Redis status
echo "🔌 Redis Status:"
docker-compose exec redis redis-cli PING
echo ""

# Show all keys
echo "🔑 Cached Keys:"
docker-compose exec redis redis-cli KEYS "*"
echo ""

# Show database size
echo "📊 Total Keys:"
docker-compose exec redis redis-cli DBSIZE
echo ""

# Show memory usage
echo "💾 Memory Usage:"
docker-compose exec redis redis-cli INFO memory | grep used_memory_human
echo ""

# Show key details
echo "📝 Key Details:"
for key in $(docker-compose exec -T redis redis-cli KEYS "*" 2>/dev/null); do
    if [ ! -z "$key" ]; then
        ttl=$(docker-compose exec -T redis redis-cli TTL "$key")
        type=$(docker-compose exec -T redis redis-cli TYPE "$key")
        echo "  $key"
        echo "    Type: $type | TTL: ${ttl}s"
    fi
done
echo ""

echo "=========================================="
echo "Commands to inspect cache:"
echo "  docker-compose exec redis redis-cli KEYS \"*\""
echo "  docker-compose exec redis redis-cli GET \"task:TASK_ID\""
echo "  docker-compose exec redis redis-cli TTL \"task:TASK_ID\""
echo "  docker-compose exec redis redis-cli FLUSHDB  # Clear all"
echo "=========================================="
