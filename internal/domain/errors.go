package domain

import "errors"

var (
	ErrInvalidRestaurantTitle = errors.New("Invalid restaurant title")
	ErrInvalidComposition     = errors.New("Invalid composition")
	ErrInvalidTitle           = errors.New("Invalid title")
	ErrInvalidDescription     = errors.New("Invalid description")
	ErrInvalidRating          = errors.New("Invalid rating")
	ErrInvalidNumber          = errors.New("Invalid number")
	ErrInvalidTime            = errors.New("Invalid time")
	ErrInvalidCost            = errors.New("Invalid cost")
	ErrInvalidEmail           = errors.New("Invalid email")
	ErrInvalidName            = errors.New("Invalid name")
	ErrInvalidPassword        = errors.New("Invalid password")
	ErrInvalidRole            = errors.New("Invalid role")
	ErrInvalidAddress         = errors.New("Invalid address")
	ErrInvalidReplenish       = errors.New("Invalid replenish")
	ErrInvalidPrice           = errors.New("Invalid price")
	ErrInvlaidPay             = errors.New("Invalid pay")

	ErrUserNotFound       = errors.New("User not found")
	ErrRestaurantNotFound = errors.New("Restaurant not found")
	ErrOrderNotFound      = errors.New("Order not found")
	ErrItemNotFound       = errors.New("Item not found")

	ErrWithInsert = errors.New("Error with saving in database")
	ErrWithUpdate = errors.New("Nothing update")
	ErrWithDelete = errors.New("Nothing delete")

	ErrStatusNotCreated = errors.New("Order status is completed or canceled")

	ErrInvalidAccessToken   = errors.New("Invalid access token")
	ErrInvalidRefreshToken  = errors.New("Invalid refresh token")
	ErrInvalidAuthorization = errors.New("Invalid email or password")
)

const (
	StatusCreated   = "Created"
	StatusCanceled  = "Canceled"
	StatusCompleted = "Completed"

	UserRole  = "User"
	AdminRole = "Admin"
)
