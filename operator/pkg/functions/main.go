package functions

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os/exec"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	nanoid "github.com/matoous/go-nanoid/v2"

	"github.com/kloudlite/operator/pkg/errors"
)

type JsonFeatures interface {
	ToB64Url(v interface{}) (string, error)
	ToB64String(v interface{}) (string, error)
	FromB64Url(s string, v interface{}) error
	FromTo(v interface{}, rt interface{}) error
	FromRawMessage(msg json.RawMessage, result interface{}) error
}

type jsonFeatures struct{}

func (j *jsonFeatures) ToB64Url(v interface{}) (string, error) {
	b, e := json.Marshal(v)

	return base64.URLEncoding.EncodeToString(b), e
}

func (j *jsonFeatures) ToB64String(v interface{}) (string, error) {
	b, e := json.Marshal(v)
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
	b, e := json.Marshal(v)
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

func DefaultIfNil[T any](v *T) T {
	if v != nil {
		return *v
	}
	var dval T
	return dval
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

// MapContains checks if `destination` contains all keys from `source`
func MapContains[T comparable](destination map[string]T, source map[string]T) bool {
	if len(destination) == 0 && len(source) == 0 {
		return true
	}

	for k, v := range source {
		if destination[k] != v {
			return false
		}
	}
	return true
}

func MapEqual[K comparable, V comparable](first map[K]V, second map[K]V) bool {
	if len(first) != len(second) {
		return false
	}

	for k := range first {
		if second[k] != first[k] {
			return false
		}
	}
	return true
}

func NN(namespace, name string) types.NamespacedName {
	return types.NamespacedName{Namespace: namespace, Name: name}
}

func NewUnstructured(t metav1.TypeMeta, m ...metav1.ObjectMeta) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": t.APIVersion,
			"kind":       t.Kind,
		},
	}

	if len(m) > 0 {
		obj.Object["metadata"] = m[0]
	}

	return obj
}

func Md5(b []byte) string {
	hash := md5.New()
	hash.Write(b)
	return hex.EncodeToString(hash.Sum(nil))
}

func New[T any](v T) *T {
	return &v
}

func Sha1Sum(b []byte) string {
	hash := sha1.New()
	hash.Write(b)
	return hex.EncodeToString(hash.Sum(nil))
}

func Exec(command ...string) (err error, stdout *bytes.Buffer, stderr *bytes.Buffer) {
	if len(command) > 100 {
		return errors.New("command is too long"), nil, nil
	}
	args := make([]string, 0, len(command)+1)
	args = append(args, "-c")
	args = append(args, command...)

	stdout = bytes.NewBuffer(nil)
	stderr = bytes.NewBuffer(nil)

	cmd := exec.Command("bash", args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err = cmd.Run(); err != nil {
		return err, stdout, stderr
	}
	return nil, stdout, stderr
}

func Filter[T comparable](from []T, items []T, filterFunc func(fromItem T, targetItem T) bool) []T {
	tm := make(map[T]bool, len(items))
	result := make([]T, 0, len(items))

	for i := range items {
		tm[items[i]] = true
	}

	for i := range from {
		if tm[from[i]] {
			result = append(result, from[i])
		}
	}
	return result
}
