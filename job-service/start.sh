#!/bin/sh

/app/wait-for-it.sh kafka 19092
/app/wait-for-it.sh elasticsearch 9200


echo "All dependencies are up. Starting job-service..."
exec ./job-service -conf /app/configs/config_doc.yaml
