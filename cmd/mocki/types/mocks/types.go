package mocks

import (
  types2 "kloudlite.io/cmd/struct-to-graphql/pkg/types"
  types1 "kloudlite.io/cmd/mocki/types"
)

type Sample[T, K any] struct {
  MockAge func() (int)
  MockName func() (string)
  MockSetAndGetUser func(name string, age int, ex types1.Example) (*types1.User)
  MockSetAndGetUser2 func(name string, age int, ex types1.Example) (user types2.Sample)
  MockSetAndGetUser3 func(name string, age int, ex types1.Example, s1 T, s2 K) (user types2.Sample)
  MockSetName func() 
  MockSetUser func() 
}

func (s Sample[T, K]) Age() (int) {
  if s.MockAge != nil {
    return s.MockAge()
  }
  panic("not implemented, yet")
}

func (s Sample[T, K]) Name() (string) {
  if s.MockName != nil {
    return s.MockName()
  }
  panic("not implemented, yet")
}

func (s Sample[T, K]) SetAndGetUser(name string, age int, ex types1.Example) (*types1.User) {
  if s.MockSetAndGetUser != nil {
    return s.MockSetAndGetUser(name, age, ex)
  }
  panic("not implemented, yet")
}

func (s Sample[T, K]) SetAndGetUser2(name string, age int, ex types1.Example) (user types2.Sample) {
  if s.MockSetAndGetUser2 != nil {
    return s.MockSetAndGetUser2(name, age, ex)
  }
  panic("not implemented, yet")
}

func (s Sample[T, K]) SetAndGetUser3(name string, age int, ex types1.Example, s1 T, s2 K) (user types2.Sample) {
  if s.MockSetAndGetUser3 != nil {
    return s.MockSetAndGetUser3(name, age, ex, s1, s2)
  }
  panic("not implemented, yet")
}

func (s Sample[T, K]) SetName()  {
  if s.MockSetName != nil {
    s.MockSetName()
  }
  panic("not implemented, yet")
}

func (s Sample[T, K]) SetUser()  {
  if s.MockSetUser != nil {
    s.MockSetUser()
  }
  panic("not implemented, yet")
}

