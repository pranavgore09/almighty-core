package account_test

import (
	"fmt"
	"os"
	"testing"

	"golang.org/x/net/context"

	"github.com/almighty/almighty-core/account"
	"github.com/almighty/almighty-core/resource"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	if _, c := os.LookupEnv(resource.Database); c == false {
		fmt.Printf(resource.StSkipReasonNotSet+"\n", resource.Database)
		return
	}

	dbhost := os.Getenv("ALMIGHTY_DB_HOST")
	if "" == dbhost {
		panic("The environment variable ALMIGHTY_DB_HOST is not specified or empty.")
	}
	var err error
	db, err = gorm.Open("postgres", fmt.Sprintf("host=%s user=postgres password=mysecretpassword sslmode=disable", dbhost))
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	defer db.Close()
	// Migrate the schema (only needed part)
	db.AutoMigrate(
		&account.Identity{},
		&account.User{},
	)
	ec := m.Run()
	os.Exit(ec)
}

func TestUserByEmails(t *testing.T) {
	resource.Require(t, resource.Database)
	// this test makes sure that UserByEmails eliminates deleted entries
	ctx := context.Background()
	userRepo := account.NewUserRepository(db)
	identityRepo := account.NewIdentityRepository(db)
	identity := account.Identity{
		FullName: "Test User Integration 123",
		ImageURL: "http://images.com/42",
	}
	email := "primary@test.com"
	err := identityRepo.Create(ctx, &identity)
	if err != nil {
		t.Fatal(err)
	}
	user1 := account.User{Email: email, Identity: identity}
	err = userRepo.Create(ctx, &user1)
	if err != nil {
		t.Fatal(err)
	}
	users, err := userRepo.Query(account.UserByEmails([]string{email}), account.UserWithIdentity())
	if err != nil {
		t.Fatal(err)
	}
	l := len(users)
	assert.NotEqual(t, 0, len(users))
	found := false
	for _, u := range users {
		if u.Email == email {
			found = true
			break
		}
	}
	if found == false {
		t.Errorf("Newly inserted email %v not found in DB", email)
	}
	// try to fetch user by identity
	u, err := userRepo.Query(account.UserFilterByIdentity(identity.ID, db))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, email, u[0].Email)

	// try filtering with non-available uuid
	u, err = userRepo.Query(account.UserFilterByIdentity(uuid.NewV4(), db))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 0, len(u))

	// try creating a user with same email, should fail.
	user2 := account.User{Email: email, Identity: identity}
	err = userRepo.Create(ctx, &user2)
	if err == nil {
		t.Error(err)
	}

	// delete the record created at the start of the test
	err = userRepo.Delete(ctx, user1.ID)
	if err != nil {
		t.Fatal(err)
	}
	identityRepo.Delete(ctx, identity.ID)

	// Check fetch after delete.
	users, err = userRepo.Query(account.UserByEmails([]string{email}), account.UserWithIdentity())
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, l-1, len(users))

	// try creating a user with same email, should succeed.
	user3 := account.User{Email: email, Identity: identity}
	err = userRepo.Create(ctx, &user3)
	if err != nil {
		t.Fatalf("Unable to create a user with same email %+v", user3)
	}

	// cleanup
	err = userRepo.Delete(ctx, user3.ID)
	if err != nil {
		t.Fatal(err)
	}

}
