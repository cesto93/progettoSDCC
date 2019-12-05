#!/bin/bash
#set up environment
#argument: instance name

ZONE="us-central1-a"
GROUP="instance-group-1"
source ./key.sh

echo "Generating VM instance..."
./create_vm.sh $1 #2> /dev/null

echo "Integration in instances group..."
gcloud compute instance-groups unmanaged add-instances $GROUP --zone=$ZONE \
--instances $1 1> /dev/null 2> /dev/null

echo "Connecting..."
#gcloud compute scp --zone=$ZONE google_monitor.go $1:~ 1> /dev/null 2> /dev/null
#gcloud compute scp --zone=$ZONE ../wordcount.go $1:~ 1> /dev/null 2> /dev/null
gcloud compute scp --zone=$ZONE initialize_instance.sh $1:~ 1> -q
gcloud compute scp --zone=$ZONE $LOCAL_DIR/$GIT_KEY_FILE $USER@$1:$GC_DIR -q
gcloud compute scp --zone=$ZONE $LOCAL_DIR/config  $USER@$1:$GC_DIR -q 

echo "Initializing..."
gcloud compute ssh --zone=$ZONE $1 \
--command "sh initialize_instance.sh && rm initialize_instance.sh" && echo "Setup done"
gcloud compute ssh --zone=$ZONE $1
