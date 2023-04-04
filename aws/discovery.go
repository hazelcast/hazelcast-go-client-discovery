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
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	hzdiscovery "github.com/hazelcast/hazelcast-go-client/cluster/discovery"
	"github.com/hazelcast/hazelcast-go-client/logger"

	"github.com/hazelcast/hazelcast-go-client-discovery"
)

type EC2DiscoveryStrategy struct {
	client      *client
	filters     []types.Filter
	portRange   cluster.PortRange
	logger      logger.Logger
	usePublicIP bool
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
	ds.usePublicIP = opts.UsePublicIP
	ds.logger = opts.Logger
	ds.client.logger = ds.logger
	ds.logInfof("Started EC2 Discovery Strategy %s", discovery.Version)
	return nil
}

func (ds *EC2DiscoveryStrategy) DiscoverNodes(ctx context.Context) ([]hzdiscovery.Node, error) {
	ds.logTrace(func() string {
		return "aws.EC2DiscoveryStrategy.DiscoverNodes"
	})
	iss, err := ds.client.GetInstances(ctx, ds.filters...)
	if err != nil {
		err = fmt.Errorf("discovering instances: %w", err)
		ds.logError(err)
		return nil, err
	}
	var nodes []hzdiscovery.Node
	for port := ds.portRange.Min; port <= ds.portRange.Max; port++ {
		for _, is := range iss {
			p := strconv.Itoa(port)
			node := hzdiscovery.Node{
				PrivateAddr: is.PrivateIP + ":" + p,
			}
			if ds.usePublicIP {
				node.PublicAddr = is.PublicIP + ":" + p
			}
			nodes = append(nodes, node)
		}
	}
	ds.logDebug(func() string {
		var sb strings.Builder
		sb.WriteString("Discovered Nodes:\n")
		for i, node := range nodes {
			p := "-"
			if node.PublicAddr != "" {
				p = node.PublicAddr
			}
			sb.WriteString(fmt.Sprintf("%d. private: %s public: %s\n", i+1, node.PrivateAddr, p))
		}
		return sb.String()
	})
	return nodes, nil
}

func (ds *EC2DiscoveryStrategy) logError(err error) {
	ds.logger.Log(logger.WeightError, func() string {
		return err.Error()
	})
}

func (ds *EC2DiscoveryStrategy) logInfof(format string, args ...any) {
	ds.logger.Log(logger.WeightInfo, func() string {
		return fmt.Sprintf(format, args...)
	})
}

func (ds *EC2DiscoveryStrategy) logDebug(f func() string) {
	ds.logger.Log(logger.WeightDebug, f)
}

func (ds *EC2DiscoveryStrategy) logTrace(f func() string) {
	ds.logger.Log(logger.WeightTrace, f)
}
