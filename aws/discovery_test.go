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
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	"github.com/hazelcast/hazelcast-go-client/cluster/discovery"
	"github.com/hazelcast/hazelcast-go-client/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockEC2Client struct {
	Err error
}

func (m MockEC2Client) DescribeInstances(context.Context, *ec2.DescribeInstancesInput, ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	out := &ec2.DescribeInstancesOutput{}
	out.Reservations = []types.Reservation{
		{
			Instances: []types.Instance{
				{
					InstanceId:       nullable("i-01475d6ea60de2207"),
					PrivateIpAddress: nullable("172.31.20.2"),
					PublicIpAddress:  nullable("18.117.130.41"),
					// running
					State: &types.InstanceState{Code: nullable(int32(0xab00 | 16))},
				},
				{
					InstanceId:       nullable("i-01b835aa71f8f5405"),
					PrivateIpAddress: nullable("172.31.18.15"),
					PublicIpAddress:  nullable("3.142.208.176"),
					// running
					State: &types.InstanceState{Code: nullable(int32(0xab00 | 16))},
				},
				{
					InstanceId:       nullable("i-01d835aa71f8f5405"),
					PrivateIpAddress: nullable("172.31.18.16"),
					PublicIpAddress:  nullable("3.142.207.17"),
					// shutting down
					State: &types.InstanceState{Code: nullable(int32(0xab00 | 32))},
				},
			},
		},
	}
	return out, nil
}

type LogWeightMessage struct {
	Weight logger.Weight
	Msg    string
}

type MockLogger struct {
	LogItems []LogWeightMessage
}

func (m *MockLogger) Log(w logger.Weight, f func() string) {
	m.LogItems = append(m.LogItems, LogWeightMessage{
		Weight: w,
		Msg:    f(),
	})
}

func TestDiscovery_Success(t *testing.T) {
	ctx := context.Background()
	mc := &MockEC2Client{}
	cfg := Config{
		ec2Client: mc,
		PortRange: cluster.PortRange{
			Min: 6000,
			Max: 6001,
		},
	}
	cfg.SetFilters(Tag("App", "foo"))
	st := mustValue(NewEC2DiscoveryStrategy(cfg))
	lg := &MockLogger{}
	must(st.Start(ctx, discovery.StrategyOptions{
		Logger:      lg,
		UsePublicIP: false,
	}))
	nodes := mustValue(st.DiscoverNodes(ctx))
	target := []discovery.Node{
		{PrivateAddr: "172.31.20.2:6000"},
		{PrivateAddr: "172.31.18.15:6000"},
		{PrivateAddr: "172.31.20.2:6001"},
		{PrivateAddr: "172.31.18.15:6001"},
	}
	require.Equal(t, target, nodes)
	targetLogs := []LogWeightMessage{
		{Weight: 400, Msg: "Started EC2 Discovery Strategy 1.0.0"},
		{Weight: 600, Msg: "aws.EC2DiscoveryStrategy.DiscoverNodes"},
		{Weight: 600, Msg: "aws.client.GetInstances"},
		{Weight: 500, Msg: `EC2 Instances:
1. ID: i-01475d6ea60de2207 private: 172.31.20.2 public: 18.117.130.41
2. ID: i-01b835aa71f8f5405 private: 172.31.18.15 public: 3.142.208.176
`},
		{Weight: 500, Msg: `Discovered Nodes:
1. private: 172.31.20.2:6000 public: -
2. private: 172.31.18.15:6000 public: -
3. private: 172.31.20.2:6001 public: -
4. private: 172.31.18.15:6001 public: -
`},
	}
	assert.Equal(t, targetLogs, lg.LogItems)
}

func TestDiscovery_Success_UsePublicIP(t *testing.T) {
	ctx := context.Background()
	mc := &MockEC2Client{}
	cfg := Config{
		ec2Client: mc,
		PortRange: cluster.PortRange{
			Min: 6000,
			Max: 6001,
		},
	}
	cfg.SetFilters(Tag("App", "foo"))
	st := mustValue(NewEC2DiscoveryStrategy(cfg))
	must(st.Start(ctx, discovery.StrategyOptions{
		Logger:      &MockLogger{},
		UsePublicIP: true,
	}))
	nodes := mustValue(st.DiscoverNodes(ctx))
	target := []discovery.Node{
		{
			PrivateAddr: "172.31.20.2:6000",
			PublicAddr:  "18.117.130.41:6000",
		},
		{
			PrivateAddr: "172.31.18.15:6000",
			PublicAddr:  "3.142.208.176:6000",
		},
		{
			PrivateAddr: "172.31.20.2:6001",
			PublicAddr:  "18.117.130.41:6001",
		},
		{
			PrivateAddr: "172.31.18.15:6001",
			PublicAddr:  "3.142.208.176:6001",
		},
	}
	assert.Equal(t, target, nodes)
}

func TestDiscovery_Failure(t *testing.T) {
	ctx := context.Background()
	mc := &MockEC2Client{
		Err: errors.New("discovery failed"),
	}
	cfg := Config{ec2Client: mc}
	st := mustValue(NewEC2DiscoveryStrategy(cfg))
	lg := &MockLogger{}
	must(st.Start(ctx, discovery.StrategyOptions{
		Logger:      lg,
		UsePublicIP: false,
	}))
	_, err := st.DiscoverNodes(ctx)
	if err == nil {
		t.Logf("should have failed")
	}
	targetLog := []LogWeightMessage{
		{Weight: 400, Msg: "Started EC2 Discovery Strategy 1.0.0"},
		{Weight: 600, Msg: "aws.EC2DiscoveryStrategy.DiscoverNodes"},
		{Weight: 600, Msg: "aws.client.GetInstances"},
		{Weight: 200, Msg: "discovering instances: discovery failed"},
	}
	assert.Equal(t, targetLog, lg.LogItems)
}

func nullable[T any](v T) *T {
	return &v
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func mustValue[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
