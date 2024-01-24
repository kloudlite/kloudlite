package kv

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kloudlite/api/pkg/nats"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/kloudlite/api/pkg/errors"
)

type natsKVBinaryRepo struct {
	keyValue jetstream.KeyValue
}

type BinaryValue struct {
	Data      []byte
	ExpiresAt time.Time
}

func (v *BinaryValue) isExpired() bool {
	if v.ExpiresAt.IsZero() {
		return false
	}
	return time.Since(v.ExpiresAt) > 0
}

func (r *natsKVBinaryRepo) Set(c context.Context, key string, value []byte) error {
	v := BinaryValue{
		Data: value,
	}
	b, err := json.Marshal(v)
	if err != nil {
		return errors.NewEf(err, "failed to marshal value")
	}
	if _, err := r.keyValue.Put(c, key, b); err != nil {
		return errors.NewE(err)
	}
	return nil
}

func (r *natsKVBinaryRepo) Get(c context.Context, key string) ([]byte, error) {
	get, err := r.keyValue.Get(c, key)
	if err != nil {
		return nil, errors.NewE(err)
	}
	var value BinaryValue
	err = json.Unmarshal(get.Value(), &value)
	if value.isExpired() {
		go func() {
			if err = r.Drop(c, key); err != nil {
				fmt.Printf("unable to drop key %s", key)
			}
		}()
		return value.Data, errors.New("Key is expired")
	}
	return value.Data, errors.NewE(err)
}

// ErrKeyNotFound implements BinaryDataRepo.
func (*natsKVBinaryRepo) ErrKeyNotFound(err error) bool {
	return errors.Is(err, jetstream.ErrKeyNotFound)
}

func (r *natsKVBinaryRepo) SetWithExpiry(c context.Context, key string, value []byte, duration time.Duration) error {
	v := BinaryValue{
		Data:      value,
		ExpiresAt: time.Now().Add(duration),
	}
	b, err := json.Marshal(v)
	if err != nil {
		return errors.NewEf(err, "failed to marshal value")
	}
	if _, err := r.keyValue.Put(c, key, b); err != nil {
		return errors.NewE(err)
	}
	return nil
}

func (r *natsKVBinaryRepo) Drop(c context.Context, key string) error {
	return r.keyValue.Delete(c, key)
}

func NewNatsKVBinaryRepo(ctx context.Context, bucketName string, jc *nats.JetstreamClient) (BinaryDataRepo, error) {
	value, err := jc.Jetstream.KeyValue(ctx, bucketName)
	if err != nil {
		return nil, err
	}
	return &natsKVBinaryRepo{
		value,
	}, nil
}
