# etcd-aws-configurer
Tool which resolves etcd configuration from AWS ELB information.

Tool can be used at server startup to configure etcd without hard
coding ip addresses or using the etcd discovery service.

## Command Line usage
```shell
./etcd-aws-configurer -target /etc/etcd_aws_configs.env
```

Creates file to `/etc/etcd_aws_configs.env` with content like:
```
ETCD_NAME=i-1a1b1c1d
ETCD_ADVERTISE_CLIENT_URLS=http://172.20.19.321:2379
ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379
ETCD_LISTEN_PEER_URLS=http://0.0.0.0:2380
ETCD_INITIAL_CLUSTER_STATE=new
ETCD_INITIAL_CLUSTER=i-3a3b3c3d=http://172.20.29.323:2380,i-2a2b2c2d=http://172.20.19.322:2380,i-1a1b1c1d=http://172.20.19.321:2380
ETCD_INITIAL_ADVERTISE_PEER_URLS=http://172.20.19.321:2380
```

`ETCD_INITIAL_CLUSTER_STATE` can be either `new` or `existing`.
The tool tries to call each node API and resolve is there leader already.
If it finds leader, `ETCD_INITIAL_CLUSTER_STATE=new` otherwise `existing`.


## CoreOS usage in EC2
Add following into your cloudinit
```yaml
coreos:
  units:
    - name: etcd-configure.service
      command: start
      content: |
        [Unit]
        Description=Configure etcd based on AWS ELB information
        Documentation=https://github.com/ernoaapa/etcd-aws-configurer
        Requires=network-online.target
        After=network-online.target

        [Service]
        ExecStartPre=/usr/bin/curl -s -L -o /opt/bin/etcd-aws-configurer https://github.com/ernoaapa/etcd-aws-configurer/releases/download/v0.1.0/etcd-aws-configurer-Linux-x86_64
        ExecStartPre=/usr/bin/chmod +x /opt/bin/etcd-aws-configurer
        ExecStart=/opt/bin/etcd-aws-configurer -target /etc/etcd_aws_configs.env
        RemainAfterExit=yes
        Type=oneshot

    - name: etcd2.service
      command: start
      drop-ins:
        - name: 30-use-aws-etcd-configs.conf
          content: |
            [Unit]
            After=etcd-configure.service
            Requires=etcd-configure.service

            [Service]
            EnvironmentFile=/etc/etcd_aws_configs.env

```

## Other similar tool
[crewjam/etcd-aws](https://github.com/crewjam/etcd-aws/) actually used this as reference implementation, but it did too much and I didn't want to wrap the etcd process.

[MonsantoCo/etcd-aws-cluster](https://github.com/MonsantoCo/etcd-aws-cluster) bash implementation which uses `awscli` but in CoreOS you cannot install it easily.
Also in my case, cannot run inside Docker container because at boot I don't have Docker yet running.
