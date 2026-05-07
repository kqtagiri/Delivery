package domain

type Item struct {
	Title       string
	RestTitle   string
	Composition string
	Time        int
	Cost        float64
}

func NewItem(title, composition, r_title string, time int, cost float64) (*Item, error) {

	if title == "" {
		return nil, ErrInvalidTitle
	}
	if composition == "" {
		return nil, ErrInvalidComposition
	}
	if cost < 0 {
		return nil, ErrInvalidCost
	}
	if time < 0 {
		return nil, ErrInvalidTime
	}
	if r_title == "" {
		return nil, ErrInvalidRestaurantTitle
	}
	return &Item{
		Title:       title,
		RestTitle:   r_title,
		Composition: composition,
		Time:        time,
		Cost:        cost,
	}, nil

}
