package main

import (
	"fmt"

	"github.com/almighty/almighty-core/account"
	"github.com/almighty/almighty-core/app"
	"github.com/almighty/almighty-core/login"
	"github.com/dgrijalva/jwt-go"
	"github.com/goadesign/goa"
	uuid "github.com/satori/go.uuid"
)

// LoginController implements the login resource.
type LoginController struct {
	*goa.Controller
	auth login.Service
}

// NewLoginController creates a login controller.
func NewLoginController(service *goa.Service, auth login.Service) *LoginController {
	return &LoginController{Controller: service.NewController("login"), auth: auth}
}

// Authorize runs the authorize action.
func (c *LoginController) Authorize(ctx *app.AuthorizeLoginContext) error {
	/*
		token := jwt.New(jwt.SigningMethodRS256)
		token.Claims.(jwt.MapClaims)["exp"] = time.Now().Add(time.Hour * 72).Unix()
		token.Claims.(jwt.MapClaims)["scopes"] = []string{"system"}

		key, err := jwt.ParseRSAPrivateKeyFromPEM(([]byte(RSAPrivateKey)))
		if err != nil {
			panic(err)
		}

		tokenStr, err := token.SignedString(key)
		if err != nil {
			panic(err)
		}
		authToken := app.AuthToken{Token: tokenStr}
		return ctx.OK(&authToken)
	*/
	return c.auth.Perform(ctx)
}

// Generate runs the authorize action.
func (c *LoginController) Generate(ctx *app.GenerateLoginContext) error {
	if !configuration.IsPostgresDeveloperModeEnabled() {
		return ctx.Unauthorized()
	}

	var scopes []account.Identity
	scopes = append(scopes, account.Identity{
		ID:       uuid.NewV4(),
		FullName: "Test Developer",
	})
	scopes = append(scopes, account.Identity{
		ID:       uuid.NewV4(),
		FullName: "Test Observer",
	})

	var tokens app.AuthTokenCollection
	for _, user := range scopes {
		tokenStr, err := c.tokenManager.Generate(user)
		if err != nil {
			fmt.Println("Failed to generate token", err)
			return ctx.Unauthorized()
		}
		tokens = append(tokens, &app.AuthToken{Token: tokenStr})
	}
	return ctx.OK(tokens)
}
