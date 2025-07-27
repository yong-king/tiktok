#!/bin/sh


/app/wait-for-it.sh mysql 3306
/app/wait-for-it.sh redis 6379
/app/wait-for-it.sh consul-server 8500
/app/wait-for-it.sh jaeger 4317

echo "All dependencies are up. Starting user-service..."
exec ./user-service -conf /app/configs/config.yaml
