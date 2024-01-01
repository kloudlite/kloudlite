package fzf

import (
	"errors"

	"github.com/koki-develop/go-fzf"
	ofzf "github.com/koki-develop/go-fzf"
)

type Option ofzf.Option

func WithPrompt(prompt string) Option {
	return Option(ofzf.WithPrompt(prompt))
}

func FindOne[T any](items []T, itemFunc func(item T) string, options ...Option) (*T, error) {
	f, err := ofzf.New(func() []ofzf.Option {
		opts := make([]ofzf.Option, 0)
		for _, o := range options {
			opts = append(opts, ofzf.Option(o))
		}

		opts = append(opts, fzf.WithInputPlaceholder("search"))
		return opts
	}()...)

	if err != nil {
		return nil, err
	}

	idxs, err := f.Find(items, func(i int) string {
		return itemFunc(items[i])
	})

	if len(idxs) == 0 {
		return nil, errors.New("you have not selected any item")
	}

	selectedIndex := idxs[0]

	return &items[selectedIndex], nil
}
