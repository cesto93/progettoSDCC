#!/bin/bash
#set up environment
#argument: instance name

ZONE="us-central1-a"
GROUP="instance-group-1"

echo "Generating VM instance..."
sh create_vm.sh $1 2> /dev/null
echo "Integration in instances group..."
gcloud compute instance-groups unmanaged add-instances $GROUP --zone=$ZONE \
--instances $1 1> /dev/null 2> /dev/null
echo "Connecting..."
#gcloud compute scp --zone=$ZONE google_monitor.go $1:~ 1> /dev/null 2> /dev/null
gcloud compute scp --zone=$ZONE initialize_instance.sh $1:~ 1> /dev/null 2> /dev/null
#gcloud compute scp --zone=$ZONE ../wordcount.go $1:~ 1> /dev/null 2> /dev/null
echo "Initializing..."
gcloud compute ssh --zone=$ZONE $1 \
--command "sh initialize_instance.sh && rm initialize_instance.sh" && echo "Setup done"
gcloud compute ssh --zone=$ZONE $1
