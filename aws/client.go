package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hazelcast/hazelcast-go-client/logger"
)

type instance struct {
	ID        string
	PublicIP  string
	PrivateIP string
}

func (is instance) String() string {
	return fmt.Sprintf("id:%s public:%s private:%s\n", is.ID, is.PublicIP, is.PrivateIP)
}

type client struct {
	ec2    ec2.DescribeInstancesAPIClient
	logger logger.Logger
}

func newClient(ec2Client ec2.DescribeInstancesAPIClient) *client {
	return &client{ec2: ec2Client}
}

func (c client) GetInstances(ctx context.Context, filters ...types.Filter) ([]instance, error) {
	c.logger.Log(logger.WeightTrace, func() string {
		return "aws.client.GetInstances"
	})
	inp := &ec2.DescribeInstancesInput{}
	if len(filters) > 0 {
		inp.Filters = filters
	}
	out, err := c.ec2.DescribeInstances(ctx, inp)
	if err != nil {
		return nil, err
	}
	var instances []instance
	for _, r := range out.Reservations {
		for _, in := range r.Instances {
			var state byte
			if in.State != nil {
				state = instanceState(in.State.Code)
			}
			if state != instanceRunning {
				continue
			}
			id := drefStr(in.InstanceId)
			pip := drefStr(in.PublicIpAddress)
			prip := drefStr(in.PrivateIpAddress)
			i := instance{
				ID:        id,
				PublicIP:  pip,
				PrivateIP: prip,
			}
			c.logger.Log(logger.WeightDebug, func() string {
				return fmt.Sprintf("EC2 instance %s", i.String())
			})
			instances = append(instances, i)
		}
	}
	return instances, nil
}
