package functions

import (
	"encoding/base64"
	libJson "encoding/json"
	nanoid "github.com/matoous/go-nanoid/v2"
	"kloudlite.io/pkg/errors"
	"regexp"
	"strings"
)

func NewBool(b bool) *bool {
	return &b
}

type JsonFeatures interface {
	ToB64Url(v interface{}) (string, error)
	ToB64String(v interface{}) (string, error)
}

type jsonFeatures struct{}

func (j *jsonFeatures) ToB64Url(v interface{}) (string, error) {
	b, e := libJson.Marshal(v)
	return base64.URLEncoding.EncodeToString(b), e
}

func (j *jsonFeatures) ToB64String(v interface{}) (string, error) {
	b, e := libJson.Marshal(v)
	return base64.StdEncoding.EncodeToString(b), e
}

var Json = &jsonFeatures{}

func ToBase64StringFromJson(v interface{}) (string, error) {
	b, e := libJson.Marshal(v)
	return base64.StdEncoding.EncodeToString(b), e
}

var re = regexp.MustCompile(`(\W|_)+`)

func Must[T any](value T, err error) T {
	if err != nil {
		panic(errors.NewEf(err, "panicking as Must() check failed"))
	}
	return value
}

func CleanerNanoid(n int) (string, error) {
	id, e := nanoid.New(n)
	if e != nil {
		return "", e
	}
	res := re.ReplaceAllString(id, "-")
	if strings.HasPrefix(res, "-") {
		res = "k" + res
	}
	if strings.HasSuffix(res, "-") {
		res = res + "k"
	}
	return res, nil
}
