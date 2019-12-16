# ProgettoSDCC
This project consist on a monitoring system that observes AWS EC2 instances and GCE compute engine instances using cloudwatch, stackdriver and prometheus, as well as application exposed metrics. The metrics measured can be changed in the specific configuration files and others application metrics can be exported as long as they are saved as a specific json file.  
The system notice monitor agent fault (using zookeeper) and restarts the faulty intances. Finally metrics are aggregated on a master instance that use influxdb and connect to that to gather all metrics.  
We provide a test app that do the word count using aws EC2, S3 and GCE compute engine services.

## Local Dependency
In order to launch the script you need to install

* jq
* konsole
* awscli
* gcloud

## AWS setup
For setup the instance to monitoring and wordcount application you need to:

* Set a IAM role with:
    - ec2fullaccess 
    - s3fullaccess
    - cloudwatchfullaccess
* Set a security group with the inbound open ports:
    - 2888,3888,2181 (zookeeper)
    - 22 (ssh) 
    - 1050-1060 (rpc)
    - 9090, 9100 (prometheus)
* Create the instance and attach the IAM role and security group, also add names to them
* Set the metrics to measure by setting /configuration/metrics_ec2.json /configuration/metrics_s3.json
* launch script /script_aws/depency.sh to install depency

## GC setup
In order to setup instances running on google compute engine, you need to:

* launch the script setup_environment.sh in the folder progettoSDCC/script_gce, this will:
	- connect to each instance, 
	- install the necessary tools
* launch the script configure_monitor.sh in the folder progettoSDCC/script_gce, create some json files in the configuration folder of each instance:
	- a file containing the specified list (instances_names.json),
	- a file to let google know the IDs of instances to monitor (instances_ids.json),
	- a file to let prometheus know the IP's of instances to monitor (instances.json)

N.B.: 
-	by default only a subset of metrics are scraped by prometheus, gce and stackdriver; those metrics are specified in metrics_prometheus.json and metrics_gce.json, and can be expanded at will
-	ensure to oper ports specified in AWS SETUP on GCE instances too

## Monitoring

### Configuration
* The monitoring agent configuration is saved in the file /configuration/monitor.json
* The metrics monitored are taken from:
    * /configuration/metrics_ec2.json
    * /configuration/metrics_s3.json
    * /configuration/metrics_gce.json
    * /configuration/metrics_prometheus.json 
* The application metrics are taken from /log/app_metrics.json

* To configure aws monitoring run script ./script_aws/configure_monitoring.sh
* To configure gce monitoring run script ./script_gce/configure_monitoring.sh
* To add monitoring at startup on AWS run ./script_aws/add_statup.sh
* To add moitoring at startup on GCE run ./script_gce/add_statup.sh
* Set the grafana datasource on influxdb on http://ADDR:8086

### Usage
After configuration you should restart all instances, then monitoring is active

## Application

###Configuration
The node configuration is done via a json file located at /configuration/word_count.json
The bucket configuration is done via a json file located at /configuration/generated/bucket.json  
The ip of the nodes are gathered by the script files and parse in a json file at /configuration/generated/app_node.json.  
To test the app in local or if you want to manually set the ip use this file directly.  

###Commands

The client of the application uses different flags to specify different operations and arguments

* These are the commands for operations:
	- load: this command loads the files specified in the AWS S3 bucket at specified names
    - delete: this command deletes the files specified by names from the AWS S3 bucket
    - list: this command lists the files in the bucket
    - count: this command executes the wordcount of the files in the bucket identified by given names

* These are the commands for arguments specification:
    - names: (USED with load/delete/count) specifies the names for the S3 files in the bucket to use/load/delete
    - paths: (USED with load) specifies the paths for the local files to upload
    - serverAddr: (USED with count) specifies the address of the server for the rpc request
