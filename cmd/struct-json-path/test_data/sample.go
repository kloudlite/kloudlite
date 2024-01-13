package test_data

type EmbeddedStruct struct {
	Hi string `json:"hi"`
}

type InlineStruct struct {
	Value          string `json:"value"`
	EmbeddedStruct `json:"embedded2"`
}

type Sample struct {
	Hello    string         `json:"hello" asdfas:"Asdfasdfa"`
	Embedded EmbeddedStruct `json:"embedded"`
	Example  string
	Inline   InlineStruct `json:",inline"`
}
