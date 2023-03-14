package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	"github.com/hazelcast/hazelcast-go-client/logger"

	"github.com/hazelcast/hazelcast-go-client-discovery/aws"
)

func main() {
	ctx := context.Background()
	// Create the AWS Configuration
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}
	awsCfg.Region = "us-east-2"
	// Create the discovery configuration
	discoveryCfg := aws.Config{
		AWS: awsCfg,
		PortRange: cluster.PortRange{
			Min: 5701,
			Max: 5703,
		},
	}
	discoveryCfg.SetFilters(aws.Tag("App", "discovery"))
	// Create the discovery strategy
	strategy, err := aws.NewEC2DiscoveryStrategy(discoveryCfg)
	if err != nil {
		panic(err)
	}
	// Create the Hazelcast Go client configuration
	var cfg hazelcast.Config
	cfg.Cluster.Discovery.Strategy = strategy
	cfg.Cluster.Discovery.UsePublicIP = true
	cfg.Logger.Level = logger.DebugLevel
	// Start the client
	client, err := hazelcast.StartNewClientWithConfig(ctx, cfg)
	if err != nil {
		panic(fmt.Errorf("starting: %w", err))
	}
	// Do set/get operations on a map
	m, err := client.GetMap(ctx, "sample-map")
	if err != nil {
		panic(err)
	}
	for i := 0; i < 10; i++ {
		if err := m.Set(ctx, int64(i), int64(i)); err != nil {
			panic(fmt.Errorf("setting %d: %w", i, err))
		}
		v, err := m.Get(ctx, int64(i))
		if err != nil {
			panic(fmt.Errorf("getting %d: %w", i, err))
		}
		fmt.Println(i, v)
	}
	// That's all!
	if err := client.Shutdown(ctx); err != nil {
		panic(fmt.Errorf("shutting down: %w", err))
	}
}
