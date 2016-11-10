package main_test

import (
	"testing"

	. "github.com/almighty/almighty-core"
	"github.com/almighty/almighty-core/account"
	"github.com/almighty/almighty-core/app/test"
	"github.com/almighty/almighty-core/gormapplication"
	"github.com/almighty/almighty-core/resource"
	testsupport "github.com/almighty/almighty-core/test"
	almtoken "github.com/almighty/almighty-core/token"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestListIdentities(t *testing.T) {
	resource.Require(t, resource.Database)
	pub, _ := almtoken.ParsePublicKey([]byte(almtoken.RSAPublicKey))
	priv, _ := almtoken.ParsePrivateKey([]byte(almtoken.RSAPrivateKey))
	service := testsupport.ServiceAsUser("TestListIdentities-Service", almtoken.NewManager(pub, priv), account.TestIdentity)

	ctx := context.Background()
	identityRepo := account.NewIdentityRepository(DB)
	identity := account.Identity{
		FullName: "Test User",
		ImageURL: "http://images.com/123",
	}

	err := identityRepo.Create(ctx, &identity)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		DB.Unscoped().Delete(&identity)
	}()
	identityController := NewIdentityController(service, gormapplication.NewGormDB(DB))

	_, ic := test.ListIdentityOK(t, service.Context, service, identityController)
	assert.NotNil(t, ic)
}
