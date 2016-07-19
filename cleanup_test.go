package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
)

func TestGetMembersToRemove(t *testing.T) {
	clusterInstances := []*ec2.Instance{
		&ec2.Instance{PrivateIpAddress: asPointerString("172.20.19.48")},
		&ec2.Instance{PrivateIpAddress: asPointerString("172.20.19.49")},
		&ec2.Instance{PrivateIpAddress: asPointerString("172.20.29.227")},
	}

	members := []*etcdMember{
		&etcdMember{Name: "172.20.19.48", PeerURLs: []string{"http://172.20.19.48:2380"}},
		&etcdMember{Name: "172.20.19.49", PeerURLs: []string{"http://172.20.19.49:2380"}},
		&etcdMember{Name: "172.20.29.205", PeerURLs: []string{"http://172.20.29.205:2380"}},
	}

	result := getMembersToRemove(clusterInstances, members)
	assert.Equal(t, 1, len(result), "Returned more than expected members to remove")
	assert.Equal(t, "172.20.29.205", result[0].Name, "Returned more than expected members to remove")
}

func asPointerString(str string) *string {
	return &str
}
