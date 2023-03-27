package aws

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type Filter func(cfg *Config) error

func Tag(name, value string) Filter {
	return func(c *Config) error {
		name = "tag:" + name
		f := types.Filter{
			Name:   &name,
			Values: []string{value},
		}
		c.Filters = append(c.Filters, f)
		return nil
	}
}
