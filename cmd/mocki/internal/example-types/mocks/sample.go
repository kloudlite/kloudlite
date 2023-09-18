package mocks

import (
	io2 "io"
	example_types "kloudlite.io/cmd/mocki/internal/example-types"
)

type SampleCallerInfo struct {
	Args []any
}

type Sample[T any, K any] struct {
	Calls             map[string][]SampleCallerInfo
	MockAge           func() int
	MockName          func() string
	MockSetAndGetUser func(name string, age int, ex example_types.Example) *example_types.User
	MockSetName       func(name string)
	MockSetUser       func(name string, age int, ex example_types.Example, writer io2.Writer)
}

func (m *Sample[T, K]) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]SampleCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], SampleCallerInfo{Args: args})
}

func (sMock *Sample[T, K]) Age() int {
	if sMock.MockAge != nil {
		sMock.registerCall("Age")
		return sMock.MockAge()
	}
	panic("method 'Age' not implemented, yet")
}

func (sMock *Sample[T, K]) Name() string {
	if sMock.MockName != nil {
		sMock.registerCall("Name")
		return sMock.MockName()
	}
	panic("method 'Name' not implemented, yet")
}

func (sMock *Sample[T, K]) SetAndGetUser(name string, age int, ex example_types.Example) *example_types.User {
	if sMock.MockSetAndGetUser != nil {
		sMock.registerCall("SetAndGetUser", name, age, ex)
		return sMock.MockSetAndGetUser(name, age, ex)
	}
	panic("method 'SetAndGetUser' not implemented, yet")
}

func (sMock *Sample[T, K]) SetName(name string) {
	if sMock.MockSetName != nil {
		sMock.registerCall("SetName", name)
		sMock.MockSetName(name)
	}
	panic("method 'SetName' not implemented, yet")
}

func (sMock *Sample[T, K]) SetUser(name string, age int, ex example_types.Example, writer io2.Writer) {
	if sMock.MockSetUser != nil {
		sMock.registerCall("SetUser", name, age, ex, writer)
		sMock.MockSetUser(name, age, ex, writer)
	}
	panic("method 'SetUser' not implemented, yet")
}

func NewSample[T any, K any]() *Sample[T, K] {
	return &Sample[T, K]{}
}
