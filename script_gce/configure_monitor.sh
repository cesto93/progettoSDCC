#!/bin/bash

ZONE="us-central1-a"
CONF_MONITOR=$(<../configuration/monitor.json)

MONITOR_NAMES=( $(echo $CONF_MONITOR | jq -r '.gc[].name') )
ZK_SRV_NAMES=( $(echo $CONF_MONITOR | jq -r '.servers_zk_gce.names[]') )
ZK_CLIENT_PORT=$(echo $CONF_MONITOR | jq -r '.servers_zk_gce.port_client')
ZK_SERVER_PORTS=$(echo $CONF_MONITOR | jq -r '.servers_zk_gce.ports_server')

#db addr
ADDR_DB_J=$(<../configuration/generated/db_addr.json)

#zk_serv_addr
for (( i=0; i<${#ZK_SRV_NAMES[@]}; i++ ));
do
	INST=$(gcloud compute instances describe ${ZK_SRV_NAMES[$i]} --zone "us-central1-a" --format "json")
	ZK_SRV_IP[$i]=$(echo $INST | jq  -r -s ".[].networkInterfaces[0].networkIP")
	ZK_SRV_IP_J[$i]=$(echo $INST | jq -r -s "[.[].networkInterfaces[0].networkIP]")
done
ZK_SRV_IPS_J=$(echo ${ZK_SRV_IP_J[@]} | jq -s --arg port ":$ZK_CLIENT_PORT" ' add |  flatten |[.[] + $port] ') 

#monitor_ids
for ((i=0; i<${#MONITOR_NAMES[@]}; i++));
do
IDS[$i]=$(gcloud compute instances describe ${MONITOR_NAMES[i]} --zone "us-central1-a" --format "json" | jq -s ".[].id")
done
IDS_MONITOR_J=$(echo ${IDS[@]} | jq -s '[.[]]')

#set monitor conf files
for ((i=0; i<${#MONITOR_NAMES[@]}; i++));
do
	MONITORED_NAMES=( $(echo $CONF_MONITOR | jq -r --argjson index "$i" '.gc[$index].monitored[]') )
	M=$(echo ${MONITORED_NAMES[@]})
	gcloud compute ssh --zone=$ZONE ${MONITOR_NAMES[i]} --command \
"
echo configuration of monitor
cd ~/go/src/progettoSDCC
git pull git@github.com:cesto93/progettoSDCC -q
go build -o ./bin/agent ./source/monitoring/agent.go
cd ./script_gce
export PATH=$PATH:/snap/bin
bash ./set_instances_to_monitor.sh $M
cd ../
echo '$IDS_MONITOR_J' | tee ./configuration/generated/zk_agent.json
echo '$ZK_SRV_IPS_J' | tee ./configuration/generated/zk_servers_addrs.json
echo '$i' | tee ./configuration/generated/id_monitor.json
echo '$ADDR_DB_J' | tee ./configuration/generated/db_addr.json
" &
done

#zookeeper server conf file
MYID=0
for (( i=0; i<${#ZK_SRV_NAMES[@]}; i++ ));
do
	((MYID++))
	gcloud compute ssh --zone=$ZONE ${ZK_SRV_NAMES[i]}  --command \
"
echo 'tickTime=250
initLimit=10
syncLimit=5
dataDir=/var/lib/zookeeper
clientPort=$ZK_CLIENT_PORT
server.1=${ZK_SRV_IP[0]}$ZK_SERVER_PORTS
server.2=${ZK_SRV_IP[1]}$ZK_SERVER_PORTS
server.3=${ZK_SRV_IP[2]}$ZK_SERVER_PORTS' | sudo tee ./zookeeper/conf/zoo.cfg
sudo sh -c 'echo '$MYID' > /var/lib/zookeeper/myid'
echo -e 'finished zookeper server configuration $MYID\n' 
" &
done

wait
