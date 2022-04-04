package functions

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"regexp"
	"strings"

	nanoid "github.com/matoous/go-nanoid/v2"
)

func ToBytes(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ToBase64String(v interface{}) (string, error) {
	b, e := ToBytes(v)
	return base64.StdEncoding.EncodeToString(b), e
}

func ToBase64StringFromJson(v interface{}) (string, error) {
	b, e := json.Marshal(v)
	return base64.StdEncoding.EncodeToString(b), e
}

var re = regexp.MustCompile("(\\W|_)+")

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
