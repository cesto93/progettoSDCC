#!/bin/bash

KEY_POS="/home/pier/Desktop/progetto_sdcc/myKey.pem"
NODE_NAMES=( "master" "worker-1" "worker-2" "worker-3" )
AGENT_POS=( 0 1 2 3 )
SERVER_POS=( 0 1 2 )

#zk_servers
for (( i=0; i<${#SERVER_POS[@]}; i++ ));
do
$j=${SERVER_POS[$i]}
INSTANCES[$j]=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${AGENT_NAMES[$j]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
INSTANCE_DNS=$(echo ${INSTANCES[$j]} | jq -r '.[].PublicDnsName')
ssh -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@$INSTANCE_DNS "./zookeeper/bin/zkServer.sh start" &
done

#monitor
for (( i=0; i<${#AGENT_NAMES[@]}; i++ ));
do
$j=${AGENT_POS[$i]}
INSTANCES[$j]=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${AGENT_NAMES[$j]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
INSTANCE_DNS=$(echo ${INSTANCES[$j]} | jq -r '.[].PublicDnsName')
#TODO load instance to monitor in ec2_instances
ssh -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@$INSTANCE_DNS "./go/src/progettoSDCC/source/monitoring/agent -aws" &
done
