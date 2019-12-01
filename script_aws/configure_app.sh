#!/bin/bash

KEY_POS="/home/pier/Desktop/progetto_sdcc/myKey.pem"
CONF=$(<../configuration/word_count.json)
GC_ADDR_J=$(<../configuration/generated/gc_workers.json)

#importing configuration
NAMES=( $(echo $CONF | jq -r '.aws[].name') )
PORTS=( $(echo $CONF | jq -r '.aws[].port') )

#workers
for (( i=1; i<${#NAMES[@]}; i++ ));
do
INSTANCES[$i]=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${NAMES[$i]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
INSTANCE_DNS[$i]=$(echo ${INSTANCES[$i]} | jq -r '.[].PublicDnsName')
IP[$i]=$(echo ${INSTANCES[$i]} | jq  -r '.[].NetworkInterfaces[].Association.PublicIp')
WORKERS[$i]=$(jq -n --arg addr "${IP[$i]}" --arg port "${PORTS[$i]}" '[{"address": $addr, "port": $port}]')
done
WORKERS_J=$(echo ${WORKERS[@]} | jq -s 'add')
APP_NODE=$(jq -n --arg mport ${PORTS[0]} --argjson workers "$WORKERS_J" '{masterport: $mport , workers : $workers}')
echo $APP_NODE

#TODO import json dynamic data for gc

#ssh conn
for (( i=1; i<${#NAMES[@]}; i++ ));
do
konsole --new-tab --noclose -e ssh -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@${INSTANCE_DNS[$i]} \
"
cd ./go/src/progettoSDCC
git pull git@github.com:cesto93/progettoSDCC
go build -o ./bin/worker ./source/application/word_counter/worker/worker.go
mkdir -p ./configuration/generated
echo '$APP_NODE' | tee ./configuration/generated/app_node.json
" &
done

#master 
INSTANCES[0]=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${NAMES[0]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
INSTANCE_DNS[0]=$(echo ${INSTANCES[0]} | jq -r '.[].PublicDnsName')
IP[0]=$(echo ${INSTANCES[0]} | jq  -r '.[].NetworkInterfaces[].Association.PublicIp')

konsole --new-tab --noclose -e ssh  -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@${INSTANCE_DNS[0]} \
"
cd ./go/src/progettoSDCC
git pull git@github.com:cesto93/progettoSDCC
go build -o ./bin/master ./source/application/word_counter/master/master.go
mkdir -p ./configuration/generated
echo '$APP_NODE' | tee ./configuration/generated/app_node.json
" &

echo "Connect client to ${IP[0]}:${PORTS[0]}"
