package domain

import (
	"net/mail"
)

type User struct {
	Name    string
	Email   string
	Address string
	Balance float64
	//Phone_Number(in future)
}

func NewUser(name, email, address string) (*User, error) {

	if _, err := mail.ParseAddress(email); err != nil {
		return nil, ErrInvalidEmail
	}

	if name == "" {
		return nil, ErrInvalidName
	}

	if address == "" {
		return nil, ErrInvalidAddress
	}

	return &User{
		Name:    name,
		Email:   email,
		Address: address,
		Balance: 0,
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
