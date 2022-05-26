package functions

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	libJson "encoding/json"
	t "operators.kloudlite.io/lib/types"
	"regexp"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	nanoid "github.com/matoous/go-nanoid/v2"

	"operators.kloudlite.io/lib/errors"
)

func NewBool(b bool) *bool {
	return &b
}

func StatusFromBool(b bool) metav1.ConditionStatus {
	if b {
		return metav1.ConditionTrue
	}
	return metav1.ConditionFalse
}

type JsonFeatures interface {
	ToB64Url(v interface{}) (string, error)
	ToB64String(v interface{}) (string, error)
	FromB64Url(s string, v interface{}) error
	FromTo(v interface{}, rt interface{}) error
	FromRawMessage(msg json.RawMessage, result interface{}) error
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

func (j *jsonFeatures) FromB64Url(s string, v interface{}) error {
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return errors.NewEf(err, "not a valid b64-url string")
	}
	err = json.Unmarshal(b, &v)
	if err != nil {
		return errors.NewEf(err, "could not unmarshal")
	}

	return nil
}

func (j *jsonFeatures) FromTo(from interface{}, to interface{}) error {
	marshal, err := json.Marshal(from)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(marshal, to); err != nil {
		return err
	}
	return nil
}

func (j *jsonFeatures) FromRawMessage(msg json.RawMessage) (map[string]interface{}, error) {
	m, err := msg.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(m, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (j *jsonFeatures) Hash(v interface{}) (string, error) {
	marshal, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	h := md5.New()
	h.Write(marshal)
	return hex.EncodeToString(h.Sum(nil)), nil
}

var Json = &jsonFeatures{}

func ToBase64StringFromJson(v interface{}) (string, error) {
	b, e := libJson.Marshal(v)
	return base64.StdEncoding.EncodeToString(b), e
}

var re = regexp.MustCompile(`(\W|_)+`)

func CleanerNanoid(n int) string {
	id := nanoid.Must(n)
	res := re.ReplaceAllString(id, "-")
	if strings.HasPrefix(res, "-") {
		res = "k" + res
	}
	if strings.HasSuffix(res, "-") {
		res = res + "k"
	}
	return res
}

func IfThenElse[T any](cond bool, v T, y T) T {
	if cond {
		return v
	}
	return y
}

func IfThenElseFn[T any](cond bool, v1 func() T, v2 func() T) T {
	if cond {
		return v1()
	}
	return v2()
}

func mapGet[T any](m map[string]any, key string) (T, bool) {
	if m == nil {
		return *new(T), false
	}
	v, ok := m[key]
	if !ok {
		return *new(T), false
	}
	tv, ok := v.(T)
	if !ok {
		return *new(T), false
	}
	return tv, ok
}

func MapGet[T any](m map[string]any, key string) (T, bool) {
	return mapGet[T](m, key)
}

func MapSet[T any](m map[string]T, key string, value T) {
	if m == nil {
		m = map[string]T{}
	}
	m[key] = value
}

func JsonGet[T any](s t.RawJson, key string) (T, error) {
	m, err := s.MarshalJSON()
	if err != nil {
		return *new(T), err
	}
	var j map[string]any
	if err := json.Unmarshal(m, &j); err != nil {
		return *new(T), err
	}

	value, ok := mapGet[T](j, key)
	if !ok {
		return *new(T), errors.NewEf(err, "key %s not found", key)
	}
	return value, nil
}
