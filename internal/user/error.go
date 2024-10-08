package user

import "fmt"

type NotFoundError struct {
	id int64
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("user with id %d not found", e.id)
}

type InvalidError struct {
	message string
}

func (e InvalidError) Error() string {
	return e.message
}
