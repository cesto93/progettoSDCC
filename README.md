# ProgettoSDCC

## Local Dependency
For launching the script you need to install

* jq
* konsole
* awscli
* gcloud

## AWS setup
For setup the instance to monitoring and wordcount application you need to:

*

## GC setup

## Application

###Commands

The client of the application use different flags for specifing different operations or differrent args

* This are the commands for operation specification
    * -load this command load the files specified in the AWS S3 bucket at specified names
    * -delete this command delete the files specified by names on the AWS S3 bucket
    * -list this command list the files on the bucket
    * -count this command do the wordcount of the files in the bucket identified by given names

* This are the commands for arg specification:
    * -bucket (OPTIONAL) specified to use a bucket that MUST be existent
    * -names (USED with load/delete/count)specifies the names for the S3 file in the bucket to use/load/delete
    * -paths (USED with load) specifies the paths for the local file to upload
    * -serverAddr (USED with count) specifie the address of the server for the rpc requiest to wordcount operation

###Configuration
The node configuration is done via a json file located at /configuration/word_count.json  
The ip of the nodes are gathered by the script files and parse in a json file at /configuration/generated/app_node.json.  
To test the app in local or if you want to manually set the ip use this file directly.  

##Monitoring