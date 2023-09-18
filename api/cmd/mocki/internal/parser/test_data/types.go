package test_data

import (
	"context"
	io2 "io"
	"kloudlite.io/pkg/repos"
)

type Type1 interface {
	Method1() string
}

//type DbRepo[T Entity] interface {
//	NewId() int
//}

type X struct {
	*repos.BaseEntity
}

type Type2[T any] interface {
	Method1() T
	Method2(x int) string
	Method3(x int, y *int, z T, p *repos.DbRepo[X], q map[string]X, r *X, s []int, u ...X) string
	//Method2(x int, y *int, z T, p *repos.DbRepo[X], q X, r *X, s []int, t []*X, u ...X) (int, string, *int, T, *repos.DbRepo[X], X, *X)
}

type Entity interface{}

type Type3[T Entity] interface {
	Method1() T
}

type Type4 interface {
	Method1(context.Context, int) string
}

type Type5 interface {
	Method1()
	Method2(x int)
}

type Type6 interface {
	Method1(writer io2.Writer)
}
