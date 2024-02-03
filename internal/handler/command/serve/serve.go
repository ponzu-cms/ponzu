package serve

import (
	"fmt"
	"github.com/fanky5g/ponzu/internal/application"
	"github.com/fanky5g/ponzu/internal/application/analytics"
	"github.com/fanky5g/ponzu/internal/application/auth"
	"github.com/fanky5g/ponzu/internal/application/config"
	"github.com/fanky5g/ponzu/internal/application/content"
	"github.com/fanky5g/ponzu/internal/application/search"
	"github.com/fanky5g/ponzu/internal/application/storage"
	"github.com/fanky5g/ponzu/internal/application/tls"
	"github.com/fanky5g/ponzu/internal/application/users"
	"github.com/fanky5g/ponzu/internal/handler/controllers"
	"github.com/fanky5g/ponzu/internal/handler/controllers/middleware"
	analyticsRepositoryFactory "github.com/fanky5g/ponzu/internal/infrastructure/db/repository/analytics"
	configRepositoryFactory "github.com/fanky5g/ponzu/internal/infrastructure/db/repository/config"
	contentRepositoryFactory "github.com/fanky5g/ponzu/internal/infrastructure/db/repository/content"
	credentialRepositoryFactory "github.com/fanky5g/ponzu/internal/infrastructure/db/repository/credential"
	recoveryKeyRepositoryFactory "github.com/fanky5g/ponzu/internal/infrastructure/db/repository/recovery-key"
	uploadRepositoryFactory "github.com/fanky5g/ponzu/internal/infrastructure/db/repository/uploads"
	usersRepositoryFactory "github.com/fanky5g/ponzu/internal/infrastructure/db/repository/users"
	bleveSearchClient "github.com/fanky5g/ponzu/internal/infrastructure/search"
	localStorageClient "github.com/fanky5g/ponzu/internal/infrastructure/storage"
	"github.com/spf13/cobra"
	"log"
	"net/http"
)

var (
	bind      string
	httpsPort int
	port      int
	https     bool
	devHttps  bool
)

func defineFlags() {
	serveCmd.Flags().StringVar(&bind, "bind", "localhost", "address for ponzu to bind the HTTP(S) server")
	serveCmd.Flags().IntVar(&httpsPort, "https-port", 443, "port for ponzu to bind its HTTPS listener")
	serveCmd.Flags().IntVar(&port, "port", 8080, "port for ponzu to bind its HTTP listener")
	serveCmd.Flags().BoolVar(&https, "https", false, "enable automatic TLS/SSL certificate management")
	serveCmd.Flags().BoolVar(&devHttps, "dev-https", false, "[dev environment] enable automatic TLS/SSL certificate management")
}

