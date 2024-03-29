// routes.go

package routes

import (
	"log"
	"net/http"
	"time"

	"github.com/a11dev/go-gen-backend/internal/config"
	"github.com/a11dev/go-gen-backend/internal/handlers"
	"github.com/a11dev/go-gen-backend/internal/middleware"
	"github.com/a11dev/go-gen-backend/internal/runtimes"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/shaj13/go-guardian/auth"
	"github.com/shaj13/go-guardian/store"
)

func InitializeRoutes(router *gin.Engine, cfg *config.Config, authenticator auth.Authenticator, store store.Cache, inc chan runtimes.BackendMsg) {
	// the jwt middleware
	authMiddleware, err := middleware.JwtAuth(authenticator, cfg)

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	errInit := authMiddleware.MiddlewareInit()

	if errInit != nil {
		log.Fatal("authMiddleware.MiddlewareInit() Error:" + errInit.Error())
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{http.MethodGet, http.MethodPatch, http.MethodPost, http.MethodHead, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Content-Type", "X-XSRF-TOKEN", "Accept", "Origin", "X-Requested-With", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.POST("/login", authMiddleware.LoginHandler)
	router.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Printf("NoRoute claims: %#v\n", claims)
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	chans := router.Group("/chans")
	chans.Use(middleware.Backend1ChannelsMiddleware(inc))
	{
		chans.GET("/routine/:id", handlers.InvokeBackend)
	}

	auth := router.Group("/auth")
	// Refresh time can be longer than token timeout
	auth.GET("/refresh_token", authMiddleware.RefreshHandler)

	auth.Use(authMiddleware.MiddlewareFunc())
	{
		auth.GET("/healthcheck", handlers.HealthCheck)
		auth.GET("/task/:id", handlers.GetTask)
		auth.GET("/tasks", handlers.GetTasks)
	}
}
