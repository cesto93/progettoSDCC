#!/bin/bash

ZONE="us-central1-a"
source ./conf/key.sh

#echo "Generating VM instance..."
#./create_vm.sh $1

#echo "Integration in instances group..."
#gcloud compute instance-groups unmanaged add-instances $GROUP --zone=$ZONE --instances $1 -q

#echo "Connecting to $1..."
gcloud compute scp --zone=$ZONE initialize_instance.sh $1:~ -q
gcloud compute scp --zone=$ZONE $LOCAL_DIR/$GIT_KEY_FILE $USER@$1:$GC_DIR -q
gcloud compute scp --zone=$ZONE $LOCAL_DIR/config  $USER@$1:$GC_DIR -q
gcloud compute scp --zone=$ZONE ../configuration/generated/prometheus.yml  $USER@$1:$HOME_DIR -q
#gcloud compute ssh --zone=$ZONE $1 --command "sudo mv prometheus.yml /etc/prometheus"

echo "Initializing $1..."
gcloud compute ssh --zone=$ZONE $1 \
--command "bash initialize_instance.sh && rm initialize_instance.sh" && echo "Setup done"
#gcloud compute ssh --zone=$ZONE $1
