package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func cleanupOldMembers(clusterInstances []*ec2.Instance, leader *etcdMember, members []*etcdMember) error {
	membersToRemove := getMembersToRemove(clusterInstances, members)
	if len(membersToRemove) > 0 {
		log.Infof("Cleanup %d old members from ETCD cluster...", len(membersToRemove))
		for _, member := range membersToRemove {
			err := removeEtcdMember(leader, member)
			if err != nil {
				return err
			}
		}
	} else {
		log.Infoln("No old members found!")
	}

	return nil
}

func getMembersToRemove(clusterInstances []*ec2.Instance, members []*etcdMember) []*etcdMember {
	oldMembers := []*etcdMember{}

	for _, member := range members {
		if !containsMember(clusterInstances, member) {
			oldMembers = append(oldMembers, member)
		}
	}
	return oldMembers
}

func containsMember(clusterInstances []*ec2.Instance, member *etcdMember) bool {
	for _, instance := range clusterInstances {
		if member.match(*instance.PrivateIpAddress) {
			return true
		}
	}
	return false
}
