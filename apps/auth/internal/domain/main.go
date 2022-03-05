package domain

type userDomain struct {
	userRepo UserRepository
}

func (u *userDomain) Create(user *User) error {
	return u.userRepo.Create(user)
}

func MakeDomain(userRepo *UserRepository) *userDomain {
	return &userDomain{}
}
