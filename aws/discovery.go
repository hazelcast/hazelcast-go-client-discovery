package aws

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	"github.com/hazelcast/hazelcast-go-client/cluster/discovery"
	"github.com/hazelcast/hazelcast-go-client/logger"
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

func (ds *EC2DiscoveryStrategy) Start(_ context.Context, opts discovery.StrategyOptions) error {
	ds.logger = opts.Logger
	ds.client.logger = ds.logger
	ds.debug(func() string {
		return "Started AWS Discovery Strategy"
	})
	return nil
}

func (ds *EC2DiscoveryStrategy) DiscoverNodes(ctx context.Context) ([]discovery.Node, error) {
	ds.debug(func() string {
		return "aws.EC2DiscoveryStrategy.DiscoverNodes"
	})
	iss, err := ds.client.GetInstances(ctx, ds.filters...)
	if err != nil {
		return nil, err
	}
	var nodes []discovery.Node
	for port := ds.portRange.Min; port <= ds.portRange.Max; port++ {
		for _, is := range iss {
			p := is.PublicIP + ":" + strconv.Itoa(port)
			pr := is.PrivateIP + ":" + strconv.Itoa(port)
			nodes = append(nodes, discovery.Node{
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
