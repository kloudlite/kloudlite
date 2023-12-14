package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kloudlite/api/pkg/nats"
	"github.com/nats-io/nats.go/jetstream"
	"time"

	"github.com/kloudlite/api/pkg/errors"
)

type natsKVRepo[T any] struct {
	keyValue jetstream.KeyValue
}

type Value[T any] struct {
	Data      T
	ExpiresAt time.Time
}

func (v *Value[T]) isExpired() bool {
	if v.ExpiresAt.IsZero() {
		return false
	}
	return time.Since(v.ExpiresAt) > 0
}

func (r *natsKVRepo[T]) Set(c context.Context, key string, value T) error {
	v := Value[T]{
		Data: value,
	}
	b, err := json.Marshal(v)
	if err != nil {
		return errors.NewEf(err, "failed to marshal value")
	}
	if _, err := r.keyValue.Put(c, key, b); err != nil {
		return err
	}
	return nil
}

func (r *natsKVRepo[T]) Get(c context.Context, key string) (T, error) {
	get, err := r.keyValue.Get(c, key)
	if err != nil {
		var x T
		return x, err
	}
	var value Value[T]
	err = json.Unmarshal(get.Value(), &value)
	if value.isExpired() {
		go func() {
			if err = r.Drop(c, key); err != nil {
				fmt.Printf("unable to drop key %s", key)
			}
		}()
		return value.Data, errors.New("Key is expired")
	}
	return value.Data, err
}

func (r *natsKVRepo[T]) SetWithExpiry(c context.Context, key string, value T, duration time.Duration) error {
	v := Value[T]{
		Data:      value,
		ExpiresAt: time.Now().Add(duration),
	}
	b, err := json.Marshal(v)
	if err != nil {
		return errors.NewEf(err, "failed to marshal value")
	}
	if _, err := r.keyValue.Put(c, key, b); err != nil {
		return err
	}
	return nil
}

func (r *natsKVRepo[T]) Drop(c context.Context, key string) error {
	return r.keyValue.Delete(c, key)
}

func (r *natsKVRepo[T]) ErrNoRecord(err error) bool {
	return err == nil
}

func NewNatsKVRepo[T any](ctx context.Context, bucketName string, jc *nats.JetstreamClient) (Repo[T], error) {
	if value, err := jc.Jetstream.KeyValue(ctx, bucketName); err != nil {
		return nil, err
	} else {
		return &natsKVRepo[T]{
			value,
		}, nil
	}
}
