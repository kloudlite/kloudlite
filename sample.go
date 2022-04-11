package main

type sample struct{}

func (sample *sample) Get() string {
	panic("not implemented") // TODO: Implement
}

func (sample *sample) Set() bool {
	panic("not implemented") // TODO: Implement
}

type Sample interface {
	Get() string
	Set() bool
}

func NewSample() Sample {
	return &sample{}
}
