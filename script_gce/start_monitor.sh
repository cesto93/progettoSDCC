#!/bin/bash
ZONE="us-central1-a"
CONF=$(<../configuration/monitor.json)

#importing configuration
ZK_SRV_NAMES=( $(echo $CONF | jq -r '.servers_zk_gce.names[]') )
MONITOR_NAMES=( $(echo $CONF | jq -r '.gc[].name') )


#zk_servers
for (( i=0; i<${#ZK_SRV_NAMES[@]}; i++ ));
do
	INST=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${ZK_SRV_NAMES[$i]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
	INST_DNS=$(echo $INST | jq -r '.[].PublicDnsName')
	gcloud compute ssh --zone=$ZONE ${ZK_SRV_NAMES[i]} --command "sudo ./zookeeper/bin/zkServer.sh start"
done

#monitor
for (( i=0; i<${#MONITOR_NAMES[@]}; i++ ));
do
	INST=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${MONITOR_NAMES[$i]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
	INST_DNS=$(echo $INST | jq -r '.[].PublicDnsName')
	gcloud compute ssh --zone=$ZONE ${MONITOR_NAMES[i]} --command \
"
echo 'This is ${MONITOR_NAMES[$i]}'
cd ./go/src/progettoSDCC/bin
./agent
" &
done
