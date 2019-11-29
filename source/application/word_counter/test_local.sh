#!/bin/bash

CONF=$(<../../../configuration/generated/app_node.json)
WORKERS_IP=( $(echo $CONF | jq -r '.workers[].address') )
WORKERS_PORT=( $(echo $CONF | jq -r '.workers[].port') )

cd ./worker
#workers
for (( i=0; i<${#WORKERS_PORT[@]}; i++ ));
do
konsole --new-tab --noclose -e ./worker ${WORKERS_PORT[$i]} &
done


