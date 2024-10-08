package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jqdurham/rest-sample/internal/api/oapi"
	"github.com/jqdurham/rest-sample/internal/user"
)

func (s *ServerHandler) ListUsers(w http.ResponseWriter, _ *http.Request) {
	users := s.userSvc.ListUsers()

	body := make([]*oapi.User, len(users))
	for i, u := range users {
		body[i] = toAPIUser(&u)
	}

	success(w, http.StatusOK, body)
}

func (s *ServerHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var userInput oapi.UserInput
	if err := json.NewDecoder(r.Body).Decode(&userInput); err != nil {
		unprocessableRequest(w)
		return
	}

	usr, err := s.userSvc.CreateUser(toUser(userInput))
	if err != nil {
		var vf *user.InvalidError
		if errors.As(err, &vf) {
			badRequest(w, vf.Error())
			return
		}
		serverError(w, err, "unable to create user")
		return
	}

	success(w, http.StatusCreated, toAPIUser(usr))
}

func (s *ServerHandler) DeleteUser(w http.ResponseWriter, _ *http.Request, id int64) {
	if _, err := s.userSvc.GetUser(id); err != nil {
		var nf *user.NotFoundError
		if errors.As(err, &nf) {
			notFound(w)
			return
		}
		serverError(w, err, "unable to locate user")
		return
	}

	if err := s.userSvc.DeleteUser(id); err != nil {
		serverError(w, err, "unable to delete user")
		return
	}

	noContent(w)
}

func (s *ServerHandler) GetUser(w http.ResponseWriter, _ *http.Request, id int64) {
	usr, err := s.userSvc.GetUser(id)
	if err != nil {
		var nf *user.NotFoundError
		if errors.As(err, &nf) {
			notFound(w)
			return
		}
		serverError(w, err, "unable to locate user")
		return
	}

	success(w, http.StatusOK, toAPIUser(usr))
}

func (s *ServerHandler) UpdateUser(w http.ResponseWriter, r *http.Request, id int64) {
	var userInput oapi.UserInput
	if err := json.NewDecoder(r.Body).Decode(&userInput); err != nil {
		unprocessableRequest(w)
		return
	}

	usr, err := s.userSvc.UpdateUser(id, toUser(userInput))
	if err != nil {
		var nf *user.NotFoundError
		if errors.As(err, &nf) {
			notFound(w)
			return
		}
		var vf *user.InvalidError
		if errors.As(err, &vf) {
			badRequest(w, vf.Error())
			return
		}
		serverError(w, err, "unable to update user")
		return
	}

	success(w, http.StatusOK, toAPIUser(usr))
}

func toUser(input oapi.UserInput) *user.User {
	return &user.User{
		Name:  input.Name,
		Email: input.Email,
	}
}

func toAPIUser(usr *user.User) *oapi.User {
	return &oapi.User{
		Id:    usr.ID,
		Name:  usr.Name,
		Email: usr.Email,
	}
}
