package domain

type UserRepository struct {
	Create func(user *User) error
}
