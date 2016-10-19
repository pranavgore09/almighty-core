package main

import (
	"crypto/rsa"

	"golang.org/x/net/context"

	"net/http"

	"github.com/almighty/almighty-core/token"
	"github.com/goadesign/goa"
)

// ConfigureExtractUser is a configuration for middleware,
// Accepts PublicKey and PrivateKey to create a token manager
// Using token manager, it is responsible for extracting the token from context if present
// Update the context with uuid found in token
func ConfigureExtractUser(publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey) (ExtractUser goa.Middleware) {
	manager := token.NewManager(publicKey, privateKey)
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			// for now ignore the error becasue still test for logged in user is not done.
			uuid, _ := manager.Locate(ctx)
			ctxWithUser := context.WithValue(ctx, "uuid", uuid.String())
			return h(ctxWithUser, rw, req)
		}
	}
}
