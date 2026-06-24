package server

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	authhandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/auth"
	budgethandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/budget"
	budgetfollowuphandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/budgetfollowup"
	budgetimporthandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/budgetimport"
	budgetstatushandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/budgetstatus"
	budgetstatushistoryhandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/budgetstatushistory"
	contacthandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/contact"
	conversationhandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/conversation"
	dashboardhandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/dashboard"
	estimatorhandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/estimator"
	healthhandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/health"
	installerhandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/installer"
	lossreasonhandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/lossreason"
	noticehandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/notice"
	priorityhandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/priority"
	productlinehandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/productline"
	projecthandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/project"
	projecttypehandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/projecttype"
	salespersonhandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/salesperson"
	systemtypehandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/systemtype"
	userhandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/user"
	budgetrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budget"
	budgetfollowuprepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budgetfollowup"
	budgetimportrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budgetimport"
	budgetstatusrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budgetstatus"
	budgetstatushistoryrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budgetstatushistory"
	contactrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/contact"
	conversationrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/conversation"
	dashboardrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/dashboard"
	estimatorrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/estimator"
	installerrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/installer"
	lossreasonrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/lossreason"
	noticepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/notice"
	priorityrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/priority"
	productlinerepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/productline"
	projectrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/project"
	projecttyperepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/projecttype"
	salespersonrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/salesperson"
	systemtyperepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/systemtype"
	userrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/user"
	authservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/auth"
	budgetservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/budget"
	budgetfollowupservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/budgetfollowup"
	budgetimportservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/budgetimport"
	budgetstatusservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/budgetstatus"
	budgetstatushistoryservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/budgetstatushistory"
	contactservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/contact"
	conversationservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/conversation"
	dashboardservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/dashboard"
	estimatorservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/estimator"
	installerservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/installer"
	lossreasonservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/lossreason"
	noticeservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/notice"
	priorityservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/priority"
	productlineservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/productline"
	projectservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/project"
	projecttypeservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/projecttype"
	salespersonservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/salesperson"
	systemtypeservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/systemtype"
	userservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/user"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/config"
	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/middleware"
)

const healthCheckTimeout = 2 * time.Second
const maxMultipartMemory = 10 << 20

type Dependencies struct {
	DB            *sql.DB
	HealthChecker healthhandler.Checker
	Config        *config.Config
}

