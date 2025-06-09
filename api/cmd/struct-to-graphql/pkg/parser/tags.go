package parser

import (
	"fmt"
	"reflect"
	"strings"
)

type GraphqlTag struct {
	Uri               *string
	Enum              []string
	Ignore            bool
	NoInput           bool
	OnlyInput         bool
	InputOmitEmpty    bool
	DefaultValue      any
	ChildrenRequired  bool
	ChildrenOmitEmpty bool
	ScalarType        *string
}

func parseGraphqlTag(field reflect.StructField) (GraphqlTag, error) {
	tag := field.Tag.Get("graphql")
	if tag == "" {
		return GraphqlTag{}, nil
	}

	var gt GraphqlTag
	sp := strings.Split(tag, ",")
	for i := range sp {
		kv := strings.Split(sp[i], "=")

		switch kv[0] {
		case "uri":
			{
				if len(kv) != 2 {
					return GraphqlTag{}, fmt.Errorf("invalid graphql tag %s, must be of form key=value", tag)
				}
				gt.Uri = &kv[1]
			}
		case "enum":
			{
				if len(kv) != 2 {
					return GraphqlTag{}, fmt.Errorf("invalid graphql tag %s, must be of form key=value", tag)
				}
				enumVals := strings.Split(kv[1], ";")

				gt.Enum = make([]string, 0, len(enumVals))
				for k := range enumVals {
					if enumVals[k] != "" {
						gt.Enum = append(gt.Enum, enumVals[k])
					}
				}
			}
		case "noinput":
			{
				gt.NoInput = true
			}
		case "onlyinput":
			{
				gt.OnlyInput = true
			}

		case "inputomitempty":
			{
				gt.InputOmitEmpty = true
			}
		case "children-required":
			{
				gt.ChildrenRequired = true
			}
		case "children-omitempty":
			{
				gt.ChildrenOmitEmpty = true
			}

		case "ignore":
			{
				gt.Ignore = true
			}
		case "default":
			{
				if strings.HasPrefix(kv[1], "'") {
					return gt, fmt.Errorf("graphql string value can not start with single-quote, use double-quotes")
				}
				gt.DefaultValue = kv[1]
			}
		case "scalar-type":
			{
				gt.ScalarType = &kv[1]
			}

		default:
			{
				return GraphqlTag{}, fmt.Errorf("unknown graphql tag %s", kv[0])
			}
		}
	}

	return gt, nil
}
