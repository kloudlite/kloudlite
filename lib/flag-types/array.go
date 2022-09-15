package flag_types

type StringArray []string

func (i *StringArray) String() string {
	return "<nothing>"
}

func (i *StringArray) Set(value string) error {
	*i = append(*i, value)
	return nil
}
