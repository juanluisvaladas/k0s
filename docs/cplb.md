# Control plane load balancing

For clusters that don't have an [externally managed load balancer](high-availability.md#load-balancer) for the k0s
control plane, there is another option to get a highly available control plane called control plane load balancing (CPLB).

CPLB provides clusters a highly available virtual IP and load balancing for **external traffic**. For internal traffic
k0s provides [NLLB](nllb.md), both features are fully compatible and it's recommended to use both together.

Currently K0s relies on [keepalived](https://www.keepalived.org) for VIPs, which internally uses the
[VRRP protocol](https://datatracker.ietf.org/doc/html/rfc3768). Load Balancing can be done through either userspace
reverse proxy (recommended for simplicty) or it can use Keepalived's virtual servers feature which ultimately relies on IPVS.

## Technical limitations

CPLB is incompatible with:

1. Running as a [single node](k0s-single-node.md), i.e. it isn't started using the `--single` flag.
2. Specifying [`spec.api.externalAddress`][specapi].

Additionally, CPLB cannot be configured with dynamic configuration. Users with dynamic configuration
must set it up in the k0s configuration file on a per-control-plane node basis.
This an intentional design decision because the configuration may vary between control nodes.
For more information refer to [dynamic configuration](dynamic-configuration.md).

[specapi]: configuration.md#specapi

## Virtual IPs

### What is a VIP (virtual IP)

A virtual IP is an IP address that isn't tied to a single network interface,
instead it changes between multiple servers. This is a failover mechanism that
grants that there is always at least a functioning server and removes a single
point of failure.

### Configuring VIPs

CPLB relies internally on Keepalived's VRRP Instances. A VRRP Instance is a
server that will manage one or more VIPs. Most users will need exactly one
VRRP instance with exactly one VIP, however k0s allows multiple VRRP servers
with multiple VIPs for more advanced use cases.

A virtualIP requires:

1. A user-defined CIDR address which must e routable in the network
(normally in the same L3 network as the physical interface).
2. A user-defined auth pass.
3. A virtual router ID, which defaults to 51.
4. A network interface, if not defined k0s will chose the network interface
that owns the default route.

Except the network interface, all the other fields must be equal on every
control plane node.

**WARNING**: The authPass is a mechanism to prevent accidental conflicts
between VRRP instances that happen to be in the same network. It is not
encrypted and does not protect against malicious attacks.

This is a minimal example:

```yaml
spec:
  k0s:
    config:
      spec:
        network:
          controlPlaneLoadBalancing:
            enabled: true
            type: Keepalived
            keepalived:
              vrrpInstances:
              - virtualIPs: ["<External address IP>/<external address IP netmask>"]
                authPass: <password>
```

## Load Balancing

Currently k0s allows to chose one of two load balancing mechanism:

1. A userspace reverse proxy default and recommended.
2. For users who may need extra performance or more flexible algorithms,  
k0s can use the keepalived virtual servers load balancer feature.

Using keepalived in some nodes and userspace in other nodes is not supported
and has undefined behavior.

### Userspace load balancing

This is the default behavior, in order to enable it simple configure a VIP
using a VRRP instance.

```yaml
spec:
  k0s:
    config:
      spec:
        network:
          controlPlaneLoadBalancing:
            enabled: true
            type: Keepalived
            keepalived:
              vrrpInstances:
              - virtualIPs: ["<External address IP>/<external address IP netmask>"]
                authPass: <password>
```

### Keepalived Virtual Servers Load Balancing

The Keepalived virtual servers Load Balancing is more performant than the
userspace proxy, however it's not recommended because it has some drawbacks:

1. It's incompatible with controller+worker.
2. May not work on every infrastructure.
3. Troubleshooting is significantly more complex.

```yaml
spec:
  k0s:
    config:
      spec:
        network:
          controlPlaneLoadBalancing:
            enabled: true
            type: Keepalived
            keepalived:
              vrrpInstances:
              - virtualIPs: ["<External address IP>/<external address IP netmask>"]
                authPass: <password>
              virtualServers:
              - ipAddress: "<External ip address>"
```

## Full example using `k0sctl`

The following example shows a full `k0sctl` configuration file featuring three
controllers and three workers with control plane load balancing enabled.

```yaml
apiVersion: k0sctl.k0sproject.io/v1beta1
kind: Cluster
metadata:
  name: k0s-cluster
spec:
  hosts:
  - role: controller
    ssh:
      address: controller-0.k0s.lab
      user: root
      keyPath: ~/.ssh/id_rsa
    k0sBinaryPath: /opt/k0s
    uploadBinary: true
  - role: controller
    ssh:
      address: controller-1.k0s.lab
      user: root
      keyPath: ~/.ssh/id_rsa
    k0sBinaryPath: /opt/k0s
    uploadBinary: true
  - role: controller
    ssh:
      address: controller-2.k0s.lab
      user: root
      keyPath: ~/.ssh/id_rsa
    k0sBinaryPath: /opt/k0s
    uploadBinary: true
  - role: worker
    ssh:
      address: worker-0.k0s.lab
      user: root
      keyPath: ~/.ssh/id_rsa
    k0sBinaryPath: /opt/k0s
    uploadBinary: true
  - role: worker
    ssh:
      address: worker-1.k0s.lab
      user: root
      keyPath: ~/.ssh/id_rsa
    k0sBinaryPath: /opt/k0s
    uploadBinary: true
  - role: worker
    ssh:
      address: worker-2.k0s.lab
      user: root
      keyPath: ~/.ssh/id_rsa
    k0sBinaryPath: /opt/k0s
    uploadBinary: true
  k0s:
    version: v{{{ extra.k8s_version }}}+k0s.0
    config:
      spec:
        api:
          sans:
          - 192.168.122.200
        network:
          controlPlaneLoadBalancing:
            enabled: true
            type: Keepalived
            keepalived:
              vrrpInstances:
              - virtualIPs: ["192.168.122.200/24"]
                authPass: Example
          nodeLocalLoadBalancing: # optional, but CPLB will normally be used with NLLB.
            enabled: true
            type: EnvoyProxy
```

Save the above configuration into a file called `k0sctl.yaml` and apply it in
order to bootstrap the cluster:

```console
$ k0sctl apply
⠀⣿⣿⡇⠀⠀⢀⣴⣾⣿⠟⠁⢸⣿⣿⣿⣿⣿⣿⣿⡿⠛⠁⠀⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠀█████████ █████████ ███
⠀⣿⣿⡇⣠⣶⣿⡿⠋⠀⠀⠀⢸⣿⡇⠀⠀⠀⣠⠀⠀⢀⣠⡆⢸⣿⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀███          ███    ███
⠀⣿⣿⣿⣿⣟⠋⠀⠀⠀⠀⠀⢸⣿⡇⠀⢰⣾⣿⠀⠀⣿⣿⡇⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠀███          ███    ███
⠀⣿⣿⡏⠻⣿⣷⣤⡀⠀⠀⠀⠸⠛⠁⠀⠸⠋⠁⠀⠀⣿⣿⡇⠈⠉⠉⠉⠉⠉⠉⠉⠉⢹⣿⣿⠀███          ███    ███
⠀⣿⣿⡇⠀⠀⠙⢿⣿⣦⣀⠀⠀⠀⣠⣶⣶⣶⣶⣶⣶⣿⣿⡇⢰⣶⣶⣶⣶⣶⣶⣶⣶⣾⣿⣿⠀█████████    ███    ██████████
k0sctl  Copyright 2023, k0sctl authors.
Anonymized telemetry of usage will be sent to the authors.
By continuing to use k0sctl you agree to these terms:
https://k0sproject.io/licenses/eula
level=info msg="==> Running phase: Connect to hosts"
level=info msg="[ssh] worker-2.k0s.lab:22: connected"
level=info msg="[ssh] controller-2.k0s.lab:22: connected"
level=info msg="[ssh] worker-1.k0s.lab:22: connected"
level=info msg="[ssh] worker-0.k0s.lab:22: connected"
level=info msg="[ssh] controller-0.k0s.lab:22: connected"
level=info msg="[ssh] controller-1.k0s.lab:22: connected"
level=info msg="==> Running phase: Detect host operating systems"
level=info msg="[ssh] worker-2.k0s.lab:22: is running Fedora Linux 38 (Cloud Edition)"
level=info msg="[ssh] controller-2.k0s.lab:22: is running Fedora Linux 38 (Cloud Edition)"
level=info msg="[ssh] controller-0.k0s.lab:22: is running Fedora Linux 38 (Cloud Edition)"
level=info msg="[ssh] controller-1.k0s.lab:22: is running Fedora Linux 38 (Cloud Edition)"
level=info msg="[ssh] worker-0.k0s.lab:22: is running Fedora Linux 38 (Cloud Edition)"
level=info msg="[ssh] worker-1.k0s.lab:22: is running Fedora Linux 38 (Cloud Edition)"
level=info msg="==> Running phase: Acquire exclusive host lock"
level=info msg="==> Running phase: Prepare hosts"
level=info msg="==> Running phase: Gather host facts"
level=info msg="[ssh] worker-2.k0s.lab:22: using worker-2.k0s.lab as hostname"
level=info msg="[ssh] controller-0.k0s.lab:22: using controller-0.k0s.lab as hostname"
level=info msg="[ssh] controller-2.k0s.lab:22: using controller-2.k0s.lab as hostname"
level=info msg="[ssh] controller-1.k0s.lab:22: using controller-1.k0s.lab as hostname"
level=info msg="[ssh] worker-1.k0s.lab:22: using worker-1.k0s.lab as hostname"
level=info msg="[ssh] worker-0.k0s.lab:22: using worker-0.k0s.lab as hostname"
level=info msg="[ssh] worker-2.k0s.lab:22: discovered eth0 as private interface"
level=info msg="[ssh] controller-0.k0s.lab:22: discovered eth0 as private interface"
level=info msg="[ssh] controller-2.k0s.lab:22: discovered eth0 as private interface"
level=info msg="[ssh] controller-1.k0s.lab:22: discovered eth0 as private interface"
level=info msg="[ssh] worker-1.k0s.lab:22: discovered eth0 as private interface"
level=info msg="[ssh] worker-0.k0s.lab:22: discovered eth0 as private interface"
level=info msg="[ssh] worker-2.k0s.lab:22: discovered 192.168.122.210 as private address"
level=info msg="[ssh] controller-0.k0s.lab:22: discovered 192.168.122.37 as private address"
level=info msg="[ssh] controller-2.k0s.lab:22: discovered 192.168.122.87 as private address"
level=info msg="[ssh] controller-1.k0s.lab:22: discovered 192.168.122.185 as private address"
level=info msg="[ssh] worker-1.k0s.lab:22: discovered 192.168.122.81 as private address"
level=info msg="[ssh] worker-0.k0s.lab:22: discovered 192.168.122.219 as private address"
level=info msg="==> Running phase: Validate hosts"
level=info msg="==> Running phase: Validate facts"
level=info msg="==> Running phase: Download k0s binaries to local host"
level=info msg="==> Running phase: Upload k0s binaries to hosts"
level=info msg="[ssh] controller-0.k0s.lab:22: uploading k0s binary from /opt/k0s"
level=info msg="[ssh] controller-2.k0s.lab:22: uploading k0s binary from /opt/k0s"
level=info msg="[ssh] worker-0.k0s.lab:22: uploading k0s binary from /opt/k0s"
level=info msg="[ssh] controller-1.k0s.lab:22: uploading k0s binary from /opt/k0s"
level=info msg="[ssh] worker-1.k0s.lab:22: uploading k0s binary from /opt/k0s"
level=info msg="[ssh] worker-2.k0s.lab:22: uploading k0s binary from /opt/k0s"
level=info msg="==> Running phase: Install k0s binaries on hosts"
level=info msg="[ssh] controller-0.k0s.lab:22: validating configuration"
level=info msg="[ssh] controller-1.k0s.lab:22: validating configuration"
level=info msg="[ssh] controller-2.k0s.lab:22: validating configuration"
level=info msg="==> Running phase: Configure k0s"
level=info msg="[ssh] controller-0.k0s.lab:22: installing new configuration"
level=info msg="[ssh] controller-2.k0s.lab:22: installing new configuration"
level=info msg="[ssh] controller-1.k0s.lab:22: installing new configuration"
level=info msg="==> Running phase: Initialize the k0s cluster"
level=info msg="[ssh] controller-0.k0s.lab:22: installing k0s controller"
level=info msg="[ssh] controller-0.k0s.lab:22: waiting for the k0s service to start"
level=info msg="[ssh] controller-0.k0s.lab:22: waiting for kubernetes api to respond"
level=info msg="==> Running phase: Install controllers"
level=info msg="[ssh] controller-2.k0s.lab:22: validating api connection to https://192.168.122.200:6443"
level=info msg="[ssh] controller-1.k0s.lab:22: validating api connection to https://192.168.122.200:6443"
level=info msg="[ssh] controller-0.k0s.lab:22: generating token"
level=info msg="[ssh] controller-1.k0s.lab:22: writing join token"
level=info msg="[ssh] controller-1.k0s.lab:22: installing k0s controller"
level=info msg="[ssh] controller-1.k0s.lab:22: starting service"
level=info msg="[ssh] controller-1.k0s.lab:22: waiting for the k0s service to start"
level=info msg="[ssh] controller-1.k0s.lab:22: waiting for kubernetes api to respond"
level=info msg="[ssh] controller-0.k0s.lab:22: generating token"
level=info msg="[ssh] controller-2.k0s.lab:22: writing join token"
level=info msg="[ssh] controller-2.k0s.lab:22: installing k0s controller"
level=info msg="[ssh] controller-2.k0s.lab:22: starting service"
level=info msg="[ssh] controller-2.k0s.lab:22: waiting for the k0s service to start"
level=info msg="[ssh] controller-2.k0s.lab:22: waiting for kubernetes api to respond"
level=info msg="==> Running phase: Install workers"
level=info msg="[ssh] worker-2.k0s.lab:22: validating api connection to https://192.168.122.200:6443"
level=info msg="[ssh] worker-1.k0s.lab:22: validating api connection to https://192.168.122.200:6443"
level=info msg="[ssh] worker-0.k0s.lab:22: validating api connection to https://192.168.122.200:6443"
level=info msg="[ssh] controller-0.k0s.lab:22: generating a join token for worker 1"
level=info msg="[ssh] controller-0.k0s.lab:22: generating a join token for worker 2"
level=info msg="[ssh] controller-0.k0s.lab:22: generating a join token for worker 3"
level=info msg="[ssh] worker-2.k0s.lab:22: writing join token"
level=info msg="[ssh] worker-0.k0s.lab:22: writing join token"
level=info msg="[ssh] worker-1.k0s.lab:22: writing join token"
level=info msg="[ssh] worker-2.k0s.lab:22: installing k0s worker"
level=info msg="[ssh] worker-1.k0s.lab:22: installing k0s worker"
level=info msg="[ssh] worker-0.k0s.lab:22: installing k0s worker"
level=info msg="[ssh] worker-2.k0s.lab:22: starting service"
level=info msg="[ssh] worker-1.k0s.lab:22: starting service"
level=info msg="[ssh] worker-0.k0s.lab:22: starting service"
level=info msg="[ssh] worker-2.k0s.lab:22: waiting for node to become ready"
level=info msg="[ssh] worker-0.k0s.lab:22: waiting for node to become ready"
level=info msg="[ssh] worker-1.k0s.lab:22: waiting for node to become ready"
level=info msg="==> Running phase: Release exclusive host lock"
level=info msg="==> Running phase: Disconnect from hosts"
level=info msg="==> Finished in 2m20s"
level=info msg="k0s cluster version v{{{ extra.k8s_version }}}+k0s.0  is now installed"
level=info msg="Tip: To access the cluster you can now fetch the admin kubeconfig using:"
level=info msg="     k0sctl kubeconfig"
```

The cluster with the two nodes should be available by now. Setup the kubeconfig
file in order to interact with it:

```shell
k0sctl kubeconfig > k0s-kubeconfig
export KUBECONFIG=$(pwd)/k0s-kubeconfig
```

All three worker nodes are ready:

```console
$ kubectl get nodes
NAME                   STATUS   ROLES           AGE     VERSION
worker-0.k0s.lab       Ready    <none>          8m51s   v{{{ extra.k8s_version }}}+k0s
worker-1.k0s.lab       Ready    <none>          8m51s   v{{{ extra.k8s_version }}}+k0s
worker-2.k0s.lab       Ready    <none>          8m51s   v{{{ extra.k8s_version }}}+k0s
```

Each controller node has a dummy interface with the VIP and /32 netmask,
but only one has it in the real nic:

```console
$ for i in controller-{0..2} ; do echo $i ; ssh $i -- ip -4 --oneline addr show | grep -e eth0 -e dummyvip0; done
controller-0
2: eth0    inet 192.168.122.37/24 brd 192.168.122.255 scope global dynamic noprefixroute eth0\       valid_lft 2381sec preferred_lft 2381sec
2: eth0    inet 192.168.122.200/24 scope global secondary eth0\       valid_lft forever preferred_lft forever
3: dummyvip0    inet 192.168.122.200/32 scope global dummyvip0\       valid_lft forever preferred_lft forever
controller-1
2: eth0    inet 192.168.122.185/24 brd 192.168.122.255 scope global dynamic noprefixroute eth0\       valid_lft 2390sec preferred_lft 2390sec
3: dummyvip0    inet 192.168.122.200/32 scope global dummyvip0\       valid_lft forever preferred_lft forever
controller-2
2: eth0    inet 192.168.122.87/24 brd 192.168.122.255 scope global dynamic noprefixroute eth0\       valid_lft 2399sec preferred_lft 2399sec
3: dummyvip0    inet 192.168.122.200/32 scope global dummyvip0\       valid_lft forever preferred_lft forever
```

The cluster is using control plane load balancing and is able to tolerate the
outage of one controller node. Shutdown the first controller to simulate a
failure condition:

```console
$ ssh controller-0 'sudo poweroff'
Connection to 192.168.122.37 closed by remote host.
```

Control plane load balancing provides high availability, the VIP will have moved to a different node:

```console
$ for i in controller-{0..2} ; do echo $i ; ssh $i -- ip -4 --oneline addr show | grep -e eth0 -e dummyvip0; done
controller-1
2: eth0    inet 192.168.122.185/24 brd 192.168.122.255 scope global dynamic noprefixroute eth0\       valid_lft 2173sec preferred_lft 2173sec
2: eth0    inet 192.168.122.200/24 scope global secondary eth0\       valid_lft forever preferred_lft forever
3: dummyvip0    inet 192.168.122.200/32 scope global dummyvip0\       valid_lft forever preferred_lft forever
controller-2
2: eth0    inet 192.168.122.87/24 brd 192.168.122.255 scope global dynamic noprefixroute eth0\       valid_lft 2182sec preferred_lft 2182sec
3: dummyvip0    inet 192.168.122.200/32 scope global dummyvip0\       valid_lft forever preferred_lft forever

$ for i in controller-{0..2} ; do echo $i ; ipvsadm --save -n; done
IP Virtual Server version 1.2.1 (size=4096)
Prot LocalAddress:Port Scheduler Flags
  -> RemoteAddress:Port           Forward Weight ActiveConn InActConn
TCP  192.168.122.200:6443 rr persistent 360
  -> 192.168.122.185:6443              Route   1      0          0
  -> 192.168.122.87:6443               Route   1      0          0
  -> 192.168.122.122:6443              Route   1      0          0
````

And the cluster will be working normally:

```console
$ kubectl get nodes
NAME                   STATUS   ROLES           AGE     VERSION
worker-0.k0s.lab       Ready    <none>          8m51s   v{{{ extra.k8s_version }}}+k0s
worker-1.k0s.lab       Ready    <none>          8m51s   v{{{ extra.k8s_version }}}+k0s
worker-2.k0s.lab       Ready    <none>          8m51s   v{{{ extra.k8s_version }}}+k0s
```

## Troubleshooting

Although Virtual IPs work together and are closely related, these are two independent
processes.

### Troubleshooting Virtual IPs

The first thing to check is that the VIP is present in exactly one node at a time,
for instance if a cluster has an `172.17.0.102/16` address and the interface is `eth0`,
the expected output is similar to:

```shell
controller0:/# ip a s eth0
53: eth0@if54: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue state UP
    link/ether 02:42:ac:11:00:02 brd ff:ff:ff:ff:ff:ff
    inet 172.17.0.2/16 brd 172.17.255.255 scope global eth0
       valid_lft forever preferred_lft forever
    inet 172.17.0.102/16 scope global secondary eth0
       valid_lft forever preferred_lft forever
```

```shell
controller1:/# ip a s eth0
55: eth0@if56: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue state UP
    link/ether 02:42:ac:11:00:03 brd ff:ff:ff:ff:ff:ff
    inet 172.17.0.3/16 brd 172.17.255.255 scope global eth0
       valid_lft forever preferred_lft forever
```

If the virtualServers feature is used, there must be a dummy interface on the node
called `dummyvip0` which has the VIP but with `32` netmask. This isn't the VIP and
has to be there even if the VIP is held by another node.

```shell
controller0:/# ip a s dummyvip0 | grep 172.17.0.102
    inet 172.17.0.102/32 scope global dummyvip0
```

```shell
controller1:/# ip a s dummyvip0 | grep 172.17.0.102
    inet 172.17.0.102/32 scope global dummyvip0
```

If this isn't present in the nodes, keepalived logs can be seen in the k0s-logs, and
can be filtered with `component=keepalived`.

```shell
controller0:/# journalctl -u k0scontroller | grep component=keepalived
time="2024-11-19 12:56:11" level=info msg="Starting to supervise" component=keepalived
time="2024-11-19 12:56:11" level=info msg="Started successfully, go nuts pid 409" component=keepalived
time="2024-11-19 12:56:11" level=info msg="Tue Nov 19 12:56:11 2024: Starting Keepalived v2.2.8 (04/04,2023), git commit v2.2.7-154-g292b299e+" component=keepalived stream=stderr
[...]
```

The keepalived configuration is stored in a file called keepalived.conf in the k0s run
directory, by default `/run/k0s/keepalived.conf`, in this file there should be a
`vrrp_instance`section for each `vrrpInstance`.

Finally, k0s should have two keepalived processes running.

### Troubleshooting the Load Balancer

K0s has a component called `cplb-reconciler` responsible for setting the load balancer's
endpoint list, this reconciler monitors constantly the endpoint `kubernetes` in the
`default`namespace.

You can see the `cplb-reconciler` updates by running:

```shell
controller0:/# journalctl -u k0scontroller | grep component=cplb-reconciler
time="2024-11-20 20:29:28" level=error msg="Failed to watch API server endpoints, last observed version is \"\", starting over in 10s ..." component=cplb-reconciler error="Get \"https://172.17.0.6:6443/api/v1/namespaces/default/endpoints?fieldSelector=metadata.name%3Dkubernetes&timeout=30s&timeoutSeconds=30\": dial tcp 172.17.0.6:6443: connect: connection refused"
time="2024-11-20 20:29:38" level=info msg="Updated the list of IPs: [172.17.0.6]" component=cplb-reconciler
time="2024-11-20 20:29:55" level=info msg="Updated the list of IPs: [172.17.0.6 172.17.0.7]" component=cplb-reconciler
time="2024-11-20 20:29:59" level=info msg="Updated the list of IPs: [172.17.0.6 172.17.0.7 172.17.0.8]" component=cplb-reconciler
```

And verify the source of truth using kubectl:

```shell
controller0:/# kubectl get ep kubernetes -n default
NAME         ENDPOINTS                                         AGE
kubernetes   172.17.0.6:6443,172.17.0.7:6443,172.17.0.8:6443   9m14s
```

### Troubleshooting the Userspace Reverse Proxy

The userspace reverse proxy works on a separate socket, by default listens on port 6444.
It doesn't create any additional process and the requests to the APIserver port on the
VIP are forwarded to this socket using three iptables rules:

```shell
-A PREROUTING -d <VIP>/32 -p tcp -m tcp --dport <apiserver port> -j DNAT --to-destination 127.0.0.1:<userspace proxy port>
-A OUTPUT -d <VIP>/32 -p tcp -m tcp --dport <apiserver port> -j DNAT --to-destination 127.0.0.1:<userspace proxy port>
-A POSTROUTING -d 127.0.0.1/32 -p tcp -m tcp --dport <userspace proxy port> -j MASQUERADE
```

### Troubleshooting Keepalived Virtual Servers

You can verify the keepalived's logs and configuration file using the steps
described in the troubleshooting section for virtual IPs above. Additionally
you can check the actual IPVS configuration using `ipvsadm -L`.