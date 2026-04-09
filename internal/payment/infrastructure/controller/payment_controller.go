package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
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
		PreferenceID:  p.PreferenceID(),
		CreatedAt:     p.CreatedAt(),
		UpdatedAt:     p.UpdatedAt(),
	}
}

func (c *PaymentController) Create(w http.ResponseWriter, r *http.Request) {
	requestID := apiInfra.GetRequestID(r)
	slog.Debug("PaymentController.Create started",
		slog.String("request_id", requestID),
	)

	claims, ok := apiInfra.GetClaims(r.Context())
	if !ok {
		slog.Warn("PaymentController.Create failed: unauthorized",
			slog.String("request_id", requestID),
		)
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	if string(claims.AccountType) != "user" {
		slog.Warn("PaymentController.Create failed: forbidden - not a user account",
			slog.String("request_id", requestID),
			slog.String("account_type", string(claims.AccountType)),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}

	var req dto.CreatePreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("PaymentController.Create failed: invalid request body",
			slog.String("request_id", requestID),
			slog.Any("error", err),
		)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := apiInfra.Validate(req); err != nil {
		slog.Warn("PaymentController.Create failed: validation error",
			slog.String("request_id", requestID),
			slog.Any("fields", apiInfra.FieldErrors(err)),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"error":  "invalid_payload",
			"fields": apiInfra.FieldErrors(err),
		})
		return
	}

	slog.Debug("PaymentController.Create calling service",
		slog.String("request_id", requestID),
		slog.String("user_id", claims.AccountID),
		slog.Int("order_count", len(req.OrderIDs)),
	)
	result, err := c.service.CreatePreference(r.Context(), application.CreatePreferenceInput{
		OrderIDs: req.OrderIDs,
		UserID:   claims.AccountID,
	})
	if err != nil {
		slog.Error("PaymentController.Create failed",
			slog.String("request_id", requestID),
			slog.Any("error", err),
		)
		c.handleError(w, err)
		return
	}

	slog.Info("PaymentController.Create completed",
		slog.String("request_id", requestID),
		slog.String("preference_id", result.PreferenceID),
		slog.String("user_id", claims.AccountID),
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(dto.CreatePreferenceResponse{ //nolint:errcheck
		PreferenceID:     result.PreferenceID,
		InitPoint:        result.InitPoint,
		SandboxInitPoint: result.SandboxInitPoint,
	})
}

func (c *PaymentController) GetByID(w http.ResponseWriter, r *http.Request) {
	requestID := apiInfra.GetRequestID(r)
	slog.Debug("PaymentController.GetByID started",
		slog.String("request_id", requestID),
	)

	id := r.PathValue("id")
	if id == "" {
		slog.Warn("PaymentController.GetByID failed: missing payment id",
			slog.String("request_id", requestID),
		)
		http.Error(w, "missing payment id", http.StatusBadRequest)
		return
	}

	claims, ok := apiInfra.GetClaims(r.Context())
	if !ok {
		slog.Warn("PaymentController.GetByID failed: unauthorized",
			slog.String("request_id", requestID),
		)
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	slog.Debug("PaymentController.GetByID calling service",
		slog.String("request_id", requestID),
		slog.String("payment_id", id),
		slog.String("account_id", claims.AccountID),
	)
	p, err := c.service.GetByID(r.Context(), id, claims.AccountID)
	if err != nil {
		slog.Error("PaymentController.GetByID failed",
			slog.String("request_id", requestID),
			slog.String("payment_id", id),
			slog.Any("error", err),
		)
		c.handleError(w, err)
		return
	}

	slog.Info("PaymentController.GetByID completed",
		slog.String("request_id", requestID),
		slog.String("payment_id", p.ID()),
	)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toPaymentResponse(p)) //nolint:errcheck
}

func (c *PaymentController) GetByOrderID(w http.ResponseWriter, r *http.Request) {
	requestID := apiInfra.GetRequestID(r)
	slog.Debug("PaymentController.GetByOrderID started",
		slog.String("request_id", requestID),
	)

	orderID := r.PathValue("order_id")
	if orderID == "" {
		slog.Warn("PaymentController.GetByOrderID failed: missing order id",
			slog.String("request_id", requestID),
		)
		http.Error(w, "missing order id", http.StatusBadRequest)
		return
	}

	claims, ok := apiInfra.GetClaims(r.Context())
	if !ok {
		slog.Warn("PaymentController.GetByOrderID failed: unauthorized",
			slog.String("request_id", requestID),
		)
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	slog.Debug("PaymentController.GetByOrderID calling service",
		slog.String("request_id", requestID),
		slog.String("order_id", orderID),
		slog.String("account_id", claims.AccountID),
	)
	p, err := c.service.GetByOrderID(r.Context(), orderID, claims.AccountID)
	if err != nil {
		slog.Error("PaymentController.GetByOrderID failed",
			slog.String("request_id", requestID),
			slog.String("order_id", orderID),
			slog.Any("error", err),
		)
		c.handleError(w, err)
		return
	}

	slog.Info("PaymentController.GetByOrderID completed",
		slog.String("request_id", requestID),
		slog.String("payment_id", p.ID()),
		slog.String("order_id", orderID),
	)
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
	requestID := r.Header.Get("x-request-id")
	if requestID == "" {
		requestID = apiInfra.GetRequestID(r)
	}
	slog.Debug("PaymentController.HandleWebhook started",
		slog.String("request_id", requestID),
	)

	// Always return 200 to prevent MP retries, even on errors.
	var payload webhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Warn("PaymentController.HandleWebhook: failed to decode payload",
			slog.String("request_id", requestID),
			slog.Any("error", err),
		)
		w.WriteHeader(http.StatusOK)
		return
	}

	if payload.Type != "payment" || payload.Data.ID == "" {
		slog.Debug("PaymentController.HandleWebhook: ignoring non-payment event",
			slog.String("request_id", requestID),
			slog.String("type", payload.Type),
		)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Validate HMAC-SHA256 signature.
	sig := r.Header.Get("x-signature")
	if sig != "" && c.webhookSecret != "" {
		if err := validateWebhookSignature(sig, payload.Data.ID, requestID, c.webhookSecret); err != nil {
			slog.Warn("PaymentController.HandleWebhook: signature validation failed",
				slog.String("request_id", requestID),
				slog.Any("error", err),
			)
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	mpPaymentID, err := strconv.ParseInt(payload.Data.ID, 10, 64)
	if err != nil {
		slog.Error("PaymentController.HandleWebhook: failed to parse payment ID",
			slog.String("request_id", requestID),
			slog.String("payment_id", payload.Data.ID),
			slog.Any("error", err),
		)
		w.WriteHeader(http.StatusOK)
		return
	}

	slog.Debug("PaymentController.HandleWebhook processing payment",
		slog.String("request_id", requestID),
		slog.Int64("mp_payment_id", mpPaymentID),
	)
	if err := c.service.HandleWebhook(r.Context(), application.WebhookInput{
		MPPaymentID: mpPaymentID,
	}); err != nil {
		slog.Error("PaymentController.HandleWebhook: failed to process webhook",
			slog.String("request_id", requestID),
			slog.Int64("mp_payment_id", mpPaymentID),
			slog.Any("error", err),
		)
	}

	slog.Info("PaymentController.HandleWebhook completed",
		slog.String("request_id", requestID),
		slog.Int64("mp_payment_id", mpPaymentID),
	)
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
	case errors.Is(err, apiDomain.ErrForbiddenTransition):
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	case errors.Is(err, apiDomain.ErrForbidden):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, apiDomain.ErrUnauthorized):
		http.Error(w, err.Error(), http.StatusUnauthorized)
	case errors.Is(err, apiDomain.ErrInvalidWebhookSig):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		slog.Error("PaymentController: internal server error", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
