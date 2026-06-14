package server

import (
	"database/sql"
	"net/http"
	"time"

	authhandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/auth"
	budgethandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/budget"
	budgetfollowuphandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/budgetfollowup"
	budgetstatushandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/budgetstatus"
	budgetstatushistoryhandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/budgetstatushistory"
	contacthandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/contact"
	healthhandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/health"
	installerhandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/installer"
	lossreasonhandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/lossreason"
	priorityhandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/priority"
	projecthandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/project"
	projecttypehandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/projecttype"
	salespersonhandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/salesperson"
	userhandler "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/handler/user"
	budgetrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budget"
	budgetfollowuprepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budgetfollowup"
	budgetstatusrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budgetstatus"
	budgetstatushistoryrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/budgetstatushistory"
	contactrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/contact"
	installerrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/installer"
	lossreasonrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/lossreason"
	priorityrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/priority"
	projectrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/project"
	projecttyperepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/projecttype"
	salespersonrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/salesperson"
	userrepository "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/repository/user"
	authservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/auth"
	budgetservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/budget"
	budgetfollowupservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/budgetfollowup"
	budgetstatusservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/budgetstatus"
	budgetstatushistoryservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/budgetstatushistory"
	contactservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/contact"
	installerservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/installer"
	lossreasonservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/lossreason"
	priorityservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/priority"
	projectservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/project"
	projecttypeservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/projecttype"
	salespersonservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/salesperson"
	userservice "github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/service/user"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/config"
)

const healthCheckTimeout = 2 * time.Second

type Dependencies struct {
	DB            *sql.DB
	HealthChecker healthhandler.Checker
	Config        *config.Config
}

func NewRouter(validate *validator.Validate, deps Dependencies) *gin.Engine {
	if validate == nil {
		validate = validator.New()
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	healthhandler.NewHandler(router, deps.HealthChecker, healthCheckTimeout).RouteList()

	if deps.DB == nil || deps.Config == nil {
		return router
	}

	userRepo := userrepository.NewRepository(deps.DB)
	budgetRepo := budgetrepository.NewRepository(deps.DB)
	budgetFollowUpRepo := budgetfollowuprepository.NewRepository(deps.DB)
	budgetStatusRepo := budgetstatusrepository.NewRepository(deps.DB)
	budgetStatusHistoryRepo := budgetstatushistoryrepository.NewRepository(deps.DB)
	installerRepo := installerrepository.NewRepository(deps.DB)
	contactRepo := contactrepository.NewRepository(deps.DB)
	lossReasonRepo := lossreasonrepository.NewRepository(deps.DB)
	priorityRepo := priorityrepository.NewRepository(deps.DB)
	projectTypeRepo := projecttyperepository.NewRepository(deps.DB)
	projectRepo := projectrepository.NewRepository(deps.DB)
	salespersonRepo := salespersonrepository.NewRepository(deps.DB)

	authService := authservice.NewService(userRepo, deps.Config)
	userService := userservice.NewService(userRepo)
	budgetService := budgetservice.NewService(budgetRepo)
	budgetFollowUpService := budgetfollowupservice.NewService(budgetFollowUpRepo, budgetRepo)
	budgetStatusService := budgetstatusservice.NewService(budgetStatusRepo)
	budgetStatusHistoryService := budgetstatushistoryservice.NewService(budgetStatusHistoryRepo, budgetRepo, budgetStatusRepo)
	installerService := installerservice.NewService(installerRepo)
	contactService := contactservice.NewService(contactRepo, installerRepo)
	lossReasonService := lossreasonservice.NewService(lossReasonRepo)
	priorityService := priorityservice.NewService(priorityRepo)
	projectTypeService := projecttypeservice.NewService(projectTypeRepo)
	projectService := projectservice.NewService(projectRepo, projectTypeRepo)
	salespersonService := salespersonservice.NewService(salespersonRepo)

	authhandler.NewHandler(router, validate, authService, deps.Config.SecretJWT).RouteList()
	userhandler.NewHandler(router, validate, userService, deps.Config.SecretJWT).RouteList()
	budgethandler.NewHandler(router, validate, budgetService, deps.Config.SecretJWT).RouteList()
	budgetfollowuphandler.NewHandler(router, validate, budgetFollowUpService, deps.Config.SecretJWT).RouteList()
	budgetstatushandler.NewHandler(router, validate, budgetStatusService, deps.Config.SecretJWT).RouteList()
	budgetstatushistoryhandler.NewHandler(router, validate, budgetStatusHistoryService, deps.Config.SecretJWT).RouteList()
	installerhandler.NewHandler(router, validate, installerService, deps.Config.SecretJWT).RouteList()
	contacthandler.NewHandler(router, validate, contactService, deps.Config.SecretJWT).RouteList()
	lossreasonhandler.NewHandler(router, validate, lossReasonService, deps.Config.SecretJWT).RouteList()
	priorityhandler.NewHandler(router, validate, priorityService, deps.Config.SecretJWT).RouteList()
	projecttypehandler.NewHandler(router, validate, projectTypeService, deps.Config.SecretJWT).RouteList()
	projecthandler.NewHandler(router, validate, projectService, deps.Config.SecretJWT).RouteList()
	salespersonhandler.NewHandler(router, validate, salespersonService, deps.Config.SecretJWT).RouteList()

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
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
