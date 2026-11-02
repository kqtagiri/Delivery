package domain

type Order struct {
	ID      int
	Number  int
	Email   string
	Address string
	Status  string //created, confirmed
	Time    int    //(min)
	Cost    float64
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
	}, nil

}

type OrderItem struct {
	ID          int     `json:"id"`
	Number      int     `json:"number"`
	MenuID      int     `json:"menu_id"`
	Title       string  `json:"title"`
	RestTitle   string  `json:"rest_title"`
	Composition string  `json:"composition"`
	Cost        float64 `json:"cost"`
}
