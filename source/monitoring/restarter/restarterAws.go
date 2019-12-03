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
    	return fmt.Errorf("failed to restart instance: %v", err)
	}
	return nil
}

func (myRestarter *AwsRestarter) getState(instanceId string) (*ec2.InstanceState, error) {
	var state ec2.InstanceState
	input := &ec2.DescribeInstancesInput{
    	InstanceIds: []*string{
        	&instanceId,
    	},
	}

	resp, err := myRestarter.client.DescribeInstances(input)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance state : %v", err)
	}

	for idx, res := range resp.Reservations {
		fmt.Println("  > Reservation Id", *res.ReservationId, " Num Instances: ", len(res.Instances))
		for _, inst := range resp.Reservations[idx].Instances {
			fmt.Println("    - Instance ID: ", *inst.InstanceId)
			fmt.Println("    - Instance state: ", *inst.State)
			state = *inst.State
		}
	}
	return &state, nil
}

func (myRestarter *AwsRestarter) start(instanceId string) error {
	input := &ec2.StartInstancesInput{
    	InstanceIds: []*string{
        	&instanceId,
    	},
	}
	_, err := myRestarter.client.StartInstances(input)
	if err != nil {
    	return fmt.Errorf("failed to start instance: %v", err)
	}
	return nil
}

//TODO implements state recovery / app restart of monitoring
func (myRestarter *AwsRestarter) Restart(instanceId string) error {
	state, err := myRestarter.getState(instanceId)
	if err != nil {
		return err
	}
	fmt.Println(state)
	if state.Code == aws.Int64(16) {
		err = myRestarter.restart(instanceId)
	} else if state.Code == aws.Int64(80) {
		err = myRestarter.start(instanceId)

	}
	return err
}