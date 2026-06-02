package domain

type Item struct {
	Title       string  `json:"title"`
	RestTitle   string  `json:"rest_title"`
	Composition string  `json:"composition"`
	Time        int     `json:"time"`
	Cost        float64 `json:"cost"`
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
