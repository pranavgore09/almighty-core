package main

import (
	"fmt"

	"github.com/almighty/almighty-core/app"
	"github.com/almighty/almighty-core/application"
	"github.com/goadesign/goa"
)

// IdentityController implements the identity resource.
type IdentityController struct {
	*goa.Controller
	db application.DB
}

// NewIdentityController creates a identity controller.
func NewIdentityController(service *goa.Service, db application.DB) *IdentityController {
	return &IdentityController{Controller: service.NewController("IdentityController"), db: db}
}

// List runs the list action.
func (c *IdentityController) List(ctx *app.ListIdentityContext) error {
	return application.Transactional(c.db, func(appl application.Application) error {
		result, err := appl.Identities().List(ctx.Context)
		if err != nil {
			return goa.ErrInternal(fmt.Sprintf("Error listing identities: %s", err.Error()))
		}
		return ctx.OK(result)
	})
}
