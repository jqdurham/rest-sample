package user

//go:generate mockery --name=Servicer
type Servicer interface {
	ListUsers() []User
	GetUser(id int64) (*User, error)
	CreateUser(usr *User) (*User, error)
	UpdateUser(id int64, usr *User) (*User, error)
	DeleteUser(id int64) error
}
