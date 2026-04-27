package domain

import "errors"

var (
	ErrInvalidNumber = errors.New("Invalid number")
	ErrInvalidTime   = errors.New("Invalid time")
	ErrInvalidCost   = errors.New("Invalid cost")
	number           = 1
)

type Order struct {
	Number  int
	Email   string
	Address string
	Status  string //created, confirmed
	Time    int    //(min)
	Cost    float64
	Check   map[Item]int //restaurant_title: item1: kolvo; item2: kolvo...
}

func NewOrder(email, address string) (error, *Order) {

	if email == "" {
		return ErrInvalidEmail, nil
	}

	if address == "" {
		return ErrInvalidAddress, nil
	}

	return nil, &Order{
		Number:  number,
		Email:   email,
		Address: address,
		Status:  "Created",
		Time:    0,
		Cost:    0,
		Check:   make(map[Item]int),
	}

}
