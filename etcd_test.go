package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEtcdMemberMatch(t *testing.T) {
	member := &etcdMember{
		PeerURLs: []string{"http://172.20.29.227:2380"},
	}

	assert.True(t, member.match("172.20.29.227"))
	assert.False(t, member.match("172.20.19.48"))
}

func TestEtcdMemberMatchFallbackToClientURLs(t *testing.T) {
	member := &etcdMember{
		ClientURLs: []string{
			"http://172.20.29.227:2379",
			"http://172.20.29.227:4001",
		},
	}

	assert.True(t, member.match("172.20.29.227"))
	assert.False(t, member.match("172.20.19.48"))
}
