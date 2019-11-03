#!/bin/bash
#set up environment
#argument: instance name

echo "Generating VM instance..."
sh create_vm.sh $1 2> /dev/null
echo "Integration in instances group..."
gcloud compute instance-groups unmanaged add-instances instance-group-1 --zone=us-central1-a --instances $1 1> /dev/null 2> /dev/null
echo "Connecting..."
#gcloud compute scp --zone=us-central1-a google_monitor.go $1:~ 1> /dev/null 2> /dev/null
gcloud compute scp --zone=us-central1-a initialize_instance.sh $1:~ 1> /dev/null 2> /dev/null
#gcloud compute scp --zone=us-central1-a ../wordcount.go $1:~ 1> /dev/null 2> /dev/null
echo "Initializing..."
gcloud compute ssh --zone=us-central1-a $1 \
--command "sh initialize_instance.sh && rm initialize_instance.sh" && echo "Setup done"
gcloud compute ssh --zone=us-central1-a $1
