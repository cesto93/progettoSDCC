#!/bin/bash

KEY_POS="/home/pier/Desktop/progetto_sdcc/myKey.pem"
CONF=$(<../../configuration/word_count.json)
#NAMES=( "master" "worker-1" "worker-2" "worker-3" )
#PORTS=( "1050" "1051" "1052" "1053" )

NAMES=( $(echo $CONF | jq '.aws[].name') )
PORTS=( $(echo $CONF | jq '.aws[].port') )

#workers
for (( i=1; i<${#NAMES[@]}; i++ ));
do
INSTANCES[$i]=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${NAMES[$i]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
INSTANCE_DNS=$(echo ${INSTANCES[$i]} | jq -r '.[].PublicDnsName')
IP[$i]=$(echo ${INSTANCES[$i]} | jq  -r '.[].NetworkInterfaces[].Association.PublicIp')
#addresses[$i]="${IP[$i]}:${PORTS[$i]}"

ssh -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@$INSTANCE_DNS "
./go/src/progettoSDCC/source/application/word_counter/worker/worker ${PORTS[$i]}" &
done

AWS_ADDR="\"${IP[1]}:${PORTS[1]}\""
for (( i=2; i<${#NAMES[@]}; i++ ));
do
AWS_ADDR="$AWS_ADDR, \"${IP[$i]}:${PORTS[$i]}\""
done

#importing json data for gc
AWS_ADDR_J="[$AWS_ADDR]"
GC_ADDR_J=$(<../common_generated_data/gc_workers.json)
ADDR_J=$(echo $AWS_ADDR_J $GC_ADDR_J | jq -s -r 'add | flatten')


#master 
#TODO this should have addresses of google cloud by using common generated data
INSTANCES[0]=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${NAMES[0]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
INSTANCE_DNS=$(echo ${INSTANCES[0]} | jq -r '.[].PublicDnsName')
IP[0]=$(echo ${INSTANCES[0]} | jq  -r '.[].NetworkInterfaces[].Association.PublicIp')
addresses[0]="${IP[0]}:${PORTS[0]}"
ssh  -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@$INSTANCE_DNS "
./go/src/progettoSDCC/source/application/word_counter/master/master -workerAddr ${addresses[1]},${addresses[2]},${addresses[3]} -masterPort ${PORTS[0]}" &