func NewRouter(validate *validator.Validate, deps Dependencies) *gin.Engine {
	if validate == nil {
		validate = validator.New()
	}

	allowedOrigins := []string(nil)
	if deps.Config != nil {
		allowedOrigins = deps.Config.AllowedOrigins
	}

	if deps.Config != nil {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())
	router.MaxMultipartMemory = maxMultipartMemory
	router.Use(securityHeadersMiddleware())
	router.Use(corsMiddleware(allowedOrigins))

	healthhandler.NewHandler(router, deps.HealthChecker, healthCheckTimeout).RouteList()

	if deps.DB == nil || deps.Config == nil {
		return router
	}

	userRepo := userrepository.NewRepository(deps.DB)
	budgetRepo := budgetrepository.NewRepository(deps.DB)
	budgetFollowUpRepo := budgetfollowuprepository.NewRepository(deps.DB)
	budgetImportRepo := budgetimportrepository.NewRepository(deps.DB)
	budgetStatusRepo := budgetstatusrepository.NewRepository(deps.DB)
	budgetStatusHistoryRepo := budgetstatushistoryrepository.NewRepository(deps.DB)
	installerRepo := installerrepository.NewRepository(deps.DB)
	contactRepo := contactrepository.NewRepository(deps.DB)
	dashboardRepo := dashboardrepository.NewRepository(deps.DB)
	estimatorRepo := estimatorrepository.NewRepository(deps.DB)
	lossReasonRepo := lossreasonrepository.NewRepository(deps.DB)
	noticeRepo := noticepository.NewRepository(deps.DB)
	conversationRepo := conversationrepository.NewRepository(deps.DB)
	priorityRepo := priorityrepository.NewRepository(deps.DB)
	productLineRepo := productlinerepository.NewRepository(deps.DB)
	projectTypeRepo := projecttyperepository.NewRepository(deps.DB)
	projectRepo := projectrepository.NewRepository(deps.DB)
	salespersonRepo := salespersonrepository.NewRepository(deps.DB)
	systemTypeRepo := systemtyperepository.NewRepository(deps.DB)

	authService := authservice.NewService(userRepo, deps.Config)
	userService := userservice.NewService(userRepo)
	budgetService := budgetservice.NewService(budgetRepo, budgetStatusRepo, priorityRepo, userRepo, salespersonRepo, estimatorRepo)
	budgetImportService := budgetimportservice.NewService(
		budgetRepo,
		budgetStatusRepo,
		priorityRepo,
		installerRepo,
		projectRepo,
		projectTypeRepo,
		salespersonRepo,
		contactRepo,
		lossReasonRepo,
		budgetImportRepo,
		productLineRepo,
	)
	budgetFollowUpService := budgetfollowupservice.NewService(budgetFollowUpRepo, budgetRepo, userRepo, salespersonRepo, estimatorRepo)
	budgetStatusService := budgetstatusservice.NewService(budgetStatusRepo)
	budgetStatusHistoryService := budgetstatushistoryservice.NewService(budgetStatusHistoryRepo, budgetRepo, budgetStatusRepo, userRepo, salespersonRepo, estimatorRepo)
	installerService := installerservice.NewService(installerRepo)
	contactService := contactservice.NewService(contactRepo, installerRepo)
	dashboardService := dashboardservice.NewService(dashboardRepo, userRepo, salespersonRepo, estimatorRepo)
	estimatorService := estimatorservice.NewService(estimatorRepo, userRepo)
	lossReasonService := lossreasonservice.NewService(lossReasonRepo)
	noticeService := noticeservice.NewService(noticeRepo, userRepo)
	conversationService := conversationservice.NewService(conversationRepo, projectRepo, userRepo)
	priorityService := priorityservice.NewService(priorityRepo)
	productLineService := productlineservice.NewService(productLineRepo)
	projectTypeService := projecttypeservice.NewService(projectTypeRepo)
	projectService := projectservice.NewService(projectRepo, projectTypeRepo)
	salespersonService := salespersonservice.NewService(salespersonRepo)
	systemTypeService := systemtypeservice.NewService(systemTypeRepo)

	authhandler.NewHandler(router, validate, authService, deps.Config).RouteList()
	userhandler.NewHandler(router, validate, userService, deps.Config.SecretJWT).RouteList()
	budgethandler.NewHandler(router, validate, budgetService, deps.Config.SecretJWT).RouteList()
	budgetimporthandler.NewHandler(router, validate, budgetImportService, deps.Config.SecretJWT).RouteList()
	budgetfollowuphandler.NewHandler(router, validate, budgetFollowUpService, deps.Config.SecretJWT).RouteList()
	budgetstatushandler.NewHandler(router, validate, budgetStatusService, deps.Config.SecretJWT).RouteList()
	budgetstatushistoryhandler.NewHandler(router, validate, budgetStatusHistoryService, deps.Config.SecretJWT).RouteList()
	installerhandler.NewHandler(router, validate, installerService, deps.Config.SecretJWT).RouteList()
	contacthandler.NewHandler(router, validate, contactService, deps.Config.SecretJWT).RouteList()
	dashboardhandler.NewHandler(router, dashboardService, deps.Config.SecretJWT).RouteList()
	estimatorhandler.NewHandler(router, validate, estimatorService, deps.Config.SecretJWT).RouteList()
	lossreasonhandler.NewHandler(router, validate, lossReasonService, deps.Config.SecretJWT).RouteList()
	noticehandler.NewHandler(router, validate, noticeService, deps.Config.SecretJWT).RouteList()
	conversationhandler.NewHandler(router, validate, conversationService, deps.Config.SecretJWT).RouteList()
	priorityhandler.NewHandler(router, validate, priorityService, deps.Config.SecretJWT).RouteList()
	productlinehandler.NewHandler(router, productLineService, deps.Config.SecretJWT).RouteList()
	projecttypehandler.NewHandler(router, validate, projectTypeService, deps.Config.SecretJWT).RouteList()
	projecthandler.NewHandler(router, validate, projectService, deps.Config.SecretJWT).RouteList()
	salespersonhandler.NewHandler(router, validate, salespersonService, deps.Config.SecretJWT).RouteList()
	systemtypehandler.NewHandler(router, validate, systemTypeService, deps.Config.SecretJWT).RouteList()

	return router
}

func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	normalizedAllowedOrigins := normalizeOrigins(allowedOrigins)

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && originAllowed(origin, normalizedAllowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")

		c.Next()
	}
}

func normalizeOrigins(origins []string) map[string]struct{} {
	values := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		normalizedOrigin := strings.TrimSpace(origin)
		if normalizedOrigin != "" {
			values[normalizedOrigin] = struct{}{}
		}
	}

	return values
}

func originAllowed(origin string, allowedOrigins map[string]struct{}) bool {
	_, exists := allowedOrigins[strings.TrimSpace(origin)]
	return exists
}
