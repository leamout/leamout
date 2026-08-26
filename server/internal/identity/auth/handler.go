package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/httputil"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Start(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req startRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(
			w,
			apperror.NewBadRequest("invalid request body"),
		)
		return
	}

	transaction, err := h.service.Start(
		r.Context(),
		normalizeEmail(req.Email),
	)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	httputil.OK(
		w,
		startResponse{
			TransactionID: transaction.ID,
			Methods:       authenticationMethods(),
		},
	)
}

func (h *Handler) LoginWithPassword(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req passwordLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(
			w,
			apperror.NewBadRequest("invalid request body"),
		)
		return
	}

	if err := validatePasswordLoginRequest(req); err != nil {
		httputil.Error(w, err)
		return
	}

	user, err := h.service.LoginWithPassword(
		r.Context(),
		req.TransactionID,
		req.Password,
	)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	httputil.OK(
		w,
		authenticationResponse{
			UserID:        user.ID,
			SessionExpiry: user.SessionExpiry.Format(time.RFC3339),
		},
	)
}

func (h *Handler) SendOTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req sendOTPRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(
			w,
			apperror.NewBadRequest("invalid request body"),
		)
		return
	}

	if _, err := h.service.SendOTP(
		r.Context(),
		req.TransactionID,
	); err != nil {
		httputil.Error(w, err)
		return
	}

	httputil.OK(
		w,
		sendOTPResponse{
			TransactionID: req.TransactionID,
		},
	)
}

func (h *Handler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	// Decode request
	// Call service
	// Write response
}

func (h *Handler) SetPassword(w http.ResponseWriter, r *http.Request) {
	// Decode request
	// Call service
	// Write response
}

func authenticationMethods() []string { return []string{"password", "otp"} }
