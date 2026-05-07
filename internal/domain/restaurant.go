package domain

type Restaurant struct {
	Title       string
	Description string
	Address     string
	Rating      float64
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
