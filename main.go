package main

import (
	"flag"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func main() {
	targetFile := flag.String("target", "/etc/etcd_aws_configs.env", "the target file")
	flag.Parse()

	localInstance, instances, err := resolveCurrentEC2Instances()
	if err != nil {
		log.Fatalf("Failed to get EC2 Cluster information: %s", err)
	}
	log.Infof("Resolved %d instance EC2 cluster", len(instances))

	leader, members := resolveCurrentEtcdMembers(instances)

	initialClusterState := "new"
	if leader != nil {
		initialClusterState = "existing"
		log.Infoln("Resolved existing etcd cluster")
		for _, member := range members {
			log.Infof("%s: %s", member.Name, member.PeerURLs[0])
		}
		cleanupOldMembers(instances, leader, members)
		ensureIsMember(instances, localInstance, leader, members)
	} else {
		log.Infoln("New etcd cluster")
	}

	writeErr := writeSystemdEtcdConfig(*targetFile, initialClusterState, localInstance, instances)
	if writeErr != nil {
		log.Fatalf("Failed to write systemd file: %s", writeErr)
	}
}

func ensureIsMember(instances []*ec2.Instance, localInstance *ec2.Instance, leader *etcdMember, members []*etcdMember) {
	log.Infoln("Will join to existing etcd cluster")

	for _, existingMember := range members {
		if existingMember.Name == *localInstance.InstanceId {
			log.Infoln("Already part of the etcd cluster, skip registering")
			return
		}
	}

	localMember := &etcdMember{
		Name:     *localInstance.InstanceId,
		PeerURLs: []string{fmt.Sprintf("http://%s:2380", *localInstance.PrivateIpAddress)},
	}
	joinEtcdMember(leader, localMember)
}
