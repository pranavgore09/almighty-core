package main

import (
	"log"

	"github.com/almighty/almighty-core/app"
	"github.com/almighty/almighty-core/models"
	"github.com/almighty/almighty-core/transaction"
	"github.com/goadesign/goa"
)

// WorkitemtypeController implements the workitemtype resource.
type WorkitemtypeController struct {
	*goa.Controller
	witRepository models.WorkItemTypeRepository
	ts            transaction.Support
}

// NewWorkitemtypeController creates a workitemtype controller.
func NewWorkitemtypeController(service *goa.Service, witRepository models.WorkItemTypeRepository, ts transaction.Support) *WorkitemtypeController {
	return &WorkitemtypeController{
		Controller:    service.NewController("WorkitemtypeController"),
		witRepository: witRepository,
		ts:            ts,
	}
}

// Show runs the show action.
func (c *WorkitemtypeController) Show(ctx *app.ShowWorkitemtypeContext) error {
	return transaction.Do(c.ts, func() error {
		res, err := c.witRepository.Load(ctx.Context, ctx.Name)
		if err != nil {
			switch err.(type) {
			case models.NotFoundError:
				log.Printf("not found, id=%s", ctx.Name)
				return goa.ErrNotFound(err.Error())
			default:
				return err
			}
		}
		return ctx.OK(res)
	})
}

// Create runs the create action.
func (c *WorkitemtypeController) Create(ctx *app.CreateWorkitemtypeContext) error {
	return transaction.Do(c.ts, func() error {
		var fields = map[string]app.FieldDefinition{}

		// need to convert field value to app.FieldDefinition
		for key, value := range ctx.Payload.Fields {
			fd, _ := value.(map[string]interface{})
			req := fd["required"].(bool)
			fields[key] = app.FieldDefinition{
				Required: req,
				Type:     fd,
			}
		}
		wit, err := c.witRepository.Create(ctx.Context, ctx.Payload.ExtendedTypeID, ctx.Payload.Name, fields)

		if err != nil {
			switch err := err.(type) {
			case models.BadParameterError, models.ConversionError:
				return goa.ErrBadRequest(err.Error())
			default:
				return goa.ErrInternal(err.Error())
			}
		}
		ctx.ResponseData.Header().Set("Location", app.WorkitemtypeHref(wit.Name))
		return ctx.Created(wit)
	})
}
