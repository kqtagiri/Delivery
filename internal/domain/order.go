package domain

type Order struct {
	Number  int
	Email   string
	Address string
	Status  string //created, confirmed
	Time    int    //(min)
	Cost    float64
	Check   map[Item]int //restaurant_title: item1: kolvo; item2: kolvo...
}

func NewOrder(email, address string) (*Order, error) {

	if email == "" {
		return nil, ErrInvalidEmail
	}

	if address == "" {
		return nil, ErrInvalidAddress
	}

	return &Order{
		Number:  0,
		Email:   email,
		Address: address,
		Status:  StatusCreated,
		Time:    0,
		Cost:    0,
		Check:   make(map[Item]int),
	}, nil

}
