#!/bin/sh

/app/wait-for-it.sh mysql 3306 -t 60
/app/wait-for-it.sh redis 6379 -t 60
/app/wait-for-it.sh consul-server 8500 -t 60
/app/wait-for-it.sh user-service 8081 -t 60
/app/wait-for-it.sh video-service 8082 -t 60


echo "All dependencies are up. Starting favorite-service..."
exec ./favorite-service -conf /app/configs/config_doc.yaml
