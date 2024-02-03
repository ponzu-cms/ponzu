package controllers

import (
	"github.com/fanky5g/ponzu/internal/application"
	"github.com/fanky5g/ponzu/internal/application/analytics"
	"github.com/fanky5g/ponzu/internal/application/auth"
	"github.com/fanky5g/ponzu/internal/application/config"
	"github.com/fanky5g/ponzu/internal/application/content"
	"github.com/fanky5g/ponzu/internal/application/search"
	"github.com/fanky5g/ponzu/internal/application/storage"
	"github.com/fanky5g/ponzu/internal/application/users"
	conf "github.com/fanky5g/ponzu/internal/config"
	"github.com/fanky5g/ponzu/internal/handler/controllers/middleware"
	localStorage "github.com/fanky5g/ponzu/internal/infrastructure/storage"
	"log"
	"net/http"
)

func RegisterRoutes(services application.Services, middlewares middleware.Middlewares) {
	Auth := middlewares.Get(middleware.AuthMiddleware)
	CacheControlMiddleware := middleware.ToHttpHandler(middlewares.Get(middleware.CacheControlMiddleware))

	analyticsService := services.Get(analytics.ServiceToken).(analytics.Service)
	configService := services.Get(config.ServiceToken).(config.Service)
	userService := services.Get(users.ServiceToken).(users.Service)
	authService := services.Get(auth.ServiceToken).(auth.Service)
	contentService := services.Get(content.ServiceToken).(content.Service)
	storageService := services.Get(storage.ServiceToken).(storage.Service)
	contentSearchService := services.Get(search.ContentSearchService).(search.Service)
	uploadSearchService := services.Get(search.UploadSearchService).(search.Service)

	http.HandleFunc("/", Auth(NewAdminHandler(analyticsService, configService)))

	http.HandleFunc("/init", NewInitHandler(configService, userService, authService))

	http.HandleFunc("/login", NewLoginHandler(configService, authService, userService))
	http.HandleFunc("/logout", LogoutHandler)

	http.HandleFunc("/recover", NewForgotPasswordHandler(configService, userService, authService))
	http.HandleFunc("/recover/key", NewRecoveryKeyHandler(configService, authService, userService))

	http.HandleFunc("/configure", Auth(NewConfigHandler(configService)))
	http.HandleFunc("/configure/users", Auth(NewConfigUsersHandler(configService, authService, userService)))
	http.HandleFunc("/configure/users/edit", Auth(NewConfigUsersEditHandler(configService, authService, userService)))
	http.HandleFunc("/configure/users/delete", Auth(NewConfigUsersDeleteHandler(configService, authService, userService)))

	http.HandleFunc("/uploads", Auth(NewUploadContentsHandler(configService, storageService)))
	http.HandleFunc("/uploads/search", Auth(NewUploadSearchHandler(configService, uploadSearchService)))

	http.HandleFunc("/contents", Auth(NewContentsHandler(configService, contentService)))
	http.HandleFunc("/contents/search", Auth(NewSearchHandler(configService, contentSearchService)))
	http.HandleFunc("/contents/export", Auth(NewExportHandler(configService, contentService)))

	http.HandleFunc("/edit", Auth(NewEditHandler(configService, contentService, storageService)))
	http.HandleFunc("/edit/delete", Auth(NewDeleteHandler(configService, contentService)))
	http.HandleFunc("/edit/approve", Auth(NewApproveContentHandler(configService, contentService)))
	http.HandleFunc("/edit/upload", Auth(NewEditUploadHandler(configService, storageService)))
	http.HandleFunc("/edit/upload/delete", Auth(NewDeleteUploadHandler(configService, storageService)))

	staticDir := conf.AdminStaticDir()
	staticFileSystem, err := localStorage.NewLocalStaticFileSystem(http.Dir(staticDir))
	if err != nil {
		log.Fatalf("Failed to create static file system: %v", err)
	}

	http.Handle("/static/", CacheControlMiddleware(
		http.StripPrefix("/static", http.FileServer(staticFileSystem)),
	))
}
