package domain

import "errors"

var (
	ErrInvalidRestaurantTitle = errors.New("Invalid restaurant title")
	ErrInvalidComposition     = errors.New("Invalid composition")
)

type Item struct {
	Title       string
	RestTitle   string
	Composition string
	Time        int
	Cost        float64
}

func NewItem(title, composition, r_title string, time int, cost float64) (error, *Item) {

	if title == "" {
		return ErrInvalidTitle, nil
	}
	if composition == "" {
		return ErrInvalidComposition, nil
	}
	if cost < 0 {
		return ErrInvalidCost, nil
	}
	if time < 0 {
		return ErrInvalidTime, nil
	}
	if r_title == "" {
		return ErrInvalidRestaurantTitle, nil
	}
	return nil, &Item{
		Title:       title,
		RestTitle:   r_title,
		Composition: composition,
		Time:        time,
		Cost:        cost,
	}

}
