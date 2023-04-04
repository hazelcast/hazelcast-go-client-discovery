/*
 * Copyright (c) 2023, Hazelcast, Inc. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License")
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package aws

import (
	"context"
	"fmt"
	"strings"

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
			instances = append(instances, i)
		}
		c.logger.Log(logger.WeightDebug, func() string {
			var sb strings.Builder
			sb.WriteString("EC2 Instances:\n")
			for i, inst := range instances {
				sb.WriteString(fmt.Sprintf("%d. ID: %s private: %s public: %s\n", i+1, inst.ID, inst.PrivateIP, inst.PublicIP))
			}
			return sb.String()
		})
	}
	return instances, nil
}
