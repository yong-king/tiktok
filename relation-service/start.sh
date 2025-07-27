#!/bin/sh

/app/wait-for-it.sh mysql 3306 -t 60
/app/wait-for-it.sh redis 6379 -t 60
/app/wait-for-it.sh consul-server 8500 -t 60
/app/wait-for-it.sh user-service 8081 -t 60
/app/wait-for-it.sh elasticsearch 9200 -t 60


echo "All dependencies are up. Starting relation-service..."
exec ./ralation-service -conf /app/configs/config_doc.yaml
