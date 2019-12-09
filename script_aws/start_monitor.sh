#!/bin/bash
source ./conf/key.sh
CONF=$(<../configuration/monitor.json)

#importing configuration
ZK_SRV_NAMES=( $(echo $CONF | jq -r '.servers_zk.names[]') )
MONITOR_NAMES=( $(echo $CONF | jq -r '.aws[].name') )
DB_NAME=$(echo $CONF | jq -r '.db.name')

#influxdb 
INST=$(aws ec2 describe-instances --filters Name=tag:Name,Values=$DB_NAME  --query 'Reservations[*].Instances[*]' | jq 'flatten')
INST_DNS_DB=$(echo $INST | jq -r '.[].PublicDnsName')
ssh  -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@$INST_DNS_DB "./influxdb-1.7.9-1/usr/bin/influxd &"

#zk_servers
for (( i=0; i<${#ZK_SRV_NAMES[@]}; i++ ));
do
	INST=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${ZK_SRV_NAMES[$i]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
	INST_DNS=$(echo $INST | jq -r '.[].PublicDnsName')
	ssh -q -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@$INST_DNS "sudo ./zookeeper/bin/zkServer.sh start"
done

#monitor
for (( i=0; i<${#MONITOR_NAMES[@]}; i++ ));
do
	INST=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${MONITOR_NAMES[$i]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
	INST_DNS=$(echo $INST | jq -r '.[].PublicDnsName')
	konsole --new-tab --noclose -e ssh -q -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@$INST_DNS \
"
echo 'This is ${MONITOR_NAMES[$i]}'
cd ./go/src/progettoSDCC/bin
./agent -aws
" &
done
