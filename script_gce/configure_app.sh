#!/bin/bash

ZONE="us-central1-a"
CONF=$(<../configuration/word_count.json)

#importing configuration
NAMES=( $(echo $CONF | jq -r '.gc[].name') )
MPORT=$(echo $CONF | jq -r '.aws[0].port')
PORTS=( $(echo $CONF | jq -r '.gc[].port') )

#workers
for (( i=0; i<${#NAMES[@]}; i++ ));
do
IP[$i]=$(gcloud compute instances describe ${NAMES[$i]} --zone=$ZONE --format='get(networkInterfaces[0].accessConfigs[0].natIP)')
WORKERS[$i]=$(jq -n --arg addr "${IP[$i]}" --arg port "${PORTS[$i]}" '[{"address": $addr, "port": $port}]')
done
WORKERS_J=$(echo ${WORKERS[@]} | jq -s 'add')
echo $WORKERS_J > "../configuration/generated/gc_workers.json"
APP_NODE=$(jq -n --arg mport $MPORT --argjson workers "$WORKERS_J" '{masterport: $mport , workers : $workers}')

#ssh conn
for (( i=0; i<${#NAMES[@]}; i++ ));
do
gcloud compute ssh --zone=$ZONE ${NAMES[$i]} --command \
"
cd ./go/src/progettoSDCC
git pull git@github.com:cesto93/progettoSDCC -q
go build -o ./bin/worker ./source/application/word_counter/worker/worker.go
echo '$APP_NODE' | tee ./configuration/generated/app_node.json
echo '$i' | tee ./configuration/generated/id_worker.json
" -q &
done

wait
