package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jqdurham/rest-sample/internal/api/oapi"
	"github.com/jqdurham/rest-sample/internal/post"
	"github.com/jqdurham/rest-sample/internal/user"
)

// compile time verification that ServerHandler is compatible with REST API interface.
var _ oapi.ServerInterface = &ServerHandler{}

type ServerHandler struct {
	userSvc user.Servicer
	postSvc post.Servicer
}

func NewServerHandler(userSvc user.Servicer, postSvc post.Servicer) *ServerHandler {
	return &ServerHandler{userSvc: userSvc, postSvc: postSvc}
}

func success(w http.ResponseWriter, code int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

func noContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func notFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

func serverError(w http.ResponseWriter, err error, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	slog.Error(err.Error())
	_ = json.NewEncoder(w).Encode(msg)
}

func unprocessableRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnprocessableEntity)
}

func badRequest(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(msg)
}
