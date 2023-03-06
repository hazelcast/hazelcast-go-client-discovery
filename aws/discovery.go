package aws

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	hzdiscovery "github.com/hazelcast/hazelcast-go-client/cluster/discovery"
	"github.com/hazelcast/hazelcast-go-client/logger"

	"github.com/hazelcast/hazelcast-go-client-discovery"
)

type EC2DiscoveryStrategy struct {
	client    *client
	filters   []types.Filter
	portRange cluster.PortRange
	logger    logger.Logger
}

// NewEC2DiscoveryStrategy creates a new EC2 discovery strategy.
func NewEC2DiscoveryStrategy(cfg Config) (*EC2DiscoveryStrategy, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	c := cfg.newEC2Client()
	fs := make([]types.Filter, len(cfg.Filters))
	copy(fs, cfg.Filters)
	return &EC2DiscoveryStrategy{
		client:    c,
		filters:   fs,
		portRange: cfg.PortRange,
	}, nil
}

func (ds *EC2DiscoveryStrategy) Start(_ context.Context, opts hzdiscovery.StrategyOptions) error {
	ds.logger = opts.Logger
	ds.client.logger = ds.logger
	ds.debug(func() string {
		return fmt.Sprintf("Started AWS Discovery Strategy %s", discovery.Version)
	})
	return nil
}

func (ds *EC2DiscoveryStrategy) DiscoverNodes(ctx context.Context) ([]hzdiscovery.Node, error) {
	ds.debug(func() string {
		return "aws.EC2DiscoveryStrategy.DiscoverNodes"
	})
	iss, err := ds.client.GetInstances(ctx, ds.filters...)
	if err != nil {
		return nil, err
	}
	var nodes []hzdiscovery.Node
	for port := ds.portRange.Min; port <= ds.portRange.Max; port++ {
		for _, is := range iss {
			p := is.PublicIP + ":" + strconv.Itoa(port)
			pr := is.PrivateIP + ":" + strconv.Itoa(port)
			nodes = append(nodes, hzdiscovery.Node{
				PublicAddr:  p,
				PrivateAddr: pr,
			})
		}
	}
	return nodes, nil
}

func (ds *EC2DiscoveryStrategy) debug(f func() string) {
	ds.logger.Log(logger.WeightDebug, f)
}
