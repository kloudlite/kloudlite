package mongo

type ErrUserExists struct {
  Message string `json:"message"`
}

func (e ErrUserExists) Error() string {
  return e.Message
}
