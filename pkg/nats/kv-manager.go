package nats

import (
	"context"
	"time"

	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/nats-io/nats.go/jetstream"
)

type KeyValueManager struct {
	jc *JetstreamClient
}

func (kvm KeyValueManager) ListStores(ctx context.Context) []string {
	kvml := kvm.jc.Jetstream.KeyValueStoreNames(ctx)
	buckets := make([]string, 0, 2)
	for e := range kvml.Name() {
		buckets = append(buckets, e)
	}
	return buckets
}

type CreateStoreArgs struct {
	Replicas     int
	TTL          *time.Duration
	MaxValueSize *int32
	Description  *string
}

func (kvm KeyValueManager) CreateStore(ctx context.Context, store string, args CreateStoreArgs) error {
	_, err := kvm.jc.Jetstream.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:       store,
		Description:  fn.DefaultIfNil(args.Description),
		MaxValueSize: fn.DefaultIfNil(args.MaxValueSize),
		TTL:          fn.DefaultIfNil(args.TTL),
		Storage:      jetstream.FileStorage,
		Replicas:     args.Replicas,
	})
	return err
}

func (kvm KeyValueManager) DeleteStore(ctx context.Context, store string) error {
	return kvm.jc.Jetstream.DeleteKeyValue(ctx, store)
}

func NewKeyValueManager(jc *JetstreamClient, bucketName string) (*KeyValueManager, error) {
	return &KeyValueManager{
		jc: jc,
	}, nil
}
