#!/bin/bash

ZONE="us-central1-a"
CONF=$(<../configuration/word_count.json)

#importing configuration
GC_NAMES=( $(echo $CONF | jq -r '.gc[].name') )

#workers
for (( i=0; i<${#GC_NAMES[@]}; i++ ));
do

konsole --new-tab --noclose -e gcloud compute ssh --zone=$ZONE ${GC_NAMES[$i]} --command \
"
cd ./go/src/progettoSDCC/bin
./worker
" &
done
