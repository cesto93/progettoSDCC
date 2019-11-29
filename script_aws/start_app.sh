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
INSTANCE_DNS=$(echo ${INSTANCES[$i]} | jq -r '.[].PublicDnsName')
IP[$i]=$(echo ${INSTANCES[$i]} | jq  -r '.[].NetworkInterfaces[].Association.PublicIp')

konsole --new-tab --noclose -e ssh -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@$INSTANCE_DNS \
"./go/src/progettoSDCC/source/application/word_counter/worker/worker ${PORTS[$i]}" &
done

AWS_ADDR=$(echo ${IP[1]}:${PORTS[1]})
for (( i=2; i<${#NAMES[@]}; i++ ));
do
AWS_ADDR=$(echo $AWS_ADDR,${IP[$i]}":"${PORTS[$i]})
done

#TODO import json dynamic data for gc
#AWS_ADDR_J="[$AWS_ADDR]"
#ADDR_J=$(echo $AWS_ADDR_J $GC_ADDR_J | jq -s -r 'add | flatten')

#master 
#TODO this should have addresses of google cloud by using common generated data

INSTANCES[0]=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${NAMES[0]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
INSTANCE_DNS=$(echo ${INSTANCES[0]} | jq -r '.[].PublicDnsName')
IP[0]=$(echo ${INSTANCES[0]} | jq  -r '.[].NetworkInterfaces[].Association.PublicIp')

konsole --new-tab --noclose -e ssh  -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@$INSTANCE_DNS \
"./go/src/progettoSDCC/source/application/word_counter/master/master -workerAddr $AWS_ADDR -masterPort ${PORTS[0]}" &

echo "Connect client to ${IP[0]}:${PORTS[0]}"

