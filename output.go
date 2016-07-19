package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func writeSystemdEtcdConfig(filePath string, initialClusterState string, localInstance *ec2.Instance, instances []*ec2.Instance) error {

	basepath := path.Dir(filePath)
	if err := os.MkdirAll(basepath, 0666); err != nil {
		return err
	}

	fileHandle, fileCreateErr := os.Create(filePath)
	if fileCreateErr != nil {
		return fileCreateErr
	}

	writer := bufio.NewWriter(fileHandle)
	defer fileHandle.Close()

	fmt.Fprintln(writer, fmt.Sprintf("ETCD_NAME=%s", *localInstance.InstanceId))
	fmt.Fprintln(writer, fmt.Sprintf("ETCD_ADVERTISE_CLIENT_URLS=http://%s:2379", *localInstance.PrivateIpAddress))
	fmt.Fprintln(writer, fmt.Sprintf("ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379"))
	fmt.Fprintln(writer, fmt.Sprintf("ETCD_LISTEN_PEER_URLS=http://0.0.0.0:2380"))
	fmt.Fprintln(writer, fmt.Sprintf("ETCD_INITIAL_CLUSTER_STATE=%s", initialClusterState))
	fmt.Fprintln(writer, fmt.Sprintf("ETCD_INITIAL_CLUSTER=%s", buildInitialCluster(instances)))
	fmt.Fprintln(writer, fmt.Sprintf("ETCD_INITIAL_ADVERTISE_PEER_URLS=http://%s:2380", *localInstance.PrivateIpAddress))

	err := writer.Flush()
	if err != nil {
		return err
	}

	log.Infof("Wrote systemd configs to: %s", filePath)
	return nil
}

func buildInitialCluster(instances []*ec2.Instance) string {
	initialCluster := []string{}

	for _, instance := range instances {
		if instance.PrivateIpAddress == nil {
			continue
		}
		initialCluster = append(initialCluster, fmt.Sprintf("%s=http://%s:2380", *instance.InstanceId, *instance.PrivateIpAddress))
	}

	return strings.Join(initialCluster, ",")
}
