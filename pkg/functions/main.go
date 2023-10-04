package functions

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	nanoid "github.com/matoous/go-nanoid/v2"
	"kloudlite.io/pkg/errors"
)

func ToBase64StringFromJson(v interface{}) (string, error) {
	b, e := json.Marshal(v)
	return base64.StdEncoding.EncodeToString(b), e
}

func Must[T any](value T, err error) T {
	if err != nil {
		panic(errors.NewEf(err, "panicking as Must() check failed"))
	}
	return value
}

var re = regexp.MustCompile(`(\W|_)+`)

func CleanerNanoid(n int) (string, error) {
	id, e := nanoid.New(n)
	if e != nil {
		return "", errors.NewEf(e, "could not get nanoid()")
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

func CleanerNanoidOrDie(n int) string {
	id, err := CleanerNanoid(n)
	if err != nil {
		panic(err)
	}
	return id
}

func JsonConversion(from any, to any) error {
	if from == nil {
		return nil
	}

	if to == nil {
		return fmt.Errorf("receiver (to) is nil")
	}

	b, err := json.Marshal(from)
	if err != nil {
		return nil
	}
	if err := json.Unmarshal(b, &to); err != nil {
		return err
	}
	return nil
}
