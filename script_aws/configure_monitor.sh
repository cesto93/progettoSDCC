#!/bin/bash

KEY_POS="/home/pier/Desktop/progetto_sdcc/myKey.pem"

#importing configuration
CONF_MONITOR=$(<../configuration/monitor.json)
ZK_SRV_NAMES=( $(echo $CONF_MONITOR | jq -r '.servers_zk.names[]') )
MONITOR_NAMES=( $(echo $CONF_MONITOR | jq -r '.aws[].name') )
ZK_CLIENT_PORT=$(echo $CONF_MONITOR | jq -r '.servers_zk.port_client')
ZK_SERVER_PORTS=$(echo $CONF_MONITOR | jq -r '.servers_zk.ports_server')

#getting instance DNS and ID
declare -A ID_MAP_J
for (( i=0; i<${#MONITOR_NAMES[@]}; i++ ));
do
	INST=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${MONITOR_NAMES[$i]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
	INST_DNS[$i]=$(echo $INST | jq -r '.[].PublicDnsName')
	ID_MONITOR_J[$i]=$(echo $INST | jq '[.[].InstanceId]')
	ID_MAP_J[${MONITOR_NAMES[$i]}]=$(echo $INST | jq '[.[].InstanceId]')
done

for (( i=0; i<${#ZK_SRV_NAMES[@]}; i++ ));
do
	INST=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${ZK_SRV_NAMES[$i]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
	#ZK_SRV_IP[$i]=$(echo $INST | jq  -r '.[].NetworkInterfaces[].Association.PublicIp')
	#ZK_SRV_IP_J[$i]=$(echo $INST | jq '[.[].NetworkInterfaces[].Association.PublicIp]')
	ZK_SRV_IP[$i]=$(echo $INST | jq  -r '.[].NetworkInterfaces[].PrivateIpAddress')
	ZK_SRV_IP_J[$i]=$(echo $INST | jq '[.[].NetworkInterfaces[].PrivateIpAddress]')
	INST_DNS_SRV[$i]=$(echo $INST | jq -r '.[].PublicDnsName')
done

IDS_MONITOR_J=$(echo ${ID_MONITOR_J[@]} | jq -s 'add | flatten')
ZK_SRV_IPS_J=$(echo ${ZK_SRV_IP_J[@]} | jq -s --arg port ":$ZK_CLIENT_PORT" ' add |  flatten |[.[] + $port] ') 

#configuration of json parameters and project pull and compile
for (( i=0; i<${#MONITOR_NAMES[@]}; i++ ));
do
	MONITORED_NAMES=( $(echo $CONF_MONITOR | jq -r --argjson index "$i" '.aws[$index].monitored[]') )
	for (( j=0; j<${#MONITORED_NAMES[@]}; j++ ));
	do
		ID_MONITORED_J[$j]=$(echo ${ID_MAP_J[${MONITORED_NAMES[$j]}]} | jq -s '.[]' )
	done
	ID_MONITORED_M_J=$(echo ${ID_MONITORED_J[@]} | jq -s 'add')
	ssh  -q -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@${INST_DNS[$i]} \
"
cd ./go/src/progettoSDCC
git pull git@github.com:cesto93/progettoSDCC -q
go build -o ./bin/agent ./source/monitoring/agent.go

echo '$IDS_MONITOR_J' | tee ./configuration/generated/zk_agent.json
echo '$ZK_SRV_IPS_J' | tee ./configuration/generated/zk_servers_addrs.json
echo '$ID_MONITORED_M_J' | tee ./configuration/generated/ec2_inst.json
echo 'finished ${MONITOR_NAMES[$i]}' 
"
done

#zookeeper server conf file
MYID=0
for (( i=0; i<${#ZK_SRV_NAMES[@]}; i++ ));
do
	((MYID++))
	ssh  -q -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@${INST_DNS_SRV[$i]} \
"
echo 'tickTime=250
initLimit=10
syncLimit=5
dataDir=/var/lib/zookeeper
clientPort=$ZK_CLIENT_PORT
server.1=${ZK_SRV_IP[0]}$ZK_SERVER_PORTS
server.2=${ZK_SRV_IP[1]}$ZK_SERVER_PORTS
server.3=${ZK_SRV_IP[2]}$ZK_SERVER_PORTS' | tee ./zookeeper/conf/zoo.cfg
sudo sh -c 'echo '$MYID' > /var/lib/zookeeper/myid'
echo -e 'finished server $MYID\n' 
"
done
