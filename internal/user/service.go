package user

import (
	"cmp"
	"log/slog"
	"regexp"
	"slices"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

// compile time check to make sure Servicer interface is satisfied.
var _ Servicer = &Service{}

type Service struct {
	mu     sync.RWMutex
	cache  map[int64]User
	lastID atomic.Int64
}

func NewService() *Service {
	return &Service{
		cache: make(map[int64]User),
	}
}

func (svc *Service) ListUsers() []User {
	out := make([]User, 0, len(svc.cache))

	svc.mu.RLock()
	defer svc.mu.RUnlock()

	for _, v := range svc.cache {
		out = append(out, v)
	}

	slices.SortFunc(out, func(a, b User) int {
		return cmp.Compare(a.ID, b.ID)
	})

	return out
}

func (svc *Service) GetUser(id int64) (*User, error) {
	svc.mu.RLock()
	defer svc.mu.RUnlock()

	out, ok := svc.cache[id]
	if !ok {
		return nil, &NotFoundError{id: id}
	}

	return &out, nil
}

func (svc *Service) CreateUser(user *User) (*User, error) {
	if err := svc.isValidUser(user); err != nil {
		return nil, err
	}

	id := svc.lastID.Add(1)
	user.ID = id

	svc.mu.Lock()
	svc.cache[id] = *user
	svc.mu.Unlock()

	return svc.GetUser(user.ID)
}

func (svc *Service) UpdateUser(id int64, user *User) (*User, error) {
	if err := svc.isValidUser(user); err != nil {
		return nil, err
	}

	svc.mu.Lock()
	defer svc.mu.Unlock()

	if _, ok := svc.cache[id]; !ok {
		return nil, &NotFoundError{id: id}
	}

	user.ID = id
	svc.cache[id] = *user

	return user, nil
}

func (svc *Service) DeleteUser(id int64) error {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	if _, ok := svc.cache[id]; !ok {
		return &NotFoundError{id: id}
	}

	delete(svc.cache, id)

	return nil
}

func (svc *Service) isValidUser(user *User) error {
	defer func(start time.Time) {
		slog.Debug("validating user", slog.Duration("dur", time.Since(start)))
	}(time.Now())

	if user == nil {
		return &InvalidError{message: "user missing"}
	}

	nameLen := utf8.RuneCountInString(user.Name)
	if nameLen < 3 || nameLen > 5000 {
		return &InvalidError{message: "name must be between 3 and 5000 characters in length"}
	}

	emailLen := utf8.RuneCountInString(user.Email)
	if emailLen < 2 || emailLen > 200 {
		return &InvalidError{message: "email must be between 3 and 200 characters in length"}
	}

	emailRE := regexp.MustCompile(`\w+([-+.']\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`)
	if !emailRE.MatchString(user.Email) {
		return &InvalidError{message: "email appears to be invalid"}
	}

	return nil
}
