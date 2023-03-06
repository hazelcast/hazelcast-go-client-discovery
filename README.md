# Hazelcast Go Client Discovery Strategies

This package implements AWS EC2 discovery strategy for [Hazelcast Go Client](https://github.com/hazelcast/hazelcast-go-client).

## Requirements

* Hazelcast Go Client v1.4.0 or higher.
* Go 1.18

## Usage

### Create the AWS Configuration

```go
awsCfg, err := config.LoadDefaultConfig(ctx)
if err != nil {
    panic(err)
}
awsCfg.Region = "us-east-2"
```

Check out [Configuring the AWS SDK for Go V2](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/) for more information.

### Create the Discovery Configuration

```go
discoveryCfg := aws.Config{
    AWS: awsCfg,
    PortRange: cluster.PortRange{
        Min: 5701,
        Max: 5703,
    },
}
discoveryCfg.SetFilters(aws.Tag("App", "discovery"))
```

### Create the Discovery Strategy

```go
strategy, err := aws.NewEC2DiscoveryStrategy(discoveryCfg)
if err != nil {
    panic(err)
}
```

### Create the Hazelcast Go Client Configuration

```go
var cfg hazelcast.Config
cfg.Cluster.Discovery.Strategy = strategy
// set this if the client and members are not on the same network.
cfg.Cluster.Discovery.UsePublicIP = true
```

### Create and Start the Hazelcast Client Instance

```go
client, err := hazelcast.StartNewClientWithConfig(ctx, cfg)
if err != nil {
    panic(err)
}
```
