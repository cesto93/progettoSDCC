#!/bin/bash
#create google compute vm
#argument: instance name

gcloud beta compute --project=concise-faculty-246814 instances \
create $1 --zone=us-central1-a --machine-type=f1-micro \
--subnet=default --network-tier=PREMIUM --maintenance-policy=MIGRATE \
--service-account=wordcount-service-account@concise-faculty-246814.iam.gserviceaccount.com \
--scopes=https://www.googleapis.com/auth/cloud-platform --tags=http-server,https-server \
--image=debian-9-stretch-v20191014 --image-project=debian-cloud --boot-disk-size=10GB \
--boot-disk-type=pd-standard --boot-disk-device-name=$1 --reservation-affinity=any
