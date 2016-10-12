package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"time"

	"golang.org/x/net/context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"

	"github.com/almighty/almighty-core/account"
	"github.com/almighty/almighty-core/app"
	"github.com/almighty/almighty-core/login"
	"github.com/almighty/almighty-core/migration"
	"github.com/almighty/almighty-core/models"
	"github.com/almighty/almighty-core/remoteworkitem"
	"github.com/almighty/almighty-core/token"
	"github.com/almighty/almighty-core/transaction"
	jwtGo "github.com/dgrijalva/jwt-go"
	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
	"github.com/goadesign/goa/middleware/security/jwt"
)

var (
	// Commit current build commit set by build script
	Commit = "0"
	// BuildTime set by build script in ISO 8601 (UTC) format: YYYY-MM-DDThh:mm:ssTZD (see https://www.w3.org/TR/NOTE-datetime for details)
	BuildTime = "0"
	// StartTime in ISO 8601 (UTC) format
	StartTime = time.Now().UTC().Format("2006-01-02T15:04:05Z")
)

func main() {
	// --------------------------------------------------------------------
	// Parse flags
	// --------------------------------------------------------------------
	var configFilePath string
	var printConfig bool
	var scheduler *remoteworkitem.Scheduler
	flag.StringVar(&configFilePath, "config", "", "Path to the config file to read")
	flag.BoolVar(&printConfig, "printConfig", false, "Prints the config (including merged environment variables) and exits")
	flag.Parse()

	// Override default -config switch with environment variable only if -config switch was
	// not explicitly given via the command line.
	configSwitchIsSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "config" {
			configSwitchIsSet = true
		}
	})
	if !configSwitchIsSet {
		if envConfigPath, ok := os.LookupEnv("ALMIGHTY_CONFIG_FILE_PATH"); ok {
			configFilePath = envConfigPath
		}
	}

	var err error
	if err = configuration.Setup(configFilePath); err != nil {
		panic(fmt.Errorf("Failed to setup the configuration: %s", err.Error()))
	}

	if printConfig {
		fmt.Printf("%s\n", configuration.String())
		os.Exit(0)
	}

	printUserInfo()

	var db *gorm.DB
	for i := 1; i <= configuration.GetPostgresConnectionMaxRetries(); i++ {
		log.Printf("Opening DB connection attempt %d of %d\n", i, configuration.GetPostgresConnectionMaxRetries())
		db, err = gorm.Open("postgres",
			fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=%s",
				configuration.GetPostgresHost(),
				configuration.GetPostgresPort(),
				configuration.GetPostgresUser(),
				configuration.GetPostgresPassword(),
				configuration.GetPostgresSSLMode(),
			))
		if err != nil {
			time.Sleep(configuration.GetPostgresConnectionRetrySleep())
		} else {
			defer db.Close()
			break
		}
	}
	if err != nil {
		panic("Could not open connection to database")
	}

	// Migrate the schema
	ts := models.NewGormTransactionSupport(db)
	witRepo := models.NewWorkItemTypeRepository(ts)
	wiRepo := models.NewWorkItemRepository(ts, witRepo)

	identityRepository := account.NewIdentityRepository(db)
	userRepository := account.NewUserRepository(db)

	if err := transaction.Do(ts, func() error {
		return migration.Perform(context.Background(), ts.TX(), witRepo)
	}); err != nil {
		panic(err.Error())
	}

	// Scheduler to fetch and import remote tracker items
	scheduler = remoteworkitem.NewScheduler(db)
	defer scheduler.Stop()
	scheduler.ScheduleAllQueries()

	// Create service
	service := goa.New("alm")

	// Mount middleware
	service.Use(middleware.RequestID())
	service.Use(middleware.LogRequest(true))
	service.Use(middleware.ErrorHandler(service, true))
	service.Use(middleware.Recover())

	publicKey, err := jwtGo.ParseRSAPublicKeyFromPEM([]byte(token.RSAPublicKey))
	if err != nil {
		panic(err)
	}

	tokenManager := token.NewManager(publicKey, privateKey)
	app.UseJWTMiddleware(service, jwt.New(publicKey, nil, app.NewJWTSecurity()))

	// Mount "login" controller
	oauth := &oauth2.Config{
		ClientID:     "875da0d2113ba0a6951d",
		ClientSecret: "2fe6736e90a9283036a37059d75ac0c82f4f5288",
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
	tokenManager := token.NewManager(token.RSAPrivateKey, token.RSAPublicKey)
	loginService := login.NewGitHubOAuth(oauth, identityRepository, userRepository, tokenManager)
	loginCtrl := NewLoginController(service, loginService)
	app.MountLoginController(service, loginCtrl)

	// Mount "version" controller
	versionCtrl := NewVersionController(service)
	app.MountVersionController(service, versionCtrl)

	// Mount "workitem" controller
	workitemCtrl := NewWorkitemController(service, wiRepo, ts)
	app.MountWorkitemController(service, workitemCtrl)

	// Mount "workitemtype" controller
	workitemtypeCtrl := NewWorkitemtypeController(service, witRepo, ts)
	app.MountWorkitemtypeController(service, workitemtypeCtrl)

	ts2 := models.NewGormTransactionSupport(db)

	// Mount "tracker" controller
	repo2 := remoteworkitem.NewTrackerRepository(ts2)
	c5 := NewTrackerController(service, repo2, ts2, scheduler)
	app.MountTrackerController(service, c5)

	// Mount "trackerquery" controller
	repo3 := remoteworkitem.NewTrackerQueryRepository(ts2)
	c6 := NewTrackerqueryController(service, repo3, ts2, scheduler)
	app.MountTrackerqueryController(service, c6)

	// Mount "user" controller
	userCtrl := NewUserController(service, identityRepository)
	app.MountUserController(service, userCtrl)

	fmt.Println("Git Commit SHA: ", Commit)
	fmt.Println("UTC Build Time: ", BuildTime)
	fmt.Println("UTC Start Time: ", StartTime)
	fmt.Println("Dev mode:       ", configuration.IsPostgresDeveloperModeEnabled())

	http.Handle("/api/", service.Mux)
	http.Handle("/", http.FileServer(assetFS()))
	http.Handle("/favicon.ico", http.NotFoundHandler())

	// Start http
	if err := http.ListenAndServe(configuration.GetHTTPAddress(), nil); err != nil {
		service.LogError("startup", "err", err)
	}

}

func printUserInfo() {
	u, err := user.Current()
	if err != nil {
		log.Printf("Failed to get current user: %s", err.Error())
	} else {
		log.Printf("Running as user name \"%s\" with UID %s.\n", u.Username, u.Uid)
		/*
			g, err := user.LookupGroupId(u.Gid)
			if err != nil {
				fmt.Printf("Failed to lookup group: %", err.Error())
			} else {
				fmt.Printf("Running with group \"%s\" with GID %s.\n", g.Name, g.Gid)
			}
		*/
	}
}
