package domain

type Item struct {
	ID          int     `json:"id"`
	RestID      int     `json:"rest_id"`
	Title       string  `json:"title"`
	Composition string  `json:"composition"`
	Time        int     `json:"time"`
	Cost        float64 `json:"cost"`
}

func NewItem(title, composition string, time int, cost float64) (*Item, error) {

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

	return &Item{
		ID:          0,
		RestID:      0,
		Title:       title,
		Composition: composition,
		Time:        time,
		Cost:        cost,
	}, nil

}
