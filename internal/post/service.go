package post

import (
	"cmp"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/jqdurham/rest-sample/internal/user"
)

// compile time check to make sure Servicer interface is satisfied.
var _ Servicer = &Service{}

type Service struct {
	mu      sync.RWMutex
	cache   map[int64]Post
	lastID  atomic.Int64
	userSvc user.Servicer
}

func NewService(userSvc user.Servicer) *Service {
	return &Service{
		cache:   make(map[int64]Post),
		userSvc: userSvc,
	}
}

func (svc *Service) ListPosts() []Post {
	out := make([]Post, 0, len(svc.cache))

	svc.mu.RLock()
	defer svc.mu.RUnlock()

	for _, v := range svc.cache {
		out = append(out, v)
	}

	slices.SortFunc(out, func(a, b Post) int {
		return cmp.Compare(a.ID, b.ID)
	})

	return out
}

func (svc *Service) GetPost(id int64) (*Post, error) {
	svc.mu.RLock()
	defer svc.mu.RUnlock()

	out, ok := svc.cache[id]
	if !ok {
		return nil, &NotFoundError{id: id}
	}

	return &out, nil
}

func (svc *Service) CreatePost(post *Post) (*Post, error) {
	if err := svc.isValidPost(post); err != nil {
		return nil, err
	}

	id := svc.lastID.Add(1)
	post.ID = id

	svc.mu.Lock()
	svc.cache[id] = *post
	svc.mu.Unlock()

	return svc.GetPost(post.ID)
}

func (svc *Service) UpdatePost(id int64, post *Post) (*Post, error) {
	if err := svc.isValidPost(post); err != nil {
		return nil, err
	}

	svc.mu.Lock()
	defer svc.mu.Unlock()

	if _, ok := svc.cache[id]; !ok {
		return nil, &NotFoundError{id: id}
	}

	post.ID = id
	svc.cache[id] = *post

	return post, nil
}

func (svc *Service) DeletePost(id int64) error {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	if _, ok := svc.cache[id]; !ok {
		return &NotFoundError{id: id}
	}

	delete(svc.cache, id)

	return nil
}

func (svc *Service) isValidPost(post *Post) error {
	defer func(start time.Time) {
		slog.Debug("validating post", slog.Duration("dur", time.Since(start)))
	}(time.Now())

	if post == nil {
		return &InvalidError{message: "post missing"}
	}

	if post.UserID < 1 {
		return &InvalidError{message: "invalid userID"}
	}

	if _, err := svc.userSvc.GetUser(post.UserID); err != nil {
		var nf *user.NotFoundError
		if errors.As(err, &nf) {
			return &InvalidError{message: "userID was not found"}
		}

		return fmt.Errorf("get user: %w", err)
	}

	titleLen := utf8.RuneCountInString(post.Title)
	if titleLen < 2 || titleLen > 200 {
		return &InvalidError{message: "title must be between 2 and 200 characters in length"}
	}

	contentLen := utf8.RuneCountInString(post.Content)
	if contentLen < 3 || contentLen > 5000 {
		return &InvalidError{message: "content must be between 3 and 5000 characters in length"}
	}

	return nil
}
