package s3client

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
)

// ClientOpts represents options for configuring an S3 client.
type ClientOpts struct {
	// Region is the AWS region to connect to.
	Region string
	// Endpoint is the endpoint to connect to.
	Endpoint string
	// AccessKey is the access key to use for authentication.
	AccessKey string
	// SecretKey is the secret key to use for authentication.
	SecretKey string
	// AssumeRoleARN is the part of assume role parameter for assume role authentication.
	AssumeRoleARN string
	// AssumeRoleSessionName is the part of assume role parameter for assume role authentication.
	AssumeRoleSessionName string
	// AssumeRoleExternalID is the part of assume role parameter for assume role authentication.
	AssumeRoleExternalID string
}

// LoadOptions returns a slice of functions that can be passed to the config.Load function
// from the AWS SDK to configure an AWS client with the specified options.
func (o *ClientOpts) LoadOptions() []func(options *config.LoadOptions) error {
	var loadOpts []func(options *config.LoadOptions) error

	if o.Region != "" {
		// This is a special case for minio.
		//	https://github.com/minio/minio/discussions/12030#discussioncomment-590564
		//	this is backwards compatible flag to make it work with minio.
		if o.Region == "minio" {
			loadOpts = append(loadOpts, config.WithEndpointResolverWithOptions(
				aws.EndpointResolverWithOptionsFunc(
					func(service string, region string, options ...interface{}) (aws.Endpoint, error) {
						return aws.Endpoint{
							URL:               o.Endpoint,
							SigningRegion:     region,
							HostnameImmutable: true,
						}, nil
					},
				),
			))
		}
		// Add a function to the slice that sets the region on the LoadOptions.
		loadOpts = append(loadOpts, config.WithRegion(o.Region))
	}

	if o.AccessKey != "" && o.SecretKey != "" {
		// Add a function to the slice that sets the credentials' provider on the LoadOptions.
		loadOpts = append(loadOpts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				o.AccessKey,
				o.SecretKey,
				"",
			),
		))
	}

	if o.AssumeRoleARN == "" {
		return loadOpts
	}

	loadOpts = append(loadOpts,
		config.WithAssumeRoleCredentialOptions(func(options *stscreds.AssumeRoleOptions) {
			options.RoleARN = o.AssumeRoleARN
			if (o.AssumeRoleSessionName != "") {
				options.RoleSessionName = o.AssumeRoleSessionName
			}
			if (o.AssumeRoleExternalID != "") {
				options.ExternalID = aws.String(o.AssumeRoleExternalID)
			}
		}),
	)

	return loadOpts
}

// ClientOptsFunc is a function that takes a *ClientOpts pointer and returns an error.
type ClientOptsFunc func(*ClientOpts) error

// WithRegion returns a ClientOptsFunc that sets the region field on the ClientOpts.
func WithRegion(region string) ClientOptsFunc {
	return func(opts *ClientOpts) error {
		opts.Region = region
		return nil
	}
}

// WithEndpoint returns a ClientOptsFunc that sets the endpoint field on the ClientOpts.
func WithEndpoint(endpoint string) ClientOptsFunc {
	return func(opts *ClientOpts) error {
		opts.Endpoint = endpoint
		return nil
	}
}

// WithStaticCredentials returns a ClientOptsFunc that sets the access key and secret key fields on the ClientOpts.
func WithStaticCredentials(a, s string) ClientOptsFunc {
	return func(opts *ClientOpts) error {
		opts.AccessKey = a
		opts.SecretKey = s
		return nil
	}
}

// WithAssumeRoleCredentialOptions returns a ClientOptsFunc that sets of parameters for AssumeRole fields on the ClientOpts.
func WithAssumeRoleCredentialOptions(a, s, id string) ClientOptsFunc {
	return func(opts *ClientOpts) error {
		opts.AssumeRoleARN = a
		if (s != "") {
			opts.AssumeRoleSessionName = s
		}
		if (id != "") {
			opts.AssumeRoleExternalID = id
		}
		return nil
	}
}
