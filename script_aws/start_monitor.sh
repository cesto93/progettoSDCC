#!/bin/bash

KEY_POS="/home/pier/Desktop/progetto_sdcc/myKey.pem"
CONF=$(<../configuration/monitor.json)

#importing configuration
ZK_SRV_NAMES=( $(echo $CONF | jq -r '.servers_zk.names[]') )
MONITOR_NAMES=( $(echo $CONF | jq -r '.aws[].name') )

#zk_servers
for (( i=0; i<${#ZK_SRV_NAMES[@]}; i++ ));
do
	INST=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${ZK_SRV_NAMES[$i]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
	INST_DNS=$(echo $INST | jq -r '.[].PublicDnsName')
	konsole --new-tab --noclose -e ssh -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@$INST_DNS "sudo ./zookeeper/bin/zkServer.sh start" &
done

#monitor
for (( i=0; i<${#MONITOR_NAMES[@]}; i++ ));
do
	INST=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${MONITOR_NAMES[$i]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
	INST_DNS=$(echo $INST | jq -r '.[].PublicDnsName')
	konsole --new-tab --noclose -e ssh -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@$INST_DNS "./go/src/progettoSDCC/source/monitoring/agent -aws" &
done