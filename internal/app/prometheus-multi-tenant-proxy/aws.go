package proxy

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	signer "github.com/aws/aws-sdk-go/aws/signer/v4"
)

const (
	// AwsRegionEnvVar is the environment variable that specifies the AWS region
	AwsRegionEnvVar = "AWS_REGION"
	// AwsServiceEnvVar is the environment variable that specifies the AWS service
	AwsServiceEnvVar = "AWS_SERVICE_NAME"

	// AwsRegionDefault is the default AWS region used for signing
	AwsRegionDefault = "us-east-1"
	// AwsServiceDefault is the default AWS service used for signing
	AwsServiceDefault = "aps"

	// Other AWS environment variables are defined by AWS SDK's NewEnvCredentials
	// See https://docs.aws.amazon.com/sdk-for-php/v3/developer-guide/guide_credentials_environment.html
)

// AWSSigner is a wrapper around the AWS SDK's Signer
// Signing is required to use the proxy against an AWS prometheus service endpoint.
// HTTP requests will be signed using the AWS credentials from the environment
// and the AWS_DEFAULT_REGION (default=us-east-1) and AWS_SERVICE_NAME (default=aps).
// See https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-header-based-auth.html
// and https://docs.aws.amazon.com/sdk-for-go/api/service/signer for more details.
type AWSSigner struct {
	signer  *signer.Signer
	region  string
	service string
}

// NewAWSSigner creates a new AWS Signer using credentials from environment variables.
func NewAWSSigner() *AWSSigner {
	creds := credentials.NewEnvCredentials()
	if _, err := creds.Get(); err != nil {
		log.Fatalf("failed to get AWS credentials: %v", err)
	}

	return &AWSSigner{
		signer:  signer.NewSigner(creds),
		region:  getEnvOrDefault(AwsRegionEnvVar, getEnvOrDefault("AWS_DEFAULT_REGION", AwsRegionDefault)),
		service: getEnvOrDefault(AwsServiceEnvVar, AwsServiceDefault),
	}
}

func (s *AWSSigner) String() string {
	return fmt.Sprintf("AWSSigner{region=%s, service=%s}", s.region, s.service)
}

// SignAfter wraps an existing Director function (see https://pkg.go.dev/net/http/httputil#ReverseProxy).
// The request will be signed after the transformation by the original director.
func (s *AWSSigner) SignAfter(director func(*http.Request)) func(*http.Request) {
	return func(req *http.Request) {
		director(req)
		s.Sign(req)
	}
}

// Sign signs the HTTP request using the AWS Signer, meaning it adds
// the proper Authorization header using signature v4.
// See https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-header-based-auth.html
func (s *AWSSigner) Sign(req *http.Request) error {
	_, err := s.signer.Sign(req, nil, s.service, s.region, time.Now())
	if err != nil {
		log.Printf("Could not sign request: %s", err)
	}
	return err
}

func getEnvOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
