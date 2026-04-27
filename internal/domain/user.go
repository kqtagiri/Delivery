package domain

import "errors"

var (
	ErrInvalidEmail     = errors.New("Invalid email")
	ErrInvalidName      = errors.New("Invalid name")
	ErrInvalidAddress   = errors.New("Invalid address")
	ErrInvalidReplenish = errors.New("Invalid replenish")
	ErrInvalidPrice     = errors.New("Invalid price")
	ErrInvlaidPay       = errors.New("Invalid pay")
)

type User struct {
	Name    string
	Email   string
	Address string
	Balance float64
	//Phone_Number(in future)
}

func NewUser(name, email, address string) (error, *User) {

	if email == "" {
		return ErrInvalidEmail, nil
	}

	if name == "" {
		return ErrInvalidName, nil
	}

	if address == "" {
		return ErrInvalidAddress, nil
	}

	return nil, &User{
		Name:    name,
		Email:   email,
		Address: address,
		Balance: 0,
	}

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
