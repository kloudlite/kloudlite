package functions

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	libJson "encoding/json"
	corev1 "k8s.io/api/core/v1"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

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

func MapContains[T comparable](target map[string]T, m map[string]T) bool {
	if target == nil || m == nil {
		return true
	}
	for k, v := range m {
		if target[k] != v {
			return false
		}
	}
	return true
}

func MapMerge[T any](m1 map[string]T, m2 map[string]T) map[string]T {
	x := map[string]T{}
	for k, v := range m1 {
		x[k] = v
	}
	for k, v := range m2 {
		x[k] = v
	}
	return x
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

func ParseSecret(s *corev1.Secret) *corev1.Secret {
	s.TypeMeta = metav1.TypeMeta{
		Kind:       "Secret",
		APIVersion: "v1",
	}
	return s
}

func ParseConfigMap(s *corev1.ConfigMap) *corev1.ConfigMap {
	s.TypeMeta = metav1.TypeMeta{
		Kind:       "ConfigMap",
		APIVersion: "v1",
	}
	return s
}

func Md5(b []byte) string {
	sum := md5.New().Sum(b)
	return hex.EncodeToString(sum)
}

func New[T any](v T) *T {
	return &v
}
