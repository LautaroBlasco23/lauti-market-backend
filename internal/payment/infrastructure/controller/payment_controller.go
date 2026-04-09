package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	apiInfra "github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/payment/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/payment/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/payment/infrastructure/dto"
)

type PaymentController struct {
	service       *application.PaymentService
	webhookSecret string
}

func NewPaymentController(service *application.PaymentService, webhookSecret string) *PaymentController {
	return &PaymentController{
		service:       service,
		webhookSecret: webhookSecret,
	}
}

func toPaymentResponse(p *domain.Payment) dto.PaymentResponse {
	return dto.PaymentResponse{
		ID:            p.ID(),
		OrderID:       p.OrderID(),
		UserID:        p.UserID(),
		MPPaymentID:   p.MPPaymentID(),
		Amount:        p.Amount(),
		Currency:      p.Currency(),
		Status:        string(p.Status()),
		StatusDetail:  p.StatusDetail(),
		PaymentMethod: p.PaymentMethod(),
		CreatedAt:     p.CreatedAt(),
		UpdatedAt:     p.UpdatedAt(),
	}
}

func (c *PaymentController) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := apiInfra.GetClaims(r.Context())
	if !ok {
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	if string(claims.AccountType) != "user" {
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}

	var req dto.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := apiInfra.Validate(req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"error":  "invalid_payload",
			"fields": apiInfra.FieldErrors(err),
		})
		return
	}

	p, err := c.service.CreatePayment(r.Context(), application.CreatePaymentInput{
		OrderID:      req.OrderID,
		UserID:       claims.AccountID,
		CardToken:    req.CardToken,
		PayerEmail:   req.PayerEmail,
		Installments: req.Installments,
	})
	if err != nil {
		c.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(toPaymentResponse(p)) //nolint:errcheck
}

func (c *PaymentController) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing payment id", http.StatusBadRequest)
		return
	}

	claims, ok := apiInfra.GetClaims(r.Context())
	if !ok {
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	p, err := c.service.GetByID(r.Context(), id, claims.AccountID)
	if err != nil {
		c.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toPaymentResponse(p)) //nolint:errcheck
}

func (c *PaymentController) GetByOrderID(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("order_id")
	if orderID == "" {
		http.Error(w, "missing order id", http.StatusBadRequest)
		return
	}

	claims, ok := apiInfra.GetClaims(r.Context())
	if !ok {
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	p, err := c.service.GetByOrderID(r.Context(), orderID, claims.AccountID)
	if err != nil {
		c.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toPaymentResponse(p)) //nolint:errcheck
}

// webhookPayload represents the MP webhook notification body.
type webhookPayload struct {
	Action string `json:"action"`
	Data   struct {
		ID string `json:"id"`
	} `json:"data"`
	Type string `json:"type"`
}

func (c *PaymentController) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Always return 200 to prevent MP retries, even on errors.
	var payload webhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	if payload.Type != "payment" || payload.Data.ID == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Validate HMAC-SHA256 signature.
	sig := r.Header.Get("x-signature")
	requestID := r.Header.Get("x-request-id")
	if sig != "" && c.webhookSecret != "" {
		if err := validateWebhookSignature(sig, payload.Data.ID, requestID, c.webhookSecret); err != nil {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	mpPaymentID, err := strconv.ParseInt(payload.Data.ID, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	_ = c.service.HandleWebhook(r.Context(), application.WebhookInput{ //nolint:errcheck
		MPPaymentID: mpPaymentID,
	})

	w.WriteHeader(http.StatusOK)
}

// validateWebhookSignature validates the x-signature header from Mercado Pago.
// The header format is: ts=<timestamp>,v1=<hash>
// The signed string is: id:<payment_id>;request-id:<x-request-id>;ts:<timestamp>;
func validateWebhookSignature(sigHeader, paymentID, requestID, secret string) error {
	parts := strings.Split(sigHeader, ",")
	var ts, v1 string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "ts=") {
			ts = strings.TrimPrefix(part, "ts=")
		} else if strings.HasPrefix(part, "v1=") {
			v1 = strings.TrimPrefix(part, "v1=")
		}
	}

	if ts == "" || v1 == "" {
		return apiDomain.ErrInvalidWebhookSig
	}

	manifest := fmt.Sprintf("id:%s;request-id:%s;ts:%s;", paymentID, requestID, ts)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(manifest))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(v1)) {
		return apiDomain.ErrInvalidWebhookSig
	}

	return nil
}

func (c *PaymentController) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, apiDomain.ErrPaymentNotFound),
		errors.Is(err, apiDomain.ErrOrderNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, apiDomain.ErrPaymentAlreadyExists):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, apiDomain.ErrPaymentFailed):
		http.Error(w, err.Error(), http.StatusBadGateway)
	case errors.Is(err, apiDomain.ErrInvalidPaymentAmount):
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	case errors.Is(err, apiDomain.ErrForbidden):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, apiDomain.ErrUnauthorized):
		http.Error(w, err.Error(), http.StatusUnauthorized)
	case errors.Is(err, apiDomain.ErrInvalidWebhookSig):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
