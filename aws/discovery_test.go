package aws

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	"github.com/hazelcast/hazelcast-go-client/cluster/discovery"
	"github.com/hazelcast/hazelcast-go-client/logger"
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

type MockLogger struct{}

func (m MockLogger) Log(logger.Weight, func() string) {
	// pass
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
	must(st.Start(ctx, discovery.StrategyOptions{
		Logger:      &MockLogger{},
		UsePublicIP: false,
	}))
	nodes := mustValue(st.DiscoverNodes(ctx))
	target := []discovery.Node{
		{PrivateAddr: "172.31.20.2:6000"},
		{PrivateAddr: "172.31.18.15:6000"},
		{PrivateAddr: "172.31.20.2:6001"},
		{PrivateAddr: "172.31.18.15:6001"},
	}
	if !reflect.DeepEqual(target, nodes) {
		t.Fatalf("\n%v\n!=\n%v", target, nodes)
	}
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
	if !reflect.DeepEqual(target, nodes) {
		t.Fatalf("\n%v\n!=\n%v", target, nodes)
	}
}

func TestDiscovery_Failure(t *testing.T) {
	ctx := context.Background()
	mc := &MockEC2Client{
		Err: errors.New("discovery failed"),
	}
	cfg := Config{ec2Client: mc}
	st := mustValue(NewEC2DiscoveryStrategy(cfg))
	must(st.Start(ctx, discovery.StrategyOptions{
		Logger:      &MockLogger{},
		UsePublicIP: false,
	}))
	_, err := st.DiscoverNodes(ctx)
	if err == nil {
		t.Logf("should have failed")
	}
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
