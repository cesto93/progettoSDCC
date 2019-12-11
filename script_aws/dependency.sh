#!/bin/bash
source ./conf/key.sh
CONF=$(<../configuration/word_count.json)
CONF_MONITOR=$(<../configuration/monitor.json)

#importing configuration
NAMES=( $(echo $CONF | jq -r '.aws[].name') )
ZK_SRV_NAMES=( $(echo $CONF_MONITOR | jq -r '.servers_zk.names[]') )
DB_NAME=$(echo $CONF_MONITOR | jq -r '.db.name')

#depency and project clone
for (( i=0; i<${#NAMES[@]}; i++ ));
do
	INST=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${NAMES[$i]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
	INST_DNS[$i]=$(echo $INST | jq -r '.[].PublicDnsName')
	#echo -e "COPYING SSH KEY FOR GIT DEPLOYMENT TO ${INST_DNS[$i]} \n"
	#scp  -q -o "StrictHostKeyChecking=no" -i $KEY_POS $LOCAL_DIR/$GIT_KEY_FILE  ec2-user@${INST_DNS[$i]}:$AWS_DIR
	#scp -q -o "StrictHostKeyChecking=no" -i $KEY_POS $LOCAL_DIR/config  ec2-user@${INST_DNS[$i]}:$AWS_DIR
	scp -q -o "StrictHostKeyChecking=no" -i $KEY_POS ./prometheus.sh  ec2-user@${INST_DNS[$i]}:/home/ec2-user/
	konsole --new-tab --noclose -e ssh  -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@${INST_DNS[$i]} \
"
sudo yum update -y -q
sudo yum install git -y -q -e 0
sudo yum install golang -y -q -e 0
go get -u github.com/aws/aws-sdk-go
go get -u cloud.google.com/go/monitoring/apiv3
go get -u github.com/samuel/go-zookeeper/zk
go get github.com/influxdata/influxdb1-client/v2

sudo ./prometheus.sh

# Generating Prometheus configuration file
sudo tee /etc/prometheus/prometheus.yml <<EOF 1> /dev/null
# my global config
global:
  scrape_interval:     15s # Set the scrape interval to every 15 seconds. Default is every 1 minute.
  evaluation_interval: 15s # Evaluate rules every 15 seconds. The default is every 1 minute.
  # scrape_timeout is set to the global default (10s).

scrape_configs:
  - job_name: 'MonitorInstances'
    file_sd_configs:
      - files:
        - /home/ec2-user/go/src/progettoSDCC/configuration/generated/instances.json
EOF

cd ./go/src
sudo rm -rf progettoSDCC
git clone git@github.com:cesto93/progettoSDCC
mkdir -p ./progettoSDCC/configuration/generated
mkdir -p ./progettoSDCC/configuration/log

echo 'finished installing preliminary dependency on  ${NAMES[$i]}'
" &
done

#influxdb install 
#TODO set db here
INST=$(aws ec2 describe-instances --filters Name=tag:Name,Values=$DB_NAME  --query 'Reservations[*].Instances[*]' | jq 'flatten')
INST_DNS_DB=$(echo $INST | jq -r '.[].PublicDnsName')
konsole --new-tab --noclose -e ssh  -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@$INST_DNS_DB \
"
echo 'intalling db in $DB_NAME'
wget -q -nc https://dl.influxdata.com/influxdb/releases/influxdb-1.7.9_linux_amd64.tar.gz
tar xvfz influxdb-1.7.9_linux_amd64.tar.gz
cd ./influxdb-1.7.9-1/usr/bin/
#./influxd &
#./influx -precision rfc3339 
#CREATE DATABASE mydb
echo 'done'
" 

#zookeeper intall
for (( i=0; i<${#ZK_SRV_NAMES[@]}; i++ ));
do
	INST=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${ZK_SRV_NAMES[$i]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
	INST_DNS_SRV[$i]=$(echo $INST | jq -r '.[].PublicDnsName')
	konsole --new-tab --noclose -e ssh  -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@${INST_DNS_SRV[$i]} \
"
sudo yum install java-1.8.0-openjdk -y -q
sudo wget -q -nc https://www-us.apache.org/dist/zookeeper/zookeeper-3.5.6/apache-zookeeper-3.5.6-bin.tar.gz
sudo tar -xzf  apache-zookeeper-3.5.6-bin.tar.gz
sudo mv -n apache-zookeeper-3.5.6-bin ./zookeeper
sudo mkdir -p /var/lib/zookeeper
echo 'finished installing zk_server on ${ZK_SRV_NAMES[$i]}' 
" &
done

#source ./configure_monitor.sh
#source ./configure_app.sh
