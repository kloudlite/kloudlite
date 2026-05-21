package pagination

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type cacheLikeReader struct {
	items []corev1.Pod
	calls int
}

func (r *cacheLikeReader) Get(context.Context, client.ObjectKey, client.Object, ...client.GetOption) error {
	return errors.New("not implemented")
}

func (r *cacheLikeReader) List(_ context.Context, list client.ObjectList, opts ...client.ListOption) error {
	r.calls++
	listOpts := &client.ListOptions{}
	for _, opt := range opts {
		opt.ApplyToList(listOpts)
	}
	if listOpts.Limit > 0 || listOpts.Continue != "" {
		return apierrors.NewBadRequest("continue list option is not supported by the cache")
	}
	pods, ok := list.(*corev1.PodList)
	if !ok {
		return apierrors.NewBadRequest("unexpected list type")
	}
	pods.Items = append([]corev1.Pod{}, r.items...)
	return nil
}

func TestListAllFallsBackWhenReaderDoesNotSupportPagination(t *testing.T) {
	reader := &cacheLikeReader{items: []corev1.Pod{{}, {}}}
	list := &corev1.PodList{}

	err := ListAll(context.Background(), reader, list)

	require.NoError(t, err)
	assert.Len(t, list.Items, 2)
	assert.Equal(t, 2, reader.calls)
}

func TestListAllReturnsOtherListErrors(t *testing.T) {
	reader := &cacheLikeReader{}
	list := &corev1.ServiceList{}

	err := ListAll(context.Background(), reader, list)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected list type")
}
