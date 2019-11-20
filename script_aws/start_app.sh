#!/bin/bash

KEY_POS="/home/pier/Desktop/progetto_sdcc/myKey.pem"
NAMES=( "master" "worker-1" "worker-2" "worker-3" )
PORTS=( "1050" "1051" "1052" "1053" )


#workers
for (( i=1; i<${#NAMES[@]}; i++ ));
do
INSTANCE_DNS=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${NAMES[$i]} --output text \
		--query 'Reservations[*].Instances[*].PublicDnsName')

IP[$i]=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${NAMES[$i]} --output text \
		--query 'Reservations[*].Instances[*].NetworkInterfaces[*].Association.PublicIp')

addresses[$i]="${IP[$i]}:${PORTS[$i]}"

echo "SSH CONNECTION TO ${NAMES[$i]} AT $INSTANCE_DNS"

konsole --new-tab --noclose -e ssh -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@$INSTANCE_DNS "
echo STARTING ${NAMES[$i]} AT RPC ${addresses[i]}
cd ./go/src/progettoSDCC/
./source/application/word_counter/worker/worker ${PORTS[$i]}
" &
done

#master
INSTANCE_DNS=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${NAMES[0]} --output text \
		--query 'Reservations[*].Instances[*].PublicDnsName')

IP[0]=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${NAMES[0]} --output text \
		--query 'Reservations[*].Instances[*].NetworkInterfaces[*].Association.PublicIp')
addresses[0]="${IP[0]}:${PORTS[0]}"

echo "SSH CONNECTION TO ${NAMES[0]} AT $INSTANCE_DNS"

konsole --new-tab --noclose -e ssh  -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@$INSTANCE_DNS "

echo STARTING MASTER AT RPC ${addresses[0]}
cd ./go/src/progettoSDCC/
./source/application/word_counter/master/master -workerAddr ${addresses[1]},${addresses[2]},${addresses[3]} -masterPort ${PORTS[0]}
" &

