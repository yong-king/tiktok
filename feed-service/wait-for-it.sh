#!/usr/bin/env sh
# wait-for-it.sh

host="$1"
port="$2"
shift 2
timeout=30

until nc -z "$host" "$port"; do
  echo "Waiting for $host:$port..."
  sleep 2
  timeout=$((timeout - 2))
  if [ "$timeout" -le 0 ]; then
    echo "Timeout waiting for $host:$port"
    exit 1
  fi
done

exec "$@"
