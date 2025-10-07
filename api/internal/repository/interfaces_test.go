package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithLabelSelector(t *testing.T) {
	opts := &ListOptions{}
	fn := WithLabelSelector("app=test")
	fn(opts)

	assert.Equal(t, "app=test", opts.LabelSelector)
}

func TestWithFieldSelector(t *testing.T) {
	opts := &ListOptions{}
	fn := WithFieldSelector("metadata.name=test")
	fn(opts)

	assert.Equal(t, "metadata.name=test", opts.FieldSelector)
}

func TestWithLimit(t *testing.T) {
	opts := &ListOptions{}
	fn := WithLimit(100)
	fn(opts)

	assert.Equal(t, int64(100), opts.Limit)
}

func TestWithContinue(t *testing.T) {
	opts := &ListOptions{}
	fn := WithContinue("continue-token-123")
	fn(opts)

	assert.Equal(t, "continue-token-123", opts.Continue)
}

func TestWithWatchLabelSelector(t *testing.T) {
	opts := &WatchOptions{}
	fn := WithWatchLabelSelector("app=test")
	fn(opts)

	assert.Equal(t, "app=test", opts.LabelSelector)
}

func TestWithWatchFieldSelector(t *testing.T) {
	opts := &WatchOptions{}
	fn := WithWatchFieldSelector("metadata.name=test")
	fn(opts)

	assert.Equal(t, "metadata.name=test", opts.FieldSelector)
}

func TestWithResourceVersion(t *testing.T) {
	opts := &WatchOptions{}
	fn := WithResourceVersion("12345")
	fn(opts)

	assert.Equal(t, "12345", opts.ResourceVersion)
}

func TestWithWatchTimeout(t *testing.T) {
	opts := &WatchOptions{}
	fn := WithWatchTimeout(60)
	fn(opts)

	assert.NotNil(t, opts.TimeoutSeconds)
	assert.Equal(t, int64(60), *opts.TimeoutSeconds)
}

func TestBuildListOptions(t *testing.T) {
	t.Run("with all options", func(t *testing.T) {
		opts := &ListOptions{
			LabelSelector: "app=test",
			FieldSelector: "metadata.name=test",
			Limit:         100,
			Continue:      "token-123",
		}

		listOpts := buildListOptions(opts)

		assert.Equal(t, "app=test", listOpts.LabelSelector)
		assert.Equal(t, "metadata.name=test", listOpts.FieldSelector)
		assert.Equal(t, int64(100), listOpts.Limit)
		assert.Equal(t, "token-123", listOpts.Continue)
	})

	t.Run("with empty options", func(t *testing.T) {
		opts := &ListOptions{}
		listOpts := buildListOptions(opts)

		assert.Equal(t, "", listOpts.LabelSelector)
		assert.Equal(t, "", listOpts.FieldSelector)
		assert.Equal(t, int64(0), listOpts.Limit)
		assert.Equal(t, "", listOpts.Continue)
	})

	t.Run("with partial options", func(t *testing.T) {
		opts := &ListOptions{
			LabelSelector: "app=test",
			Limit:         50,
		}

		listOpts := buildListOptions(opts)

		assert.Equal(t, "app=test", listOpts.LabelSelector)
		assert.Equal(t, "", listOpts.FieldSelector)
		assert.Equal(t, int64(50), listOpts.Limit)
		assert.Equal(t, "", listOpts.Continue)
	})
}

func TestBuildWatchOptions(t *testing.T) {
	t.Run("with all options", func(t *testing.T) {
		timeout := int64(60)
		opts := &WatchOptions{
			LabelSelector:       "app=test",
			FieldSelector:       "metadata.name=test",
			ResourceVersion:     "12345",
			TimeoutSeconds:      &timeout,
			AllowWatchBookmarks: true,
		}

		watchOpts := buildWatchOptions(opts)

		assert.True(t, watchOpts.Watch)
		assert.Equal(t, "app=test", watchOpts.LabelSelector)
		assert.Equal(t, "metadata.name=test", watchOpts.FieldSelector)
		assert.Equal(t, "12345", watchOpts.ResourceVersion)
		assert.NotNil(t, watchOpts.TimeoutSeconds)
		assert.Equal(t, int64(60), *watchOpts.TimeoutSeconds)
		assert.True(t, watchOpts.AllowWatchBookmarks)
	})

	t.Run("with empty options", func(t *testing.T) {
		opts := &WatchOptions{}
		watchOpts := buildWatchOptions(opts)

		assert.True(t, watchOpts.Watch)
		assert.Equal(t, "", watchOpts.LabelSelector)
		assert.Equal(t, "", watchOpts.FieldSelector)
		assert.Equal(t, "", watchOpts.ResourceVersion)
		assert.Nil(t, watchOpts.TimeoutSeconds)
		assert.False(t, watchOpts.AllowWatchBookmarks)
	})

	t.Run("with partial options", func(t *testing.T) {
		opts := &WatchOptions{
			LabelSelector:   "app=test",
			ResourceVersion: "12345",
		}

		watchOpts := buildWatchOptions(opts)

		assert.True(t, watchOpts.Watch)
		assert.Equal(t, "app=test", watchOpts.LabelSelector)
		assert.Equal(t, "", watchOpts.FieldSelector)
		assert.Equal(t, "12345", watchOpts.ResourceVersion)
		assert.Nil(t, watchOpts.TimeoutSeconds)
		assert.False(t, watchOpts.AllowWatchBookmarks)
	})
}

func TestListOptions_MultipleOptionFunctions(t *testing.T) {
	opts := &ListOptions{}

	WithLabelSelector("app=test")(opts)
	WithFieldSelector("metadata.name=test")(opts)
	WithLimit(100)(opts)
	WithContinue("token-123")(opts)

	assert.Equal(t, "app=test", opts.LabelSelector)
	assert.Equal(t, "metadata.name=test", opts.FieldSelector)
	assert.Equal(t, int64(100), opts.Limit)
	assert.Equal(t, "token-123", opts.Continue)
}

func TestWatchOptions_MultipleOptionFunctions(t *testing.T) {
	opts := &WatchOptions{}

	WithWatchLabelSelector("app=test")(opts)
	WithWatchFieldSelector("metadata.name=test")(opts)
	WithResourceVersion("12345")(opts)
	WithWatchTimeout(60)(opts)

	assert.Equal(t, "app=test", opts.LabelSelector)
	assert.Equal(t, "metadata.name=test", opts.FieldSelector)
	assert.Equal(t, "12345", opts.ResourceVersion)
	assert.NotNil(t, opts.TimeoutSeconds)
	assert.Equal(t, int64(60), *opts.TimeoutSeconds)
}
