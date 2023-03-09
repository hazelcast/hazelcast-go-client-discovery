# Hazelcast Go Client Discovery Strategies

This package implements AWS EC2 discovery strategy for [Hazelcast Go Client](https://github.com/hazelcast/hazelcast-go-client).

## Requirements

* Hazelcast Go Client v1.4.0 or higher.
* Go 1.18
* The user or role which is used to run the discovery strategy code must have the `ec2:DescribeInstances` permission.  

## Usage

### Add the Package Dependency

Use the following `go.mod` file or incorporate the `require` and `replace` sections in your own `go.mod` file.
Note that the instructions in this section will change once the GA version is released. 

```
go 1.18

require (
    github.com/aws/aws-sdk-go-v2/config v1.18.15
    github.com/hazelcast/hazelcast-go-client v1.4.0
    github.com/hazelcast/hazelcast-go-client-discovery v0.0.0-20230309191213-822637b03020
)

replace github.com/hazelcast/hazelcast-go-client v1.4.0 => github.com/hazelcast/hazelcast-go-client v1.3.1-0.20230309185934-ce3e7d2b1ade
```

And run `go mod tidy`

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
