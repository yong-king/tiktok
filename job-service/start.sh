#!/bin/sh

/app/wait-for-it.sh mysql 3306 -t 60
/app/wait-for-it.sh canal 11111 -t 60
/app/wait-for-it.sh kafka 19092 -t 60
/app/wait-for-it.sh elasticsearch 9200 -t 60

echo "All dependencies are up. Starting job-service..."
exec ./job-service -conf /app/configs/config_doc.yaml
