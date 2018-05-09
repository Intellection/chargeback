package cloud

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cristim/ec2-instances-info"
)

type AWSEBSVolume struct {
	ID string
}

type AWSInstance struct {
	ID         string
	EBSVolumes []AWSEBSVolume
	Region     string
	Type       string
}

func GetAWSInstanceInfo(id string, region string) (AWSInstance, error) {
	var instance AWSInstance

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)
	ids := []*string{}
	ids = append(ids, aws.String(id))
	input := ec2.DescribeInstancesInput{InstanceIds: ids}

	result, err := svc.DescribeInstances(&input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}

		return instance, err
	}

	ebsVolumes := []AWSEBSVolume{}
	for _, ebsVolume := range result.Reservations[0].Instances[0].BlockDeviceMappings {
		ebsVolumes = append(ebsVolumes, AWSEBSVolume{ID: *ebsVolume.Ebs.VolumeId})
	}

	instance = AWSInstance{
		ID:         *result.Reservations[0].Instances[0].InstanceId,
		EBSVolumes: ebsVolumes,
		Region:     region,
		Type:       *result.Reservations[0].Instances[0].InstanceType,
	}

	fmt.Println(*result.Reservations[0].Instances[0])

	return instance, nil
}

func GetAWSInstanceOnDemandHourlyPrice(instance *AWSInstance) (float64, error) {
	var pricePerHour float64

	data, err := ec2instancesinfo.Data()
	if err != nil {
		return pricePerHour, err
	}

	for _, typeInfo := range *data {
		if typeInfo.InstanceType == instance.Type {
			for region, price := range typeInfo.Pricing {
				if region == instance.Region {
					pricePerHour = price.Linux.OnDemand
				}
			}
		}
	}

	return pricePerHour, nil
}
