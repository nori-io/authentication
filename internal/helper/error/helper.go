package error

import (
	"net/http"

	"github.com/nori-plugins/authentication/pkg/errors"
)

func (e errorHelper) Error(w http.ResponseWriter, err error) {
	e.logger.Error("%s", err)

	switch e := err.(type) {
	case errors.Error:
		switch e.Type {
		case errors.ErrValidation:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.ErrUnauthorized:
			http.Error(w, err.Error(), http.StatusUnauthorized)
		case errors.ErrForbidden:
			http.Error(w, err.Error(), http.StatusForbidden)
		case errors.ErrNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.ErrAlreadyExists:
			http.Error(w, err.Error(), http.StatusConflict)
		case errors.ErrInternal:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
