package domain

type Restaurant struct {
	Title       string  `json:"title"`
	Description string  `json:"desc"`
	Address     string  `json:"address"`
	Rating      float64 `json:"rating"`
}

func NewRestaurant(title, description, address string, rating float64) (*Restaurant, error) {

	if title == "" {
		return nil, ErrInvalidTitle
	}

	if description == "" {
		return nil, ErrInvalidDescription
	}

	if address == "" {
		return nil, ErrInvalidAddress
	}

	if rating > 5 || rating < 0 {
		return nil, ErrInvalidRating
	}

	return &Restaurant{
		Title:       title,
		Description: description,
		Address:     address,
		Rating:      rating,
	}, nil

}
