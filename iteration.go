package main

import (
	"fmt"

	"github.com/almighty/almighty-core/app"
	"github.com/almighty/almighty-core/application"
	"github.com/almighty/almighty-core/errors"
	"github.com/almighty/almighty-core/iteration"
	"github.com/almighty/almighty-core/jsonapi"
	"github.com/almighty/almighty-core/login"
	"github.com/goadesign/goa"
	uuid "github.com/satori/go.uuid"
)

// IterationController implements the iteration resource.
type IterationController struct {
	*goa.Controller
	db application.DB
}

// NewIterationController creates a iteration controller.
func NewIterationController(service *goa.Service, db application.DB) *IterationController {
	return &IterationController{Controller: service.NewController("IterationController"), db: db}
}

// CreateChild runs the create-child action.
func (c *IterationController) CreateChild(ctx *app.CreateChildIterationContext) error {
	_, err := login.ContextIdentity(ctx)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, goa.ErrUnauthorized(err.Error()))
	}
	parentID, err := uuid.FromString(ctx.IterationID)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, goa.ErrNotFound(err.Error()))
	}

	return application.Transactional(c.db, func(appl application.Application) error {

		parent, err := appl.Iterations().Load(ctx, parentID)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, goa.ErrNotFound(err.Error()))
		}

		reqIter := ctx.Payload.Data
		if reqIter.Attributes.Name == nil {
			return jsonapi.JSONErrorResponse(ctx, errors.NewBadParameterError("data.attributes.name", nil).Expected("not nil"))
		}

		newItr := iteration.Iteration{
			SpaceID:  parent.SpaceID,
			ParentID: parentID,
			Name:     *reqIter.Attributes.Name,
			StartAt:  reqIter.Attributes.StartAt,
			EndAt:    reqIter.Attributes.EndAt,
		}

		err = appl.Iterations().Create(ctx, &newItr)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}

		res := &app.IterationSingle{
			Data: ConvertIteration(ctx.RequestData, &newItr),
		}
		ctx.ResponseData.Header().Set("Location", AbsoluteURL(ctx.RequestData, app.IterationHref(res.Data.ID)))
		return ctx.Created(res)
	})
}

// Show runs the show action.
func (c *IterationController) Show(ctx *app.ShowIterationContext) error {
	id, err := uuid.FromString(ctx.IterationID)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, goa.ErrNotFound(err.Error()))
	}

	return application.Transactional(c.db, func(appl application.Application) error {
		c, err := appl.Iterations().Load(ctx, id)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}

		res := &app.IterationSingle{}
		res.Data = ConvertIteration(
			ctx.RequestData,
			c)

		return ctx.OK(res)
	})
}

// Update runs the update action.
func (c *IterationController) Update(ctx *app.UpdateIterationContext) error {
	_, err := login.ContextIdentity(ctx)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, goa.ErrUnauthorized(err.Error()))
	}
	id, err := uuid.FromString(ctx.IterationID)
	if err != nil {
		return jsonapi.JSONErrorResponse(ctx, goa.ErrNotFound(err.Error()))
	}

	return application.Transactional(c.db, func(appl application.Application) error {
		itr, err := appl.Iterations().Load(ctx.Context, id)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}
		if ctx.Payload.Data.Attributes.Name != nil {
			itr.Name = *ctx.Payload.Data.Attributes.Name
		}
		if ctx.Payload.Data.Attributes.StartAt != nil {
			itr.StartAt = ctx.Payload.Data.Attributes.StartAt
		}
		if ctx.Payload.Data.Attributes.EndAt != nil {
			itr.EndAt = ctx.Payload.Data.Attributes.EndAt
		}
		if ctx.Payload.Data.Attributes.Description != nil {
			itr.Description = ctx.Payload.Data.Attributes.Description
		}
		if ctx.Payload.Data.Attributes.State != nil {
			if *ctx.Payload.Data.Attributes.State == iteration.IterationStateStart {
				res, err := appl.Iterations().CanStartIteration(ctx, itr)
				if res == false && err != nil {
					return jsonapi.JSONErrorResponse(ctx, err)
				}
			}
			itr.State = *ctx.Payload.Data.Attributes.State
		}
		itr, err = appl.Iterations().Save(ctx.Context, *itr)
		if err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}

		response := app.IterationSingle{
			Data: ConvertIteration(ctx.RequestData, itr),
		}

		return ctx.OK(&response)
	})
}

// IterationConvertFunc is a open ended function to add additional links/data/relations to a Iteration during
// convertion from internal to API
type IterationConvertFunc func(*goa.RequestData, *iteration.Iteration, *app.Iteration)

// ConvertIterations converts between internal and external REST representation
func ConvertIterations(request *goa.RequestData, Iterations []*iteration.Iteration, additional ...IterationConvertFunc) []*app.Iteration {
	var is = []*app.Iteration{}
	for _, i := range Iterations {
		is = append(is, ConvertIteration(request, i, additional...))
	}
	return is
}

// ConvertIteration converts between internal and external REST representation
func ConvertIteration(request *goa.RequestData, itr *iteration.Iteration, additional ...IterationConvertFunc) *app.Iteration {
	iterationType := iteration.APIStringTypeIteration
	spaceType := "spaces"

	spaceID := itr.SpaceID.String()

	selfURL := AbsoluteURL(request, app.IterationHref(itr.ID))
	spaceSelfURL := AbsoluteURL(request, "/api/spaces/"+spaceID)

	i := &app.Iteration{
		Type: iterationType,
		ID:   &itr.ID,
		Attributes: &app.IterationAttributes{
			Name:        &itr.Name,
			StartAt:     itr.StartAt,
			EndAt:       itr.EndAt,
			Description: itr.Description,
			State:       &itr.State,
		},
		Relationships: &app.IterationRelations{
			Space: &app.RelationGeneric{
				Data: &app.GenericData{
					Type: &spaceType,
					ID:   &spaceID,
				},
				Links: &app.GenericLinks{
					Self: &spaceSelfURL,
				},
			},
		},
		Links: &app.GenericLinks{
			Self: &selfURL,
		},
	}
	if itr.ParentID != uuid.Nil {
		parentSelfURL := AbsoluteURL(request, app.IterationHref(itr.ParentID))
		parentID := itr.ParentID.String()
		i.Relationships.Parent = &app.RelationGeneric{
			Data: &app.GenericData{
				Type: &iterationType,
				ID:   &parentID,
			},
			Links: &app.GenericLinks{
				Self: &parentSelfURL,
			},
		}
	}
	for _, add := range additional {
		add(request, itr, i)
	}
	return i
}

// ConvertIterationSimple converts a simple Iteration ID into a Generic Reletionship
func ConvertIterationSimple(request *goa.RequestData, id interface{}) *app.GenericData {
	t := "identities"
	i := fmt.Sprint(id)
	return &app.GenericData{
		Type:  &t,
		ID:    &i,
		Links: createIterationLinks(request, id),
	}
}

func createIterationLinks(request *goa.RequestData, id interface{}) *app.GenericLinks {
	selfURL := AbsoluteURL(request, app.IterationHref(id))
	return &app.GenericLinks{
		Self: &selfURL,
	}
}
