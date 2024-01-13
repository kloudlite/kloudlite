package test_data

var PkgPath = "github.com/kloudlite/api/cmd/struct-json-path/test_data"

type Test1 struct {
	Sample string
}

var Test1Input = "Test1"
var Test1Output = map[string][]string{
	"Sample": {},
}

type Test2 struct {
	Sample string `json:"sample"`
}

var Test2Input = "Test2"
var Test2Output = map[string][]string{
	"sample": {},
}

type Test3 struct {
	Sample  string `json:"sample"`
	Example string
}

var Test3Input = "Test3"
var Test3Output = map[string][]string{
	"sample":  {},
	"Example": {},
}

type Test4 struct {
	Sample  string `json:"sample"`
	Example struct {
		ExampleA string `json:"example-a"`
		ExampleB string `json:"example-b"`
		ExampleC string
	}
}

var Test4Input = "Test4"
var Test4Output = map[string][]string{
	"sample": {},
	"Example": {
		"example-a",
		"example-b",
		"ExampleC",
	},
}

type Test5 struct {
	Sample  string `json:"sample"`
	Example struct {
		ExampleA string `json:"example-a"`
		ExampleB string `json:"example-b"`
		ExampleC string
	}
	Example2 struct {
		ExampleA string `json:"example-a"`
		ExampleB string `json:"example-b"`
		ExampleC string
	} `json:"example2"`
}

var Test5Input = "Test5"
var Test5Output = map[string][]string{
	"sample": {},
	"Example": {
		"example-a",
		"example-b",
		"ExampleC",
	},
	"example2": {
		"example-a",
		"example-b",
		"ExampleC",
	},
}

type Test6 struct {
	Sample  string `json:"sample"`
	Example struct {
		ExampleA string `json:"example-a"`
		ExampleB string `json:"example-b"`
		ExampleC string
	}
	Example2 struct {
		ExampleA string `json:"example-a"`
		ExampleB string `json:"example-b"`
		ExampleC string
	} `json:",inline"`
}

var Test6Input = "Test6"
var Test6Output = map[string][]string{
	"sample": {},
	"Example": {
		"example-a",
		"example-b",
		"ExampleC",
	},
	"": {
		"example-a",
		"example-b",
		"ExampleC",
	},
}

type Test7 struct {
	Sample  string `json:"sample"`
	Example struct {
		ExampleA string `json:"example-a"`
		ExampleB string `json:"example-b"`
		ExampleC string
	}
	Example2 struct {
		ExampleA string `json:"example-a"`
		ExampleB string `json:"example-b"`
		ExampleC string
	} `json:",inline"`
	Example3 struct {
		ExampleA string `json:"example-c"`
		ExampleB string `json:"example-d"`
		ExampleC string
	} `json:",inline"`
}

var Test7Input = "Test7"
var Test7Output = map[string][]string{
	"sample": {},
	"Example": {
		"example-a",
		"example-b",
		"ExampleC",
	},
	"": {
		"example-a",
		"example-b",
		"example-c",
		"example-d",
		"ExampleC",
	},
}

type Test8Example struct {
	Hello string `json:"hello"`
	World string `json:"world"`
}

type Test8 struct {
	Example Test8Example `json:"example"`
}

var Test8Input = "Test8"
var Test8Output = map[string][]string{
	"example": {
		"hello",
		"world",
	},
}

type Test9Example struct {
	Hello string `json:"hello"`
	World string `json:"-"`
}

type Test9 struct {
	Example Test9Example `json:"example"`
}

var Test9Input = "Test9"
var Test9Output = map[string][]string{
	"example": {
		"hello",
	},
}

type Test10Example struct {
	Hello string `json:"hello"`
}
type Test10Example2 struct {
	World string `json:"world"`
}

type Test10 struct {
	Test10Example `json:"example"`
	Test10Example2
}

var Test10Input = "Test10"
var Test10Output = map[string][]string{
	"example": {
		"hello",
	},
	"Test10Example2": {
		"world",
	},
}

type Test11Example struct {
	Hello string `json:"hello"`
}
type Test11Example2 struct {
	World string `json:"world"`
}

type Test11 struct {
	Test11Example `json:"example"`
	Test11Example2
}

var Test11Input = "Test11"
var Test11Output = map[string][]string{
	"example": {
		"hello",
	},
	"Test11Example2": {
		"world",
	},
}
