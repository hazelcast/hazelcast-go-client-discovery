package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	"github.com/hazelcast/hazelcast-go-client/hzerrors"
)

type Config struct {
	AWS       aws.Config
	Filters   []types.Filter
	PortRange cluster.PortRange
	err       error
	ec2Client ec2.DescribeInstancesAPIClient
}

func (c *Config) Validate() error {
	if c.err != nil {
		return c.err
	}
	if c.PortRange.Min <= 0 {
		c.PortRange.Min = 5701
	}
	if c.PortRange.Max <= 0 {
		c.PortRange.Max = 5703
	}
	if c.PortRange.Max < c.PortRange.Min {
		return fmt.Errorf("port range max should be greater or equal to %d: %w", c.PortRange.Min, hzerrors.ErrInvalidConfiguration)
	}
	return nil
}

func (c *Config) SetFilters(fs ...Filter) {
	for _, f := range fs {
		if err := f(c); err != nil {
			c.err = err
			break
		}
	}
}

func (c *Config) newEC2Client() *client {
	if c.ec2Client != nil {
		return newClient(c.ec2Client)
	}
	return newClient(ec2.NewFromConfig(c.AWS))
}
