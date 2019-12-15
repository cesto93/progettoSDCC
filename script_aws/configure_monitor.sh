#!/bin/bash
source ./conf/key.sh
CONF_MONITOR=$(<../configuration/monitor.json)

#importing configuration
MONITOR_NAMES=( $(echo $CONF_MONITOR | jq -r '.aws[].name') )
DB_NAME=$(echo $CONF_MONITOR | jq -r '.db.name')
ZK_SRV_NAMES=( $(echo $CONF_MONITOR | jq -r '.servers_zk_aws.names[]') )
ZK_CLIENT_PORT=$(echo $CONF_MONITOR | jq -r '.servers_zk_aws.port_client')
ZK_SERVER_PORTS=$(echo $CONF_MONITOR | jq -r '.servers_zk_aws.ports_server')
PORT_DB="8086"

#getting instance DNS and ID
for (( i=0; i<${#MONITOR_NAMES[@]}; i++ ));
do
	INST=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${MONITOR_NAMES[$i]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
	INST_DNS[$i]=$(echo $INST | jq -r '.[].PublicDnsName')
	ID_MONITOR_J[$i]=$(echo $INST | jq '[.[].InstanceId]')
done
IDS_MONITOR_J=$(echo ${ID_MONITOR_J[@]} | jq -s 'add | flatten')

#getting zk_srv data
for (( i=0; i<${#ZK_SRV_NAMES[@]}; i++ ));
do
	INST=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${ZK_SRV_NAMES[$i]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
	ZK_SRV_IP[$i]=$(echo $INST | jq  -r '.[].NetworkInterfaces[].PrivateIpAddress')
	ZK_SRV_IP_J[$i]=$(echo $INST | jq -r '[.[].NetworkInterfaces[].PrivateIpAddress]')
	INST_DNS_SRV[$i]=$(echo $INST | jq -r '.[].PublicDnsName')
done
ZK_SRV_IPS_J=$(echo ${ZK_SRV_IP_J[@]} | jq -s --arg port ":$ZK_CLIENT_PORT" ' add |  flatten |[.[] + $port] ') 

#getting db data
INST=$(aws ec2 describe-instances --filters Name=tag:Name,Values=$DB_NAME  --query 'Reservations[*].Instances[*]' | jq 'flatten')
IP_DB=$(echo $INST | jq -r '.[].NetworkInterfaces[].Association.PublicIp')
ADDR_DB="$IP_DB:$PORT_DB"
ADDR_DB_J=$(jq -n --arg addr "$ADDR_DB" '$addr' ) 

echo $ADDR_DB_J > ../configuration/generated/db_addr.json

#configuration of json parameters and project pull and compile
for (( i=0; i<${#MONITOR_NAMES[@]}; i++ ));
do
	unset TARGS
	unset BASH_TARGS
	unset IP_MONITORED_J
	unset ID_MONITORED_J
	MONITORED_NAMES=( $(echo $CONF_MONITOR | jq -r --argjson index "$i" '.aws[$index].monitored[]') )
	for (( j=0; j<${#MONITORED_NAMES[@]}; j++ ));
	do
		INST=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${MONITORED_NAMES[$j]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
		ID_J=$(echo $INST | jq '[.[].InstanceId]')
		IP_J=$(echo $INST | jq '[.[].NetworkInterfaces[].PrivateIpAddress]')
		ID_MONITORED_J[$j]=$(echo $ID_J | jq -s '.[]' )
		IP_MONITORED_J[$j]=$(echo $IP_J | jq -s '.[]' )
	done
	ID_MONITORED_M_J=$(echo ${ID_MONITORED_J[@]} | jq -s 'add')
	BASH_TARGS=( $(echo ${IP_MONITORED_J[@]} | jq -r '.[]' ) )
	
	TARGS="'${BASH_TARGS[0]}:9100'"
	for ((j=1; j<${#BASH_TARGS[@]}; j++));
	do
		TARGS="$TARGS, '${BASH_TARGS[j]}:9100'"
	done

	ssh  -q -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@${INST_DNS[$i]} \
"
echo 'starting  ${MONITOR_NAMES[$i]} configuration'
cd ./go/src/progettoSDCC
git pull http://github.com/cesto93/progettoSDCC -q
go build -o ./bin/agent ./source/monitoring/agent.go

echo '$IDS_MONITOR_J' | tee ./configuration/generated/zk_agent.json
echo '$ZK_SRV_IPS_J' | tee ./configuration/generated/zk_servers_addrs.json
echo '$i' | tee ./configuration/generated/id_monitor.json
echo '$ID_MONITORED_M_J' | tee ./configuration/generated/ec2_inst.json
echo '$ADDR_DB_J' | tee ./configuration/generated/db_addr.json
#prometheus
sudo tee /etc/prometheus/prometheus.yml<<EOF
global:
  scrape_interval:     15s
  evaluation_interval: 15s

rule_files:
  # - "first.rules"
  # - "second.rules"

scrape_configs:
  - job_name: prometheus
    static_configs:
      - targets: [$TARGS]
EOF
sudo systemctl restart prometheus
echo 'finished' 
" &
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
echo -e 'finished zookeper server configuration $MYID\n' 
" &

done

wait