var serveCmd = &cobra.Command{
	Use:     "serve [flags]",
	Aliases: []string{"s"},
	Short:   "run the server",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := GetDatabase()
		defer Close(store)

		// Initialize repositories
		configRepository, err := configRepositoryFactory.New(store)
		if err != nil {
			log.Fatalf("Failed to initialize config repository %v", err)
		}

		analyticsRepository, err := analyticsRepositoryFactory.New(store)
		if err != nil {
			log.Fatalf("Failed to initialize analytics repository: %v", err)
		}

		userRepository, err := usersRepositoryFactory.New(store)
		if err != nil {
			log.Fatalf("Failed to initialize user repository: %v", err)
		}

		contentRepository, err := contentRepositoryFactory.New(store)
		if err != nil {
			log.Fatalf("Failed to initialize content repository: %v", err)
		}

		credentialRepository, err := credentialRepositoryFactory.New(store)
		if err != nil {
			log.Fatalf("Failed to initialize credential repository: %v", err)
		}

		recoveryKeyRepository, err := recoveryKeyRepositoryFactory.New(store)
		if err != nil {
			log.Fatalf("Failed to initialize recovery key repository: %v", err)
		}

		uploadRepository, err := uploadRepositoryFactory.New(store)
		if err != nil {
			log.Fatalf("Failed to initialize upload repository: %v", err)
		}
		// End initialize repositories

		// Initialize clients
		storageClient, err := localStorageClient.New()
		if err != nil {
			log.Fatalf("Failed to initialize storage client: %v", err)
		}

		contentSearchClient, err := bleveSearchClient.New(contentRepository)
		if err != nil {
			log.Fatalf("Failed to initialize search client: %v\n", err)
		}

		uploadsSearchClient, err := bleveSearchClient.New(uploadRepository)
		if err != nil {
			log.Fatalf("Failed to initialize upload search client")
		}
		// End initialize clients

		// Initialize services
		services := make(application.Services)

		tlsService, err := tls.New(configRepository)
		if err != nil {
			log.Fatalf("Failed to initialize tls services %v", err)
		}
		services[tls.ServiceToken] = tlsService

		userService, err := users.New(userRepository)
		if err != nil {
			log.Fatalf("Failed to initialize user services: %v", err)
		}
		services[users.ServiceToken] = userService

		authService, err := auth.New(configRepository, userRepository, credentialRepository, recoveryKeyRepository)
		if err != nil {
			log.Fatalf("Failed to initialize auth services: %v", err)
		}
		services[auth.ServiceToken] = authService

		analyticsService, err := analytics.New(analyticsRepository)
		if err != nil {
			log.Fatalf("Failed to initialize analytics services: %v", err)
		}
		services[analytics.ServiceToken] = analyticsService

		configService, err := config.New(configRepository)
		if err != nil {
			log.Fatalf("Failed to initialize config services: %v", err)
		}
		services[config.ServiceToken] = configService

		contentService, err := content.New(contentRepository, configRepository, contentSearchClient)
		if err != nil {
			log.Fatalf("Failed to initialize content services: %v", err)
		}
		services[content.ServiceToken] = contentService

		contentSearchService, err := search.New(contentSearchClient)
		if err != nil {
			log.Fatalf("Failed to initialize search service: %v", err)
		}
		services[search.ContentSearchService] = contentSearchService

		uploadSearchService, err := search.New(uploadsSearchClient)
		if err != nil {
			log.Fatalf("Failed to initialize search service: %v", err)
		}
		services[search.UploadSearchService] = uploadSearchService

		storageService, err := storage.New(uploadRepository, configRepository, uploadsSearchClient, storageClient)
		if err != nil {
			log.Fatalf("Failed to initialize storage services: %v", err)
		}
		services[storage.ServiceToken] = storageService
		// End initialize services

		// Initialize Middlewares
		middlewares := make(middleware.Middlewares)
		middlewares[middleware.CacheControlMiddleware] = middleware.NewCacheControlMiddleware(configRepository)
		middlewares[middleware.AnalyticsRecorderMiddleware] = middleware.NewAnalyticsRecorderMiddleware(analyticsService)
		middlewares[middleware.AuthMiddleware] = middleware.NewAuthMiddleware(authService)
		// End initialize middlewares

		// Initialize Handlers
		controllers.RegisterRoutes(services, middlewares)
		// End Initialize Handlers

		// Initialize Application
		// save the https port the system is listening on
		err = configRepository.PutConfig("https_port", fmt.Sprintf("%d", httpsPort))
		if err != nil {
			log.Fatalln("System failed to save config. Please try to run again.", err)
		}

		// save the https port the system is listening on so internal system can make
		// HTTP api calls while in dev or production w/o adding more cli flags
		err = configRepository.PutConfig("http_port", fmt.Sprintf("%d", port))
		if err != nil {
			log.Fatalln("System failed to save config. Please try to run again.", err)
		}

		// save the bound address the system is listening on so internal system can make
		// HTTP api calls while in dev or production w/o adding more cli flags
		if bind == "" {
			bind = "localhost"
		}
		err = configRepository.PutConfig("bind_addr", bind)
		if err != nil {
			log.Fatalln("System failed to save config. Please try to run again.", err)
		}

		// cannot run production HTTPS and development HTTPS together
		if devHttps {
			fmt.Println("Enabling self-signed HTTPS... [DEV]")

			go tlsService.EnableDev()
			fmt.Println("Server listening on https://localhost:10443 for requests... [DEV]")
			fmt.Println("----")
			fmt.Println("If your browser rejects HTTPS requests, try allowing insecure connections on localhost.")
			fmt.Println("on Chrome, visit chrome://flags/#allow-insecure-localhost")

		} else if https {
			fmt.Println("Enabling HTTPS...")

			go tlsService.Enable()
			fmt.Printf("Server listening on :%s for HTTPS requests...\n", configRepository.Cache().GetByKey("https_port").(string))
		}

		fmt.Printf("Server listening at %s:%d for HTTP requests...\n", bind, port)
		fmt.Println("\nVisit '/' to get started.")
		log.Fatalln(http.ListenAndServe(fmt.Sprintf("%s:%d", bind, port), nil))
		return nil
	},
}

func RegisterCommandRecursive(parent *cobra.Command) {
	defineFlags()
	parent.AddCommand(serveCmd)
}
