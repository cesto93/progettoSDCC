#!/bin/bash
#create prometheus configuration from instances names

CONF=$(<../configuration/word_count.json)
NAMES=( $(echo $CONF | jq -r '.aws[].name') )
OUTPUT_JSON_FILE="./configuration/generated/instances.json"

#workers
for (( i=0; i<${#NAMES[@]}; i++ ));
do
ID[$i]=$(aws ec2 describe-instances --filters Name=tag:Name,Values=${NAMES[$i]}  --query 'Reservations[*].Instances[*]' | jq 'flatten')
INSTANCE_DNS[$i]=$(echo ${ID[$i]} | jq -r '.[].PublicDnsName')
IP[$i]=$(echo ${ID[$i]} | jq  -r '.[].NetworkInterfaces[].Association.PublicIp')
done
PROM_J=$(echo ${IP[@]}| jq -s '[{targets:[(.[] + ":9100")],labels:{job:"prometheus"}}]')

for (( i=0; i<${#NAMES[@]}; i++ ));
do
ssh -o "StrictHostKeyChecking=no" -i "$KEY_POS" ec2-user@${INSTANCE_DNS[$i]} \
"
cd ./go/src/progettoSDCC
echo '$PROM_J' | tee $OUTPUT_JSON_FILE
" &
done
