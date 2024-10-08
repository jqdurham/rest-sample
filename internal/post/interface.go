package post

//go:generate mockery --name=Servicer
type Servicer interface {
	ListPosts() []Post
	GetPost(id int64) (*Post, error)
	CreatePost(pst *Post) (*Post, error)
	UpdatePost(id int64, pst *Post) (*Post, error)
	DeletePost(id int64) error
}
