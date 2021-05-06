package awsglobalcache

import (
	"strings"

	"github.com/go-redis/redis/v8"
)

type Operation uint8

const (
	Read Operation = iota
	Write
)

type Configuration struct {
	masterNodeAWSRegion       AWSRegion
	localEnvironmentAWSRegion AWSRegion
	writer                    *redis.Client
	readers                   mappedRedisRegions
}

type ReaderRegionMapper struct {
	Region string
	*redis.Client
}

type AWSRegion uint64

const (
	USEast1 AWSRegion = iota
	USWest1
	EUCentral1
	Apac
)

type mappedRedisRegions map[AWSRegion]*redis.Client

func NewConfiguration(awsRegion string, writer *redis.Client, readers ...ReaderRegionMapper) *Configuration {
	readerMap := make(mappedRedisRegions)
	currentRegion := strings.ToLower(awsRegion)
	for _, r := range readers {
		localRegion := strings.ToLower(awsRegion)
		readerMap[castRegion(localRegion)] = r.Client
	}
	return &Configuration{
		masterNodeAWSRegion:       USEast1, // defaulting this for my use case
		localEnvironmentAWSRegion: castRegion(currentRegion),
		writer:                    writer,
		readers:                   readerMap,
	}
}

func castRegion(awsRegion string) AWSRegion {
	var region AWSRegion
	switch awsRegion {
	case "us-east-1", "us-east-2":
		region = USEast1
	case "us-west-1", "us-west-2":
		region = USWest1
	case "ap-south-1", "  ap-southeast-1", "ap-southeast-2":
		region = Apac
	case "eu-central-1":
		region = EUCentral1
	default:
		WarningLogger.Printf("The AWS Region %s does not have a proper mapping, falling back to us-east-1", awsRegion)
		region = USEast1
	}
	return region
}

// RetrieveRedisClient provides a simple wrapper that pick the redis client based on region and whether or not it's the operation you
// intend to do is a write or read operation, this is because the AWS Global Cache is set up as a “active-passive” database
// meaning that there is only 1 writer and many readers
func (c *Configuration) RetrieveRedisClient(operation Operation) *redis.Client {
	if c.localEnvironmentAWSRegion == c.masterNodeAWSRegion {
		return c.writer
	}

	switch operation {
	case Write:
		return c.writer
	case Read:
		if r, ok := c.readers[c.localEnvironmentAWSRegion]; ok {
			return r
		}
		return c.writer
	default:
		return c.writer
	}
}
