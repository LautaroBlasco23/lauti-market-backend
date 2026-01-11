package infrastructure

import (
	"errors"
	"net/http"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
)

func HandleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrStoreNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrInvalidName),
		errors.Is(err, domain.ErrInvalidDescription),
		errors.Is(err, domain.ErrInvalidAddress),
		errors.Is(err, domain.ErrInvalidPhoneNumber):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
