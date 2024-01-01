package types

type Example struct {
	Message string `json:"message"`
}

type SampleString string

const (
	Item1 SampleString = "item_1"
	Item2 SampleString = "item_2"
	Item3 SampleString = "item_3"
)
