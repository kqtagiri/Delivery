package domain

import "errors"

var (
	ErrInvalidTitle       = errors.New("Invalid title")
	ErrInvalidDescription = errors.New("Invalid description")
	ErrInvalidRating      = errors.New("Invalid rating")
)

type Restaurant struct {
	Title       string
	Description string
	Address     string
	Rating      float64
}

func NewRestaurant(title, description, address string, rating float64) (error, *Restaurant) {

	if title == "" {
		return ErrInvalidTitle, nil
	}

	if description == "" {
		return ErrInvalidDescription, nil
	}

	if address == "" {
		return ErrInvalidAddress, nil
	}

	if rating > 5 || rating < 0 {
		return ErrInvalidRating, nil
	}

	return nil, &Restaurant{
		Title:       title,
		Description: description,
		Address:     address,
		Rating:      rating,
	}

}
