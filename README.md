# Hazelcast Go Client Discovery Strategies

This package implements AWS EC2 discovery strategy for [Hazelcast Go Client](https://github.com/hazelcast/hazelcast-go-client).

## Requirements

* Hazelcast Go Client v1.4.0 or higher.
* Go 1.18
* The user or role which is used to run the discovery strategy code must have the `ec2:DescribeInstances` permission.  

## Usage

### Add the Package Dependency

Add the following dependencies to your project.
Note that the instructions in this section will change once the GA version is released. 

```
$ go get github.com/aws/aws-sdk-go-v2/config@v1.18.15
$ go get go get github.com/hazelcast/hazelcast-go-client-discovery@b4bf756df29708606c737f1fdb4c384f5a41c004
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
// note that, in that case the members should be configured to provide public addresses.
cfg.Cluster.Discovery.UsePublicIP = true
```

If you need to enable `UsePublicIP`, see the following documentation to configure the members:

* `hazelcast.discovery.public.ip.enabled` system property at https://docs.hazelcast.com/hazelcast/5.2/system-properties#hide-nav
* https://docs.hazelcast.com/hazelcast/5.2/clusters/network-configuration#public-address
* https://docs.hazelcast.com/hazelcast/5.2/deploy/deploying-on-aws#ec2-configuration

### Create and Start the Hazelcast Client Instance

```go
client, err := hazelcast.StartNewClientWithConfig(ctx, cfg)
if err != nil {
    panic(err)
}
```
