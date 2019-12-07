#!/bin/bash
#set up environment
#argument: instance name

ZONE="us-central1-a"
GROUP="instance-group-1"
source ./conf/key.sh

echo "Generating VM instance..."
./create_vm.sh $1

echo "Integration in instances group..."
gcloud compute instance-groups unmanaged add-instances $GROUP --zone=$ZONE --instances $1 -q

echo "Connecting..."
gcloud compute scp --zone=$ZONE initialize_instance.sh $1:~ -q
gcloud compute scp --zone=$ZONE $LOCAL_DIR/$GIT_KEY_FILE $USER@$1:$GC_DIR -q
gcloud compute scp --zone=$ZONE $LOCAL_DIR/config  $USER@$1:$GC_DIR -q 

# Generating Prometheus configuration file
tee ../configuration/generated/prometheus.yml<<EOF 1> /dev/null
# my global config
global:
  scrape_interval:     15s # Set the scrape interval to every 15 seconds. Default is every 1 minute.
  evaluation_interval: 15s # Evaluate rules every 15 seconds. The default is every 1 minute.
  # scrape_timeout is set to the global default (10s).

scrape_configs:
  - job_name: 'MonitorInstances'
    file_sd_configs:
      - files:
        - $HOME_DIR/go/src/progettoSDCC/configuration/generated/instances.json
EOF
gcloud compute scp --zone=$ZONE ../configuration/generated/prometheus.yml  $USER@$1:$HOME_DIR -q
#gcloud compute ssh --zone=$ZONE $1 --command "sudo mv prometheus.yml /etc/prometheus"

echo "Initializing..."
gcloud compute ssh --zone=$ZONE $1 \
--command "bash initialize_instance.sh && rm initialize_instance.sh" && echo "Setup done"
gcloud compute ssh --zone=$ZONE $1
