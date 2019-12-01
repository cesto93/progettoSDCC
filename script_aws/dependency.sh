#!/bin/bash

KEY_POS="/home/pier/Desktop/progetto_sdcc/myKey.pem"
GIT_KEY_FILE="sdcc_git"
LOCAL_DIR="/home/pier/.ssh"
AWS_DIR="/home/ec2-user/.ssh"

#depency and project clone
for (( i=0; i<${#NAMES[@]}; i++ ));
do
	INST=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${NAMES[$i]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
	INST_DNS[$i]=$(echo $INST | jq -r '.[].PublicDnsName')
	echo -e "COPYING SSH KEY FOR GIT DEPLOYMENT TO ${INST_DNS[$i]} \n"
	scp  -q -o "StrictHostKeyChecking=no" -i $KEY_POS $LOCAL_DIR/$GIT_KEY_FILE  ec2-user@${INST_DNS[$i]}:$AWS_DIR
	scp -q -o "StrictHostKeyChecking=no" -i $KEY_POS $LOCAL_DIR/config  ec2-user@${INST_DNS[$i]}:$AWS_DIR
	konsole --new-tab --noclose -e ssh  -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@${INST_DNS[$i]} \
"
sudo yum update -y -q
sudo yum install git -y -q -e 0
sudo yum install golang -y -q -e 0
#sudo yum install jq -y -q -e 0
go get -u github.com/aws/aws-sdk-go
go get -u github.com/samuel/go-zookeeper/zk

cd ./go/src
sudo rm -rf progettoSDCC
git clone git@github.com:cesto93/progettoSDCC
mkdir -p ./configuration/generated
echo 'finished installing' 
" &
done

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
echo 'finished installing zk_server' 
" &
done

source ./configure_monitor.sh
source ./configure_app.sh
