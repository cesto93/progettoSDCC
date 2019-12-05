#!/bin/bash
source ./conf/key.sh
CONF=$(<../configuration/word_count.json)

#importing configuration
NAMES=( $(echo $CONF | jq -r '.aws[].name') )

#workers
for (( i=1; i<${#NAMES[@]}; i++ ));
do
INSTANCES[$i]=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${NAMES[$i]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
INSTANCE_DNS[$i]=$(echo ${INSTANCES[$i]} | jq -r '.[].PublicDnsName')
konsole --new-tab --noclose -e ssh -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@${INSTANCE_DNS[$i]} \
"
cd ./go/src/progettoSDCC/bin
./worker
" &
done

#master 
INSTANCES[0]=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${NAMES[0]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
INSTANCE_DNS[0]=$(echo ${INSTANCES[0]} | jq -r '.[].PublicDnsName')
konsole --new-tab --noclose -e ssh  -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@${INSTANCE_DNS[0]} \
"
cd ./go/src/progettoSDCC/bin
./master
" &


