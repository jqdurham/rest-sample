package post

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/jqdurham/rest-sample/internal/user"
	userMocks "github.com/jqdurham/rest-sample/internal/user/mocks"
)

var errMockedFailure = errors.New("mocked failure")

func TestService_NewService(t *testing.T) {
	t.Parallel()
	got := NewService(userMocks.NewServicer(t))
	if got == nil {
		t.Errorf("NewService() returned nil")
	}
}

func TestService_CreatePost(t *testing.T) {
	t.Parallel()
	type fields struct {
		lastID  int64
		userSvc func() user.Servicer
	}
	type args struct {
		post *Post
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Post
		errMsg string
	}{
		{
			name: "Creates first post in cache",
			fields: fields{
				lastID: 0,
				userSvc: func() user.Servicer {
					m := userMocks.NewServicer(t)
					m.On("GetUser", int64(9)).Return(nil, nil)
					return m
				},
			},
			args: args{
				post: &Post{Title: "My Post", Content: "My Content", UserID: 9},
			},
			want: &Post{ID: 1, Title: "My Post", Content: "My Content", UserID: 9},
		},
		{
			name: "Creates 1+Nth in cache",
			fields: fields{
				lastID: 1336,
				userSvc: func() user.Servicer {
					m := userMocks.NewServicer(t)
					m.On("GetUser", int64(1)).Return(nil, nil)
					return m
				},
			},
			args: args{
				post: &Post{Title: "My Post", Content: "My Content", UserID: 1},
			},
			want: &Post{ID: 1337, Title: "My Post", Content: "My Content", UserID: 1},
		},
		{
			name: "Rejects invalid post",
			args: args{
				post: &Post{Title: "My Post", Content: "My Content", UserID: 0},
			},
			errMsg: "invalid userID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &Service{
				mu:     sync.RWMutex{},
				cache:  map[int64]Post{},
				lastID: atomic.Int64{},
			}
			if tt.fields.userSvc != nil {
				svc.userSvc = tt.fields.userSvc()
			}
			svc.lastID.Store(tt.fields.lastID)
			got, err := svc.CreatePost(tt.args.post)
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("CreatePost() error = %v, errMsg %v", err, tt.errMsg)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreatePost() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_DeletePost(t *testing.T) {
	t.Parallel()
	type fields struct {
		cache map[int64]Post
	}
	type args struct {
		id int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[int64]Post
		errMsg string
	}{
		{
			name: "Successfully removes post from cache",
			fields: fields{
				cache: map[int64]Post{1: {}, 2: {}, 3: {}},
			},
			args: args{
				id: 2,
			},
			want: map[int64]Post{1: {}, 3: {}},
		},
		{
			name: "Returns NotFound error when post doesn't exist",
			fields: fields{
				cache: map[int64]Post{1: {}, 2: {}, 3: {}},
			},
			args: args{
				id: 8,
			},
			want:   map[int64]Post{1: {}, 2: {}, 3: {}},
			errMsg: "post with id 8 not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &Service{
				cache: tt.fields.cache,
			}
			err := svc.DeletePost(tt.args.id)
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("DeletePost() error = %v, errMsg %v", err, tt.errMsg)
			}
			if !reflect.DeepEqual(tt.fields.cache, tt.want) {
				t.Errorf("DeletePost() got = %v, want %v", tt.fields.cache, tt.want)
			}
		})
	}
}

func TestService_GetPost(t *testing.T) {
	t.Parallel()
	type fields struct {
		cache map[int64]Post
	}
	type args struct {
		id int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Post
		errMsg string
	}{
		{
			name: "Successfully gets post from cache",
			fields: fields{
				cache: map[int64]Post{1: {}, 2: {
					ID:      2,
					Title:   "My Post",
					Content: "My Content",
					UserID:  123,
				}, 3: {}},
			},
			args: args{
				id: 2,
			},
			want: &Post{
				ID:      2,
				Title:   "My Post",
				Content: "My Content",
				UserID:  123,
			},
		},
		{
			name: "Returns NotFound error when post doesn't exist",
			fields: fields{
				cache: map[int64]Post{1: {}, 2: {}, 3: {}},
			},
			args: args{
				id: 9,
			},
			want:   nil,
			errMsg: "post with id 9 not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &Service{
				cache: tt.fields.cache,
			}
			pst, err := svc.GetPost(tt.args.id)
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("GetPost() error = %v, errMsg %v", err, tt.errMsg)
			}
			if !reflect.DeepEqual(pst, tt.want) {
				t.Errorf("GetPost() got = %v, want %v", pst, tt.want)
			}
		})
	}
}

