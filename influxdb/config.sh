#!/bin/bash
#set -e
INFLUX=./influx

USER=admin
PWD="admin123!"

$INFLUX setup -u $USER -o mntr -b mntr -p $PWD -n mntr -f
$INFLUX auth create --description "Mntr collector auth" -o mntr --write-buckets
$INFLUX auth create --description "Mntr frontend auth" -o mntr --read-buckets
echo "❗❗❗ NOTE - Make a note of the token and place it in your collector .env as MNTR_INFLUXDB_TOKEN"