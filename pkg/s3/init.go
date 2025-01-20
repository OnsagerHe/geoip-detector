package s3

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Config struct {
	AccessKey           string
	SecretKey           string
	SessionToken        string
	Bucket              string
	Region              string
	RegionEndpoint      string
	RootDirectory       string
	Workers             uint
	CredentialsEndpoint string
}

type Bucket struct {
	*Config
	S3 *s3.Client
}

func initConfig() *Config {
	return &Config{
		RegionEndpoint: os.Getenv("REGION_ENDPOINT"),
		Bucket:         os.Getenv("BUCKET_NAME"),
		Region:         os.Getenv("BUCKET_ENDPOINT"),
		AccessKey:      os.Getenv("OBJ_ACCESS_KEY"),
		SecretKey:      os.Getenv("OBJ_SECRET_KEY"),
	}
}

func InitBucket() (*Bucket, error) {
	conf := initConfig()
	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(conf.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(conf.AccessKey, conf.SecretKey, conf.SessionToken)),
		config.WithBaseEndpoint(conf.RegionEndpoint),
	)

	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsConfig)

	return &Bucket{
		Config: conf,
		S3:     client,
	}, err
}
