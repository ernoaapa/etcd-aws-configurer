package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/crewjam/awsregion"
	"github.com/crewjam/ec2cluster"
)

func resolveCurrentEC2Instances() (*ec2.Instance, []*ec2.Instance, error) {
	cluster, err := resolveCluster()
	if err != nil {
		return nil, nil, fmt.Errorf("Error resolving EC2 cluster: %s", err)
	}

	localInstance, err := cluster.Instance()
	if err != nil {
		return nil, nil, fmt.Errorf("Error resolving local instance of EC2: %s", err)
	}

	allInstances, err := cluster.Members()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to get EC2 Cluster members: %s", err)
	}

	runningInstances := []*ec2.Instance{}
	for _, instance := range allInstances {
		if *instance.State.Name == ec2.InstanceStateNameRunning {
			runningInstances = append(runningInstances, instance)
		}
	}

	return localInstance, runningInstances, nil
}

func resolveCluster() (*ec2cluster.Cluster, error) {
	instanceID, err := ec2cluster.DiscoverInstanceID()
	if err != nil {
		return nil, err
	}

	awsSession := session.New()
	if region := os.Getenv("AWS_REGION"); region != "" {
		awsSession.Config.WithRegion(region)
	}
	awsregion.GuessRegion(awsSession.Config)

	cluster := &ec2cluster.Cluster{
		AwsSession: awsSession,
		InstanceID: instanceID,
		TagName:    "aws:autoscaling:groupName",
	}

	return cluster, nil
}
