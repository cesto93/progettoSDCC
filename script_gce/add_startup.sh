#!/bin/bash

source ./conf/key.sh
ZONE="us-central1-a"
CONF=$(<../configuration/word_count.json)
NAMES=( $(echo $CONF | jq -r '.gc[].name') )

for (( i=0; i<${#NAMES[@]}; i++ ));
do
gcloud compute instances add-metadata ${NAMES[$i]} --zone=$ZONE --metadata startup-script=\
"
sudo $HOME_DIR/zookeeper/bin/zkServer.sh start
sleep 5
cd $HOME_DIR/go/src/progettoSDCC/bin
./agent
"
done
