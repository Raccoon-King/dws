package s3

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
)

// Client wraps the S3 client with convenience methods
type Client struct {
	s3Client   *s3.S3
	downloader *s3manager.Downloader
}

// Config holds S3 client configuration
type Config struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	RoleARN         string
	Timeout         time.Duration
}

// NewClient creates a new S3 client with the provided configuration
func NewClient(config Config) (*Client, error) {
	// Set default timeout if not specified
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	awsConfig := &aws.Config{
		Region:     aws.String(config.Region),
		HTTPClient: &http.Client{Timeout: config.Timeout},
		MaxRetries: aws.Int(3), // Add retry logic
	}

	// Configure credentials based on what's provided
	var sess *session.Session
	var err error

	if config.RoleARN != "" {
		// Use IAM role
		sess, err = session.NewSession(awsConfig)
		if err != nil {
			return nil, err
		}
		creds := stscreds.NewCredentials(sess, config.RoleARN)
		awsConfig.Credentials = creds
	} else if config.AccessKeyID != "" && config.SecretAccessKey != "" {
		// Use access keys
		creds := credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, config.SessionToken)
		awsConfig.Credentials = creds
	}
	// If neither is provided, it will use the default credential chain (environment variables, IAM role, etc.)

	sess, err = session.NewSession(awsConfig)
	if err != nil {
		return nil, err
	}

	s3Client := s3.New(sess)
	downloader := s3manager.NewDownloader(sess)

	return &Client{
		s3Client:   s3Client,
		downloader: downloader,
	}, nil
}

// ParseS3URL parses an S3 URL and returns bucket and key
func ParseS3URL(s3URL string) (bucket, key string, err error) {
	u, err := url.Parse(s3URL)
	if err != nil {
		return "", "", err
	}

	if u.Scheme != "s3" {
		return "", "", fmt.Errorf("invalid S3 URL scheme: %s", u.Scheme)
	}

	bucket = u.Host
	key = strings.TrimPrefix(u.Path, "/")

	return bucket, key, nil
}

// DownloadFile downloads a file from S3 and returns its contents
func (c *Client) DownloadFile(ctx context.Context, bucket, key string) ([]byte, error) {
	logrus.WithFields(logrus.Fields{
		"bucket": bucket,
		"key":    key,
	}).Info("Downloading file from S3")

	buf := aws.NewWriteAtBuffer([]byte{})

	// Create a context with timeout
	downloadCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	_, err := c.downloader.DownloadWithContext(downloadCtx, buf, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"bucket": bucket,
			"key":    key,
			"error":  err,
		}).Error("Failed to download file from S3")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"bucket": bucket,
		"key":    key,
		"size":   len(buf.Bytes()),
	}).Info("Successfully downloaded file from S3")

	return buf.Bytes(), nil
}

// DownloadFileFromURL downloads a file from S3 using a full S3 URL
func (c *Client) DownloadFileFromURL(ctx context.Context, s3URL string) ([]byte, string, error) {
	bucket, key, err := ParseS3URL(s3URL)
	if err != nil {
		return nil, "", err
	}

	data, err := c.DownloadFile(ctx, bucket, key)
	if err != nil {
		return nil, "", err
	}

	return data, key, nil
}

// CheckFileExists checks if a file exists in S3
func (c *Client) CheckFileExists(ctx context.Context, bucket, key string) (bool, error) {
	_, err := c.s3Client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey, "NotFound":
				return false, nil
			}
		}
		return false, err
	}

	return true, nil
}