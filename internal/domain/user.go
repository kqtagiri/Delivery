package domain

import (
	"net/mail"
)

type User struct {
	Name     string
	Email    string
	Password string
	Role     string //(user, admin)
	Address  string
	Balance  float64
	//Phone_Number(in future)
}

func NewUser(name, email, password, role, address string) (*User, error) {

	if name == "" {
		return nil, ErrInvalidName
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return nil, ErrInvalidEmail
	}

	if password == "" || len(password) < 6 {
		return nil, ErrInvalidPassword
	}

	if role != AdminRole && role != UserRole {
		return nil, ErrInvalidRole
	}

	if address == "" {
		return nil, ErrInvalidAddress
	}

	return &User{
		Name:     name,
		Email:    email,
		Password: password,
		Role:     role,
		Address:  address,
		Balance:  0,
	}, nil

}

func (u *User) ReplenishBalance(money float64) error {

	if money < 0 {
		return ErrInvalidReplenish
	}

	u.Balance += money
	return nil

}

func (u *User) TakeBalance(money float64) error {

	if money < 0 {
		return ErrInvalidPrice
	}

	if u.Balance-money < 0 {
		return ErrInvlaidPay
	}

	u.Balance -= money
	return nil

}
