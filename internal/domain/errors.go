package domain

import "errors"

type ErrorResponse struct {
	Error string `json:"error"`
}

var (
	ErrInvalidRestaurantTitle = errors.New("invalid restaurant title")
	ErrInvalidComposition     = errors.New("invalid composition")
	ErrInvalidTitle           = errors.New("invalid title")
	ErrInvalidDescription     = errors.New("invalid description")
	ErrInvalidRating          = errors.New("invalid rating")
	ErrInvalidNumber          = errors.New("invalid number")
	ErrInvalidTime            = errors.New("invalid time")
	ErrInvalidCost            = errors.New("invalid cost")
	ErrInvalidEmail           = errors.New("invalid email")
	ErrInvalidName            = errors.New("invalid name")
	ErrInvalidPassword        = errors.New("invalid password")
	ErrInvalidRole            = errors.New("invalid role")
	ErrInvalidAddress         = errors.New("invalid address")
	ErrInvalidReplenish       = errors.New("invalid replenish")
	ErrInvalidPrice           = errors.New("invalid price")
	ErrInvlaidPay             = errors.New("invalid pay")

	ErrUserNotFound       = errors.New("user not found")
	ErrRestaurantNotFound = errors.New("restaurant not found")
	ErrOrderNotFound      = errors.New("order not found")
	ErrItemNotFound       = errors.New("item not found")

	ErrWithInsert = errors.New("error with saving in database")
	ErrWithUpdate = errors.New("nothing update")
	ErrWithDelete = errors.New("nothing delete")

	ErrStatusNotCreated = errors.New("order status is completed or canceled")

	ErrInvalidAccessToken   = errors.New("invalid access token")
	ErrInvalidRefreshToken  = errors.New("invalid refresh token")
	ErrInvalidAuthorization = errors.New("invalid email or password")
)

const (
	StatusCreated   = "Created"
	StatusCanceled  = "Canceled"
	StatusCompleted = "Completed"

	UserRole  = "User"
	AdminRole = "Admin"
)
