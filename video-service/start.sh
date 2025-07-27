#!/bin/sh

/app/wait-for-it.sh mysql 3306
/app/wait-for-it.sh redis 6379
/app/wait-for-it.sh minio 9000
/app/wait-for-it.sh consul-server 8500
/app/wait-for-it.sh elasticsearch 9200


echo "All dependencies are up. Starting video-service..."
exec ./video-service -conf /app/configs/config_doc.yaml
