# vipd
vipd is an application that runs as a group and sets an IP alias based on group election.
It uses an "oldest peer" method for leader election where the oldest known node becomes the
leader.

# Usage
To use vipd, install the binary on each node that you want in the availability group.
Vipd uses a config file.  Create the following replacing with your values:

```toml
NodeName = "nodeA"
ClusterAddress = "10.10.0.10:7946"
Peers = ["10.10.0.11:7946", "10.10.0.12:7946"]
LeaderPromotionTimeout = "30s"

[[VirtualIPs]]
  Interface = "eth0"
  IPAddress = "10.10.255.2/16"
```

In this example, it expects 3 nodes.  Replace the `10.10.0.x` IPs with your own.  Once
each node is started you should see output similar to the following:

```
time="2021-12-28T17:01:12Z" level=debug msg="checking peer nodeB: 2021-12-28 15:42:35.282503973 +0000 UTC"
time="2021-12-28T17:01:12Z" level=debug msg="checking peer nodeC: 2021-12-28 16:11:02.558418013 +0000 UTC"
time="2021-12-28T17:01:12Z" level=debug msg="new leader: nodeA"
time="2021-12-28T17:01:12Z" level=debug msg="activating vip 10.10.255.2/16 on eth0"
```
