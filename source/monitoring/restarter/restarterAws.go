package restarter

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws/session"
)

const (
	awsRegion = "us-east-1"
)

type AwsRestarter struct {
	client *ec2.EC2
}

func NewAws() *AwsRestarter {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion), }))
	return &AwsRestarter{ec2.New(sess)}
}

func (myRestarter *AwsRestarter) restart(instanceId string) error {
	input := &ec2.RebootInstancesInput{
    	InstanceIds: []*string{
        	&instanceId,
    	},
	}
	_, err := myRestarter.client.RebootInstances(input)
	if err != nil {
    	return fmt.Errorf("failed to start instance: %v", err)
	}
	return nil
}

func (myRestarter *AwsRestarter) start(instanceId string) error {
	input := &ec2.StartInstancesInput{
    	InstanceIds: []*string{
        	&instanceId,
    	},
	}
	_, err := myRestarter.client.StartInstances(input)
	if err != nil {
    	return fmt.Errorf("failed to restart instance: %v", err)
	}
	return nil
}

//TODO implements state recovery / app restart of monitoring
func (myRestarter *AwsRestarter) Restart(instanceId string) error {

	err := myRestarter.restart(instanceId)
	if (err != nil) {
		err = myRestarter.start(instanceId)
	}
	return err
}