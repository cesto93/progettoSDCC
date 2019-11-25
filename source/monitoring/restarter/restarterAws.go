package restarter

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws/session"
)

type AwsRestarter struct {
	client *ec2.EC2
}

func NewAws() *AwsRestarter {
	return &AwsRestarter{ec2.New(session.New())}
}

func (myRestarter *AwsRestarter) Restart(instanceId string) error {
	input := &ec2.RebootInstancesInput{
    	InstanceIds: []*string{
        	&instanceId,
    	},
	}
	_, err := myRestarter.client.RebootInstances(input)
	if err != nil {
    	return fmt.Errorf("failed to restart instance: %v", err)
	}
	return nil
}