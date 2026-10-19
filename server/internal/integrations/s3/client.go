package s3

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type Config struct {
	Endpoint     string
	Region       string
	Bucket       string
	AccessKey    string
	SecretKey    string
	UsePathStyle bool
	PlaybackTTL  time.Duration
}

type Client struct {
	client      *s3.Client
	presigner   *s3.PresignClient
	bucket      string
	playbackTTL time.Duration
}

func New(_ context.Context, cfg Config) (*Client, error) {
	_, secure, err := parseEndpoint(cfg.Endpoint)
	if err != nil {
		return nil, err
	}
	if !secure {
		return nil, fmt.Errorf("customer S3 endpoint must use HTTPS")
	}

	cfg.Endpoint = strings.TrimSpace(cfg.Endpoint)
	cfg.Region = strings.TrimSpace(cfg.Region)
	cfg.Bucket = strings.TrimSpace(cfg.Bucket)
	cfg.AccessKey = strings.TrimSpace(cfg.AccessKey)
	cfg.SecretKey = strings.TrimSpace(cfg.SecretKey)

	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("S3 bucket is required")
	}
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("S3 credentials are required")
	}
	if cfg.PlaybackTTL <= 0 {
		cfg.PlaybackTTL = 15 * time.Minute
	}

	// Only customer-supplied credentials are used. No environment, instance
	// metadata or shared AWS profile is consulted for a tenant integration.
	awsConfig := aws.Config{
		Region: cfg.Region,
		Credentials: aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		),
		HTTPClient: &http.Client{
			Transport: publicTransport(),
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return fmt.Errorf("S3 HTTP redirects are not permitted")
			},
		},
		// Some S3-compatible providers do not implement optional CRC32
		// trailers. Required checksums remain enabled; optional ones are not
		// requested unless the application explicitly sets an algorithm.
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
	}
	client := s3.NewFromConfig(awsConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(cfg.Endpoint)
		options.UsePathStyle = cfg.UsePathStyle
	})

	return &Client{
		client:      client,
		presigner:   s3.NewPresignClient(client),
		bucket:      cfg.Bucket,
		playbackTTL: cfg.PlaybackTTL,
	}, nil
}

func (c *Client) Test(ctx context.Context) error {
	if _, err := c.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(c.bucket),
	}); err != nil {
		return fmt.Errorf("check S3 bucket: %w", err)
	}

	key := "leamout-storage-test/" + uuid.NewString()
	body := "leamout"
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(c.bucket),
		Key:           aws.String(key),
		Body:          strings.NewReader(body),
		ContentLength: aws.Int64(int64(len(body))),
		ContentType:   aws.String("text/plain"),
	})
	if err != nil {
		return fmt.Errorf("test S3 write access: %w", err)
	}

	cleanup := true
	defer func() {
		if cleanup {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, _ = c.client.DeleteObject(cleanupCtx, &s3.DeleteObjectInput{
				Bucket: aws.String(c.bucket),
				Key:    aws.String(key),
			})
		}
	}()

	if _, err := c.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}); err != nil {
		return fmt.Errorf("test S3 read access: %w", err)
	}
	if _, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}); err != nil {
		return fmt.Errorf("test S3 delete access: %w", err)
	}
	cleanup = false
	return nil
}

func (c *Client) Put(
	ctx context.Context,
	key string,
	contentType string,
	reader io.Reader,
	size int64,
) error {
	if err := validateKey(key); err != nil {
		return err
	}
	if size < 0 {
		return fmt.Errorf("S3 object size must be non-negative")
	}
	// The recording ingestion job supplies an *os.File. S3 SigV4 signing
	// requires a seekable reader for direct PutObject; do not buffer large
	// customer recordings into process memory.
	seeker, ok := reader.(io.ReadSeeker)
	if !ok {
		return fmt.Errorf("S3 upload source must be seekable")
	}
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(c.bucket),
		Key:           aws.String(key),
		Body:          seeker,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("upload S3 object: %w", err)
	}
	return nil
}

func (c *Client) PlaybackURL(
	ctx context.Context,
	key string,
) (string, time.Time, error) {
	if err := validateKey(key); err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().UTC().Add(c.playbackTTL)
	result, err := c.presigner.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(c.bucket),
			Key:    aws.String(key),
		},
		func(options *s3.PresignOptions) {
			options.Expires = c.playbackTTL
		},
	)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("presign S3 object: %w", err)
	}
	return result.URL, expiresAt, nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	if err := validateKey(key); err != nil {
		return err
	}
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete S3 object: %w", err)
	}
	return nil
}

func (c *Client) Bucket() string {
	return c.bucket
}

func parseEndpoint(value string) (string, bool, error) {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Host == "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") ||
		parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" ||
		(parsed.Path != "" && parsed.Path != "/") {
		return "", false, fmt.Errorf("S3 endpoint must be an HTTP or HTTPS origin")
	}
	return parsed.Host, parsed.Scheme == "https", nil
}

func publicTransport() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	transport.DialContext = func(
		ctx context.Context,
		network string,
		address string,
	) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, fmt.Errorf("invalid S3 endpoint address: %w", err)
		}
		addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("resolve S3 endpoint: %w", err)
		}
		var dialErr error
		for _, resolved := range addresses {
			if !publicEndpointIP(resolved.IP) {
				continue
			}
			conn, err := dialer.DialContext(
				ctx,
				network,
				net.JoinHostPort(resolved.IP.String(), port),
			)
			if err == nil {
				return conn, nil
			}
			dialErr = err
		}
		if dialErr != nil {
			return nil, dialErr
		}
		return nil, fmt.Errorf("S3 endpoint resolves to a non-public network")
	}
	return transport
}

func publicEndpointIP(ip net.IP) bool {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	addr = addr.Unmap()
	if !addr.IsGlobalUnicast() || addr.IsPrivate() ||
		addr.IsLoopback() || addr.IsLinkLocalUnicast() ||
		addr.IsMulticast() || addr.IsUnspecified() {
		return false
	}
	for _, blocked := range blockedPublicRanges {
		if blocked.Contains(addr) {
			return false
		}
	}
	return true
}

var blockedPublicRanges = []netip.Prefix{
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("198.18.0.0/15"),
}

func validateKey(key string) error {
	key = strings.TrimSpace(key)
	if key == "" || strings.HasPrefix(key, "/") || strings.Contains(key, "..") {
		return fmt.Errorf("S3 object key is invalid")
	}
	return nil
}
