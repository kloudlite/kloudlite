package operator

import (
	"github.com/kloudlite/operator/pkg/errors"
)

type fstring string

const (
	ErrNotInInputs        fstring = "key=%s not found in .Spec.Inputs"
	ErrNotInGeneratedVars fstring = "key=%s not found in .Status.GeneratedVars"
	ErrNotInDisplayVars   fstring = "key=%s not found in .Status.DisplayVars"
)

func ErrNotInReqLocals(key string) error {
	return errors.Newf("key '%s' not found in req.Locals", key)
}

func (f fstring) Format(args ...string) error {
	return errors.Newf(string(f), args)
}
