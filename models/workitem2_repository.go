package models

import (
	"log"
	"strconv"

	"golang.org/x/net/context"

	"fmt"

	"github.com/almighty/almighty-core/account"
	"github.com/almighty/almighty-core/app"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

type GormWorkItem2Repository struct {
	db  *gorm.DB
	wir *GormWorkItemTypeRepository
}

// NewWorkItem2Repository creates a wi repository based on gorm
func NewWorkItem2Repository(db *gorm.DB) *GormWorkItem2Repository {
	return &GormWorkItem2Repository{db, &GormWorkItemTypeRepository{db}}
}

// UpdateAssignee set/remove assignee for given work item
func (r *GormWorkItem2Repository) UpdateAssignee(ctx context.Context, ID string, assignee *app.WorkItemRelationAssignee) (*app.WorkItem, error) {
	id, err := strconv.ParseUint(ID, 10, 64)
	if err != nil {
		return nil, NotFoundError{entity: "work item", ID: ID}
	}
	log.Printf("looking for id %d", id)
	tx := r.db
	wi := WorkItem{}
	if tx.First(&wi, id).RecordNotFound() {
		log.Printf("not found, wi=%v", wi)
		return nil, NewNotFoundError("work item", ID)
	}

	strVersion := strconv.Itoa(wi.Version)
	if strVersion != assignee.Version {
		return nil, VersionConflictError{simpleError{"version conflict"}}
	}

	wiType, err := r.wir.LoadTypeFromDB(ctx, wi.Type)
	if err != nil {
		// ideally should not reach this, if reach it means something went wrong while CREATE WI
		return nil, NewBadParameterError("Type", wi.Type)
	}
	newWi := WorkItem{
		ID:      id,
		Type:    wi.Type, // read WIT from DB object and not from payload relationship
		Version: wi.Version + 1,
		Fields:  wi.Fields,
	}
	if assignee.Data != nil {
		fmt.Println("I will UPDATE the assingee now---", assignee.Data.ID, assignee.Data.Type)
		identityRepo := account.NewIdentityRepository(r.db)
		assigneeUUID, err := uuid.FromString(assignee.Data.ID)
		if err != nil {
			return nil, NewBadParameterError("data.relationships.assignee.data.id", assignee.Data.ID)
		}
		_, err = identityRepo.Load(ctx, assigneeUUID)
		if err != nil {
			return nil, NewBadParameterError("data.relationships.assignee.data.id", assignee.Data.ID)
		}
		wi.Fields[SystemAssignee] = assignee.Data.ID
	} else {
		fmt.Println("I will remove the assingee now---")
		wi.Fields[SystemAssignee] = nil
	}
	if err := tx.Save(&newWi).Error; err != nil {
		log.Print(err.Error())
		return nil, InternalError{simpleError{err.Error()}}
	}
	log.Printf("updated item to %v\n", newWi)
	result, err := wiType.ConvertFromModel(newWi)
	if err != nil {
		return nil, InternalError{simpleError{err.Error()}}
	}
	return result, nil
}
