#!/bin/bash
#create google compute vm
#argument: instance name

PROJ="concise-faculty-246814"
ZONE="us-central1-a"
SERV="wordcount-service-account"
MTYPE="f1-micro"
HDD="10GB"

gcloud beta compute --project=$PROJ instances \
create $1 --zone=$ZONE --machine-type=$MTYPE \
--subnet=default --network-tier=PREMIUM --maintenance-policy=MIGRATE \
--service-account=$SERV@$MTYPE.iam.gserviceaccount.com \
--scopes=https://www.googleapis.com/auth/cloud-platform --tags=http-server,https-server \
--image=debian-9-stretch-v20191014 --image-project=debian-cloud --boot-disk-size=$HDD \
--boot-disk-type=pd-standard --boot-disk-device-name=$1 --reservation-affinity=any
