package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jqdurham/rest-sample/internal/api/oapi"
	"github.com/jqdurham/rest-sample/internal/post"
)

func (s *ServerHandler) ListPosts(w http.ResponseWriter, _ *http.Request) {
	posts := s.postSvc.ListPosts()

	body := make([]*oapi.Post, len(posts))
	for i, p := range posts {
		body[i] = toAPIPost(&p)
	}

	success(w, http.StatusOK, body)
}

func (s *ServerHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var postInput oapi.PostInput
	if err := json.NewDecoder(r.Body).Decode(&postInput); err != nil {
		unprocessableRequest(w)
		return
	}

	body, err := s.postSvc.CreatePost(toPost(postInput))
	if err != nil {
		var vf *post.InvalidError
		if errors.As(err, &vf) {
			badRequest(w, vf.Error())
			return
		}
		serverError(w, err, "unable to create post")
		return
	}

	success(w, http.StatusCreated, toAPIPost(body))
}

func (s *ServerHandler) DeletePost(w http.ResponseWriter, _ *http.Request, id int64) {
	if _, err := s.postSvc.GetPost(id); err != nil {
		var nf *post.NotFoundError
		if errors.As(err, &nf) {
			notFound(w)
			return
		}
		serverError(w, err, "unable to locate post")
		return
	}

	if err := s.postSvc.DeletePost(id); err != nil {
		serverError(w, err, "unable to delete post")
		return
	}

	noContent(w)
}

func (s *ServerHandler) GetPost(w http.ResponseWriter, _ *http.Request, id int64) {
	pst, err := s.postSvc.GetPost(id)
	if err != nil {
		var nf *post.NotFoundError
		if errors.As(err, &nf) {
			notFound(w)
			return
		}
		serverError(w, err, "unable to locate post")
		return
	}

	success(w, http.StatusOK, toAPIPost(pst))
}

func (s *ServerHandler) UpdatePost(w http.ResponseWriter, r *http.Request, id int64) {
	var postInput oapi.PostInput
	if err := json.NewDecoder(r.Body).Decode(&postInput); err != nil {
		unprocessableRequest(w)
		return
	}

	pst, err := s.postSvc.UpdatePost(id, toPost(postInput))
	if err != nil {
		var nf *post.NotFoundError
		if errors.As(err, &nf) {
			notFound(w)
			return
		}
		var vf *post.InvalidError
		if errors.As(err, &vf) {
			badRequest(w, vf.Error())
			return
		}
		serverError(w, err, "unable to update post")
		return
	}

	success(w, http.StatusOK, toAPIPost(pst))
}

func toPost(input oapi.PostInput) *post.Post {
	return &post.Post{
		Content: input.Content,
		Title:   input.Title,
		UserID:  input.UserId,
	}
}

func toAPIPost(pst *post.Post) *oapi.Post {
	return &oapi.Post{
		Id:      &pst.ID,
		Title:   pst.Title,
		Content: pst.Content,
		UserId:  pst.UserID,
	}
}
