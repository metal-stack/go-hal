package s3client

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/metal-stack/go-hal/pkg/api"
	"time"
)

func New(cfg *api.S3Config) (*s3.S3, error) {
	s, err := newSession(cfg)
	if err != nil {
		return nil, err
	}
	return s3.New(s), nil
}

func newSession(cfg *api.S3Config) (client.ConfigProvider, error) {
	dummyRegion := "dummy" // we don't use AWS S3, we don't need a proper region
	hostnameImmutable := true
	return session.NewSession(&aws.Config{
		Region:           &dummyRegion,
		Endpoint:         &cfg.Url,
		Credentials:      credentials.NewStaticCredentials(cfg.Key, cfg.Secret, ""),
		S3ForcePathStyle: &hostnameImmutable,
		SleepDelay:       time.Sleep,
		Retryer: client.DefaultRetryer{
			NumMaxRetries: 3,
			MinRetryDelay: 10 * time.Second,
		},
	})
}
