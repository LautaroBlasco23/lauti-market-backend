package infrastructure

import (
	"database/sql"
	"net/http"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	apiInfra "github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	orderDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/order/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/payment/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/payment/infrastructure/controller"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/payment/infrastructure/mp"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/payment/infrastructure/repository"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/payment/infrastructure/routes"
)

type PaymentModule struct {
	Repository *repository.PaymentPostgresRepository
	Service    *application.PaymentService
	Controller *controller.PaymentController
}

func Wire(
	mux *http.ServeMux,
	db *sql.DB,
	idGen apiDomain.IDGenerator,
	orderRepo orderDomain.Repository,
	authMw *apiInfra.AuthMiddleware,
	mpAccessToken string,
	mpWebhookSecret string,
	frontendBaseURL string,
	notificationURL string,
) *PaymentModule {
	mpClient, err := mp.NewMPClient(mpAccessToken)
	if err != nil {
		panic("failed to create MP client: " + err.Error())
	}

	repo := repository.NewPaymentPostgresRepository(db)
	service := application.NewPaymentService(repo, orderRepo, mpClient, idGen, application.Config{
		FrontendBaseURL: frontendBaseURL,
		NotificationURL: notificationURL,
	})
	ctrl := controller.NewPaymentController(service, mpWebhookSecret)

	routes.RegisterRoutes(mux, ctrl, authMw)

	return &PaymentModule{
		Repository: repo,
		Service:    service,
		Controller: ctrl,
	}
}
