package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type etcdState struct {
	Name       string         `json:"name"`
	ID         string         `json:"id"`
	State      string         `json:"state"`
	StartTime  time.Time      `json:"startTime"`
	LeaderInfo etcdLeaderInfo `json:"leaderInfo"`
}

type etcdLeaderInfo struct {
	Leader               string    `json:"leader"`
	Uptime               string    `json:"uptime"`
	StartTime            time.Time `json:"startTime"`
	RecvAppendRequestCnt int       `json:"recvAppendRequestCnt"`
	RecvPkgRate          int       `json:"recvPkgRate"`
	RecvBandwidthRate    int       `json:"recvBandwidthRate"`
	SendAppendRequestCnt int       `json:"sendAppendRequestCnt"`
}

type etcdMembers struct {
	Members []etcdMember `json:"members,omitempty"`
}

type etcdMember struct {
	ID         string   `json:"id,omitempty"`
	Name       string   `json:"name,omitempty"`
	PeerURLs   []string `json:"peerURLs,omitempty"`
	ClientURLs []string `json:"clientURLs,omitempty"`
}

func (m *etcdMember) match(instanceIP string) bool {
	for _, peerURL := range m.PeerURLs {
		if strings.Contains(peerURL, instanceIP) {
			return true
		}
	}
	for _, clientURL := range m.ClientURLs {
		if strings.Contains(clientURL, instanceIP) {
			return true
		}
	}
	return false
}

func resolveCurrentEtcdMembers(instances []*ec2.Instance) (*etcdMember, []*etcdMember) {
	for _, instance := range instances {

		// Get member state
		memberStateResponse := etcdState{}
		if err := httpGet(fmt.Sprintf("http://%s:2379/v2/stats/self", *instance.PrivateIpAddress), &memberStateResponse); err != nil {
			log.Infof("%s: http://%s:2379/v2/stats/self: %s", *instance.InstanceId, *instance.PrivateIpAddress, err)
			continue
		}

		// Fetch member list from the instance
		etcdMembersResponse := etcdMembers{}
		if err := httpGet(fmt.Sprintf("http://%s:2379/v2/members", *instance.PrivateIpAddress), &etcdMembersResponse); err != nil {
			log.Printf("%s: http://%s:2379/v2/members: %s", *instance.InstanceId, *instance.PrivateIpAddress, err)
			continue
		}

		members := []*etcdMember{}
		for _, member := range etcdMembersResponse.Members {
			currentMember := member
			members = append(members, &currentMember)
		}

		// Find out the leader
		leader := findMemberByID(memberStateResponse.LeaderInfo.Leader, members)
		if leader == nil {
			log.Printf("%s: http://%s:2379/v2/stats/self: alive, no leader", *instance.InstanceId, *instance.PrivateIpAddress)
			continue
		}

		// Resolved existing etcd cluster information
		return leader, members
	}

	// None of the instances replied, propbably brand new cluster
	return nil, nil
}

func removeEtcdMember(leader *etcdMember, member *etcdMember) error {
	log.Infof("Remove etcd member [%s]", member.Name)
	for _, clientURL := range leader.ClientURLs {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/v2/members/%s", clientURL, member.ID), nil)
		_, err := http.DefaultClient.Do(req)
		if err == nil {
			return nil
		}
		log.Warningf("Failed to call leader client url [DELETE %s], continue trying", clientURL)
	}

	return fmt.Errorf("Failed to remove etcd member [%s], failed to call all leader node client urls [%s]", member.Name, leader.Name)
}

func joinEtcdMember(leader *etcdMember, member *etcdMember) error {
	log.Infof("Join etcd member [%s] via [%s]", member.Name, leader.Name)
	body, _ := json.Marshal(member)

	for _, clientURL := range leader.ClientURLs {
		_, err := http.Post(fmt.Sprintf("%s/v2/members", clientURL), "application/json", bytes.NewReader(body))
		if err == nil {
			return nil
		}
		log.Warningf("Failed to call leader client url [POST %s], continue trying: %s", clientURL, err)
	}

	return fmt.Errorf("Failed to join etcd member [%s], failed to call all leader node client urls [%s]", member.Name, leader.Name)
}

// httpGet make GET request to the target url and decode response into the target interface
func httpGet(url string, target interface{}) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}

	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return err
	}
	return nil
}

func findMemberByID(id string, members []*etcdMember) *etcdMember {
	for _, member := range members {
		if member.ID == id {
			return member
		}
	}

	return nil
}
