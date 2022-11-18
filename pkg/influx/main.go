package influx

import (
	"context"

	"github.com/influxdata/influxdb-client-go/v2"
)

type Client struct {
	influxCli   influxdb2.Client
	isConnected bool
}

type Bucket struct {
	BucketId string `json:"bucketId"`
	OrgId    string `json:"orgId"`
}

func (c *Client) Connect(ctx context.Context) error {
	if _, err := c.influxCli.UsersAPI().Me(ctx); err != nil {
		return err
	}
	c.isConnected = true
	return nil
}

func (c *Client) Close() {
	c.influxCli.Close()
}

func (c *Client) bucketExists(ctx context.Context, bucketId string) error {
	if _, err := c.influxCli.BucketsAPI().FindBucketByID(ctx, bucketId); err != nil {
		return err
	}
	return nil
}

func (c *Client) BucketExists(ctx context.Context, bucket *Bucket) (bool, error) {
	if bucket == nil {
		return false, nil
	}
	if err := c.bucketExists(ctx, bucket.BucketId); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Client) UpsertBucket(ctx context.Context, orgName string, bucketName string) (*Bucket, error) {
	org, err := c.influxCli.OrganizationsAPI().FindOrganizationByName(ctx, orgName)
	if err != nil {
		return nil, err
	}

	bucket, err := c.influxCli.BucketsAPI().FindBucketByName(ctx, bucketName)
	if err != nil {
		return nil, err
	}

	if bucket != nil {
		return &Bucket{
			BucketId: *bucket.Id,
			OrgId:    *bucket.OrgID,
		}, nil
	}

	bucket, err = c.influxCli.BucketsAPI().CreateBucketWithNameWithID(ctx, *org.Id, bucketName)
	if err != nil {
		return nil, err
	}
	return &Bucket{
		BucketId: *bucket.Id,
		OrgId:    *bucket.OrgID,
	}, nil
}

func (c *Client) DeleteBucket(ctx context.Context, bucketId string) error {
	return c.influxCli.BucketsAPI().DeleteBucketWithID(ctx, bucketId)
}

func (c *Client) DeleteOrganisation(ctx context.Context, orgId string) error {
	return c.influxCli.OrganizationsAPI().DeleteOrganizationWithID(ctx, orgId)
}

func NewClient(httpUrl, token string) *Client {
	client := influxdb2.NewClient(httpUrl, token)
	return &Client{influxCli: client}
}
