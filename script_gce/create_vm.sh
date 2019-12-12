#!/bin/bash
#create google compute vm
#argument: instance name

source ./conf/project.sh
MTYPE="g1-small"
HDD="30GB"
#IMG="debian-9-stretch-v20191115"
#IMG_PROJ="debian-cloud"
IMG="ubuntu-1804-bionic-v20191113"
IMG_PROJ="ubuntu-os-cloud"

for i in $@;
do
	gcloud beta compute --project=$PROJ instances \
	create $i --zone=$ZONE --machine-type=$MTYPE \
	--subnet=default --network-tier=PREMIUM --maintenance-policy=MIGRATE \
	--service-account=$SERV@$PROJ.iam.gserviceaccount.com \
	--scopes=https://www.googleapis.com/auth/cloud-platform --tags=http-server,https-server \
	--image=$IMG --image-project=$IMG_PROJ --boot-disk-size=$HDD \
	--boot-disk-type=pd-standard --boot-disk-device-name=$i --reservation-affinity=any -q
done
