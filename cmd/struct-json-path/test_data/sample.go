package test_data

import corev1 "k8s.io/api/core/v1"

type EmbeddedStruct struct {
	Hi string `json:"hi"`
}

type InlineStruct struct {
	Value          string `json:"value"`
	EmbeddedStruct `json:"embedded2"`
}

//go:generate go run ../ --struct github.com/kloudlite/api/cmd/struct-json-path/test_data.Sample --out ./generated_jsonpath.go
type Sample struct {
	Hello            string `json:"hello" asdfas:"Asdfasdfa"`
	corev1.ConfigMap `json:",inline"`
	Embedded         EmbeddedStruct `json:"embedded"`
	Example          string
	Inline           InlineStruct `json:",inline"`
}
