package kv

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kloudlite/api/pkg/nats"
	"github.com/nats-io/nats.go/jetstream"

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

func (r *natsKVRepo[T]) Set(c context.Context, _key string, value T) error {
	key := sanitiseKey(_key)
	v := Value[T]{
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

func (r *natsKVRepo[T]) Get(c context.Context, _key string) (T, error) {
	key := sanitiseKey(_key)
	get, err := r.keyValue.Get(c, key)
	if err != nil {
		var x T
		return x, errors.NewE(err)
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
	return value.Data, errors.NewE(err)
}

func (r *natsKVRepo[T]) ErrKeyNotFound(err error) bool {
	return errors.Is(err, jetstream.ErrKeyNotFound)
}

func sanitiseKey(key string) string {
	return strings.ReplaceAll(key, ":", "-")
}

func (r *natsKVRepo[T]) SetWithExpiry(c context.Context, _key string, value T, duration time.Duration) error {
	key := sanitiseKey(_key)
	v := Value[T]{
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

func (r *natsKVRepo[T]) Drop(c context.Context, key string) error {
	return r.keyValue.Delete(c, sanitiseKey(key))
}

func (r *natsKVRepo[T]) ErrNoRecord(err error) bool {
	return errors.Is(err, jetstream.ErrKeyNotFound)
}

func NewNatsKVRepo[T any](ctx context.Context, bucketName string, jc *nats.JetstreamClient) (Repo[T], error) {
	if value, err := jc.Jetstream.KeyValue(ctx, bucketName); err != nil {
		return nil, errors.NewE(err)
	} else {
		return &natsKVRepo[T]{
			value,
		}, nil
	}
}
