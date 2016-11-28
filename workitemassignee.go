package main

import (
	"fmt"
	"log"

	"github.com/almighty/almighty-core/app"
	"github.com/almighty/almighty-core/application"
	"github.com/almighty/almighty-core/models"
	"github.com/goadesign/goa"
)

// WorkitemassigneeController implements the workitemassignee resource.
type WorkitemassigneeController struct {
	*goa.Controller
	db application.DB
}

// NewWorkitemassigneeController creates a workitemassignee controller.
func NewWorkitemassigneeController(service *goa.Service, db application.DB) *WorkitemassigneeController {
	if db == nil {
		panic("db must not be nil")
	}
	return &WorkitemassigneeController{Controller: service.NewController("WorkitemassigneeController"), db: db}
}

// Update runs the update action.
func (c *WorkitemassigneeController) Update(ctx *app.UpdateWorkitemassigneeContext) error {
	return application.Transactional(c.db, func(appl application.Application) error {
		wi, err := appl.WorkItems2().UpdateAssignee(ctx, ctx.ID, ctx.Payload)
		if err != nil {
			switch err := err.(type) {
			case models.BadParameterError:
				return ctx.BadRequest(goa.ErrBadRequest(fmt.Sprintf("Error updating work item: %s", err.Error())))
			case models.NotFoundError:
				return ctx.NotFound()
			case models.VersionConflictError:
				return ctx.BadRequest(goa.ErrBadRequest(fmt.Sprintf("Error updating work item: %s", err.Error())))
			default:
				log.Printf("Error updating work items: %s", err.Error())
				return ctx.InternalServerError()
			}
		}
		fmt.Println("Updated Coool=====", wi)
		return ctx.OK([]byte{})
	})
}