func TestService_ListPosts(t *testing.T) {
	t.Parallel()
	type fields struct {
		cache map[int64]Post
	}
	tests := []struct {
		name   string
		fields fields
		want   []Post
	}{
		{
			name: "Fetches posts from post service",
			fields: fields{
				cache: map[int64]Post{
					2: {ID: 2, Title: "Second", Content: "Second Content", UserID: 2},
					5: {ID: 5, Title: "Fifth", Content: "Fifth Content", UserID: 5},
					1: {ID: 1, Title: "First", Content: "First Content", UserID: 1},
				},
			},
			want: []Post{
				{ID: 1, Title: "First", Content: "First Content", UserID: 1},
				{ID: 2, Title: "Second", Content: "Second Content", UserID: 2},
				{ID: 5, Title: "Fifth", Content: "Fifth Content", UserID: 5},
			},
		},
		{
			name: "Fetches empty posts from post service",
			fields: fields{
				cache: map[int64]Post{},
			},
			want: []Post{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &Service{
				cache: tt.fields.cache,
			}
			if got := svc.ListPosts(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListPosts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_UpdatePost(t *testing.T) {
	t.Parallel()
	type fields struct {
		cache   map[int64]Post
		userSvc func() user.Servicer
	}
	type args struct {
		id   int64
		post *Post
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Post
		errMsg string
	}{
		{
			name: "Updates post in cache without updating ID",
			fields: fields{
				cache: map[int64]Post{5: {ID: 5, Title: "My Post", Content: "My Content", UserID: 9}, 2: {}},
				userSvc: func() user.Servicer {
					m := userMocks.NewServicer(t)
					m.On("GetUser", int64(10)).Return(nil, nil)
					return m
				},
			},
			args: args{
				id:   5,
				post: &Post{ID: 10, Title: "My NEW Post", Content: "My NEW Content", UserID: 10},
			},
			want: &Post{ID: 5, Title: "My NEW Post", Content: "My NEW Content", UserID: 10},
		},
		{
			name: "Returns NotFound when post doesn't exist",
			fields: fields{
				cache: map[int64]Post{},
				userSvc: func() user.Servicer {
					m := userMocks.NewServicer(t)
					m.On("GetUser", int64(10)).Return(nil, nil)
					return m
				},
			},
			args: args{
				id:   5,
				post: &Post{ID: 10, Title: "My NEW Post", Content: "My NEW Content", UserID: 10},
			},
			errMsg: "post with id 5 not found",
		},
		{
			name: "Returns validation failure when post fails validation",
			fields: fields{
				cache: map[int64]Post{},
				userSvc: func() user.Servicer {
					return userMocks.NewServicer(t)
				},
			},
			args: args{
				id:   5,
				post: &Post{ID: 10, Title: "My Post", Content: "My Content", UserID: 0},
			},
			errMsg: "invalid userID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &Service{
				cache:   tt.fields.cache,
				userSvc: tt.fields.userSvc(),
			}
			got, err := svc.UpdatePost(tt.args.id, tt.args.post)
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("UpdatePost() error = %v, errMsg %v", err, tt.errMsg)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdatePost() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_isValidPost(t *testing.T) {
	t.Parallel()
	type fields struct {
		userSvc func() user.Servicer
	}
	type args struct {
		post *Post
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		errMsg string
	}{
		{
			name: "Confirms post with all fields at minimums (3-byte char)",
			fields: fields{
				userSvc: func() user.Servicer {
					m := userMocks.NewServicer(t)
					m.On("GetUser", int64(1)).Return(nil, nil)
					return m
				},
			},
			args: args{
				post: &Post{
					ID:      1,
					Title:   strings.Repeat("a", 2),
					Content: strings.Repeat("a", 3),
					UserID:  1,
				},
			},
		},
		{
			name: "Confirms post with all fields at minimums (4-byte char)",
			fields: fields{
				userSvc: func() user.Servicer {
					m := userMocks.NewServicer(t)
					m.On("GetUser", int64(1)).Return(nil, nil)
					return m
				},
			},
			args: args{
				post: &Post{
					ID:      1,
					Title:   strings.Repeat("😀", 2),
					Content: strings.Repeat("😀", 3),
					UserID:  1,
				},
			},
		},
		{
			name: "Confirms post with all fields at maximums (3-byte char)",
			fields: fields{
				userSvc: func() user.Servicer {
					m := userMocks.NewServicer(t)
					m.On("GetUser", int64(math.MaxInt64)).Return(nil, nil)
					return m
				},
			},
			args: args{
				post: &Post{
					ID:      math.MaxInt64,
					Title:   strings.Repeat("a", 200),
					Content: strings.Repeat("a", 5000),
					UserID:  math.MaxInt64,
				},
			},
		},
		{
			name: "Confirms post with all fields at maximums (4-byte char)",
			fields: fields{
				userSvc: func() user.Servicer {
					m := userMocks.NewServicer(t)
					m.On("GetUser", int64(math.MaxInt64)).Return(nil, nil)
					return m
				},
			},
			args: args{
				post: &Post{
					ID:      math.MaxInt64,
					Title:   strings.Repeat("😀", 200),
					Content: strings.Repeat("😀", 5000),
					UserID:  math.MaxInt64,
				},
			},
		},
		{
			name:   "Rejects nil post",
			fields: fields{},
			args: args{
				post: (*Post)(nil),
			},
			errMsg: "post missing",
		},
		{
			name:   "Rejects post when userID is zero or less",
			fields: fields{},
			args: args{
				post: &Post{},
			},
			errMsg: "invalid userID",
		},
		{
			name: "Rejects post with userID that does not exist",
			fields: fields{
				userSvc: func() user.Servicer {
					m := userMocks.NewServicer(t)
					m.On("GetUser", int64(6)).Return(nil, &user.NotFoundError{})
					return m
				},
			},
			args: args{
				post: &Post{Title: "My Post", Content: "My Content", UserID: 6},
			},
			errMsg: "userID was not found",
		},
		{
			name: "Rejects post when user lookup fails",
			fields: fields{
				userSvc: func() user.Servicer {
					m := userMocks.NewServicer(t)
					m.On("GetUser", int64(5)).Return(nil, errMockedFailure)
					return m
				},
			},
			args: args{
				post: &Post{Title: "My Post", Content: "My Content", UserID: 5},
			},
			errMsg: "get user: mocked failure",
		},
		{
			name: "Rejects post when title is too short",
			fields: fields{
				userSvc: func() user.Servicer {
					m := userMocks.NewServicer(t)
					m.On("GetUser", int64(4)).Return(nil, nil)
					return m
				},
			},
			args: args{
				post: &Post{Title: "😀", Content: "My Content", UserID: 4},
			},
			errMsg: "title must be between 2 and 200 characters in length",
		},
		{
			name: "Rejects post when title is too long",
			fields: fields{
				userSvc: func() user.Servicer {
					m := userMocks.NewServicer(t)
					m.On("GetUser", int64(4)).Return(nil, nil)
					return m
				},
			},
			args: args{
				post: &Post{Title: strings.Repeat("😀", 201), Content: "My Content", UserID: 4},
			},
			errMsg: "title must be between 2 and 200 characters in length",
		},
		{
			name: "Rejects post when content is too short",
			fields: fields{
				userSvc: func() user.Servicer {
					m := userMocks.NewServicer(t)
					m.On("GetUser", int64(4)).Return(nil, nil)
					return m
				},
			},
			args: args{
				post: &Post{Title: "My Title", Content: "😀", UserID: 4},
			},
			errMsg: "content must be between 3 and 5000 characters in length",
		},
		{
			name: "Rejects post when content is too long",
			fields: fields{
				userSvc: func() user.Servicer {
					m := userMocks.NewServicer(t)
					m.On("GetUser", int64(4)).Return(nil, nil)
					return m
				},
			},
			args: args{
				post: &Post{Title: "My Title", Content: strings.Repeat("😀", 5001), UserID: 4},
			},
			errMsg: "content must be between 3 and 5000 characters in length",
		},
		{
			name: "Rejects post when content is too long",
			fields: fields{
				userSvc: func() user.Servicer {
					m := userMocks.NewServicer(t)
					m.On("GetUser", int64(4)).Return(nil, nil)
					return m
				},
			},
			args: args{
				post: &Post{Title: "My Title", Content: strings.Repeat("😀", 5001), UserID: 4},
			},
			errMsg: "content must be between 3 and 5000 characters in length",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &Service{}
			if tt.fields.userSvc != nil {
				svc.userSvc = tt.fields.userSvc()
			}
			err := svc.isValidPost(tt.args.post)

			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("isValidPost() error = %v, errMsg %v", err, tt.errMsg)
				return
			}
		})
	}
}
