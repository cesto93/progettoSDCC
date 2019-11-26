#!/bin/bash

KEY_POS="/home/pier/Desktop/progetto_sdcc/myKey.pem"
NAMES=( "master" "worker-1" "worker-2" "worker-3" )
GIT_KEY_FILE="sdcc_git"
LOCAL_DIR="/home/pier/.ssh"
AWS_DIR="/home/ec2-user/.ssh"
ZK_CLIENT_PORT=":2181"

for (( i=0; i<${#NAMES[@]}; i++ ));
do
ID[$i]=$(aws ec2 describe-instances --filters Name=tag:Name,Values=$MASTER_NAME  --query 'Reservations[*].Instances[*].InstanceId')

INSTANCE_DNS[$i]=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${NAMES[$i]} --output text \
		--query 'Reservations[*].Instances[*].PublicDnsName')

IP[$i]=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${NAMES[$i]} --output text \
		--query 'Reservations[*].Instances[*].NetworkInterfaces[*].Association.PublicIp')
done

for (( i=0; i<${#NAMES[@]}; i++ ));
do

echo -e "COPYING SSH KEY FOR GIT DEPLOYMENT TO ${INSTANCE_DNS[$i]} \n"
scp  -q -o "StrictHostKeyChecking=no" -i $KEY_POS $LOCAL_DIR/$GIT_KEY_FILE  ec2-user@${INSTANCE_DNS[$i]}:$AWS_DIR
scp -q -o "StrictHostKeyChecking=no" -i $KEY_POS $LOCAL_DIR/config  ec2-user@${INSTANCE_DNS[$i]}:$AWS_DIR

echo -e "SSH CONNECTION TO ${NAMES[$i]}\n"

konsole --new-tab --noclose -e ssh  -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@${INSTANCE_DNS[$i]} "

sudo yum update -y -q
sudo yum install git -y -q -e 0
sudo yum install golang -y -q -e 0
sudo yum install jq -y -q -e 0
go get -u github.com/aws/aws-sdk-go
go get -u github.com/samuel/go-zookeeper/zk

#zookeeper configuration
sudo wget -q -nc https://www-us.apache.org/dist/zookeeper/zookeeper-3.5.6/apache-zookeeper-3.5.6-bin.tar.gz
sudo tar -xzf  apache-zookeeper-3.5.6-bin.tar.gz
#sudo rm -rf ./zookeeper
sudo mv -n apache-zookeeper-3.5.6-bin ./zookeeper

#zookeeper conf file
echo 'tickTime=250
initLimit=10
syncLimit=5
dataDir=/var/lib/zookeeper
clientPort=$ZK_CLIENT_PORT
server.1=${IP[0]}:2888:3888
server.2=${IP[1]}:2888:3888
server.3=${IP[2]}:2888:3888' | tee ./zookeeper/conf/zoo.cfg

#project conf
cd ./go/src
sudo rm -rf progettoSDCC
git clone git@github.com:cesto93/progettoSDCC
cd ./progettoSDCC
go build -o ./source/application/word_counter/worker/worker ./source/application/word_counter/worker/worker.go
go build -o ./source/application/word_counter/master/master ./source/application/word_counter/master/master.go
go build -o ./source/monitoring/agent ./source/monitoring/agent.go

#configuration of parameters
echo ${ID[@]} | jq -s 'add | flatten' | tee ./configuration/ec2_inst.json
echo ${IP[@]} | jq -s --arg port $ZK_CLIENT_PORT ' add |  flatten |[.[] + $port] ' | tee ./configuration/zk_servers_addrs.json

" &

done
