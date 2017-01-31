package workitem_test

import (
	"fmt"
	"testing"

	"github.com/almighty/almighty-core/errors"
	"github.com/almighty/almighty-core/gormsupport"
	"github.com/almighty/almighty-core/iteration"
	"github.com/almighty/almighty-core/rendering"
	"github.com/almighty/almighty-core/space"
	"github.com/almighty/almighty-core/workitem"
	errs "github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"
)

type workItemRepoBlackBoxTest struct {
	gormsupport.DBTestSuite
	repo workitem.WorkItemRepository
}

func TestRunWorkTypeRepoBlackBoxTest(t *testing.T) {
	suite.Run(t, &workItemRepoBlackBoxTest{DBTestSuite: gormsupport.NewDBTestSuite("../config.yaml")})
}

func (s *workItemRepoBlackBoxTest) SetupTest() {
	s.repo = workitem.NewWorkItemRepository(s.DB)
}

func (s *workItemRepoBlackBoxTest) TestFailDeleteZeroID() {
	defer gormsupport.DeleteCreatedEntities(s.DB)()

	// Create at least 1 item to avoid RowsEffectedCheck
	_, err := s.repo.Create(
		context.Background(), workitem.SystemBug,
		map[string]interface{}{
			workitem.SystemTitle: "Title",
			workitem.SystemState: workitem.SystemStateNew,
		}, "xx")
	require.Nil(s.T(), err, "Could not create work item")

	err = s.repo.Delete(context.Background(), "0")
	require.IsType(s.T(), errors.NotFoundError{}, errs.Cause(err))
}

func (s *workItemRepoBlackBoxTest) TestFailSaveZeroID() {
	defer gormsupport.DeleteCreatedEntities(s.DB)()

	// Create at least 1 item to avoid RowsEffectedCheck
	wi, err := s.repo.Create(
		context.Background(), workitem.SystemBug,
		map[string]interface{}{
			workitem.SystemTitle: "Title",
			workitem.SystemState: workitem.SystemStateNew,
		}, "xx")
	require.Nil(s.T(), err, "Could not create workitem")
	wi.ID = "0"
	_, err = s.repo.Save(context.Background(), *wi)
	require.IsType(s.T(), errors.NotFoundError{}, errs.Cause(err))
}

func (s *workItemRepoBlackBoxTest) TestFaiLoadZeroID() {
	defer gormsupport.DeleteCreatedEntities(s.DB)()

	// Create at least 1 item to avoid RowsEffectedCheck
	_, err := s.repo.Create(
		context.Background(), workitem.SystemBug,
		map[string]interface{}{
			workitem.SystemTitle: "Title",
			workitem.SystemState: workitem.SystemStateNew,
		}, "xx")
	require.Nil(s.T(), err, "Could not create workitem")

	_, err = s.repo.Load(context.Background(), "0")
	require.IsType(s.T(), errors.NotFoundError{}, errs.Cause(err))
}

func (s *workItemRepoBlackBoxTest) TestSaveAssignees() {
	defer gormsupport.DeleteCreatedEntities(s.DB)()

	wi, err := s.repo.Create(
		context.Background(), workitem.SystemBug,
		map[string]interface{}{
			workitem.SystemTitle:     "Title",
			workitem.SystemState:     workitem.SystemStateNew,
			workitem.SystemAssignees: []string{"A", "B"},
		}, "xx")
	require.Nil(s.T(), err, "Could not create workitem")

	wi, err = s.repo.Load(context.Background(), wi.ID)

	assert.Equal(s.T(), "A", wi.Fields[workitem.SystemAssignees].([]interface{})[0])
}

func (s *workItemRepoBlackBoxTest) TestSaveForUnchangedCreatedDate() {
	defer gormsupport.DeleteCreatedEntities(s.DB)()

	wi, err := s.repo.Create(
		context.Background(), workitem.SystemBug,
		map[string]interface{}{
			workitem.SystemTitle: "Title",
			workitem.SystemState: workitem.SystemStateNew,
		}, "xx")
	require.Nil(s.T(), err, "Could not create workitem")

	wi, err = s.repo.Load(context.Background(), wi.ID)

	wiNew, err := s.repo.Save(context.Background(), *wi)

	assert.Equal(s.T(), wi.Fields[workitem.SystemCreatedAt], wiNew.Fields[workitem.SystemCreatedAt])
}

func (s *workItemRepoBlackBoxTest) TestCreateWorkItemWithDescriptionNoMarkup() {
	defer gormsupport.DeleteCreatedEntities(s.DB)()

	wi, err := s.repo.Create(
		context.Background(), workitem.SystemBug,
		map[string]interface{}{
			workitem.SystemTitle:       "Title",
			workitem.SystemDescription: workitem.NewMarkupContentFromLegacy("Description"),
			workitem.SystemState:       workitem.SystemStateNew,
		}, "xx")
	require.Nil(s.T(), err, "Could not create workitem")

	wi, err = s.repo.Load(context.Background(), wi.ID)
	// app.WorkItem does not contain the markup associated with the description (yet)
	assert.Equal(s.T(), workitem.NewMarkupContentFromLegacy("Description"), wi.Fields[workitem.SystemDescription])
}

func (s *workItemRepoBlackBoxTest) TestCreateWorkItemWithDescriptionMarkup() {
	defer gormsupport.DeleteCreatedEntities(s.DB)()
	wi, err := s.repo.Create(
		context.Background(), workitem.SystemBug,
		map[string]interface{}{
			workitem.SystemTitle:       "Title",
			workitem.SystemDescription: workitem.NewMarkupContent("Description", rendering.SystemMarkupMarkdown),
			workitem.SystemState:       workitem.SystemStateNew,
		}, "xx")
	require.Nil(s.T(), err, "Could not create workitem")
	wi, err = s.repo.Load(context.Background(), wi.ID)
	// app.WorkItem does not contain the markup associated with the description (yet)
	assert.Equal(s.T(), workitem.NewMarkupContent("Description", rendering.SystemMarkupMarkdown), wi.Fields[workitem.SystemDescription])
}

// TestTypeChangeIsNotProhibitedOnDBLayer tests that you can change the type of
// a work item. NOTE: This functionality only works on the DB layer and is not
// exposed to REST.
func (s *workItemRepoBlackBoxTest) TestTypeChangeIsNotProhibitedOnDBLayer() {
	defer gormsupport.DeleteCreatedEntities(s.DB)()

	// Create at least 1 item to avoid RowsAffectedCheck
	wi, err := s.repo.Create(
		context.Background(), "bug",
		map[string]interface{}{
			workitem.SystemTitle: "Title",
			workitem.SystemState: workitem.SystemStateNew,
		}, "xx")

	require.Nil(s.T(), err)

	wi.Type = "feature"

	newWi, err := s.repo.Save(context.Background(), *wi)
	require.Nil(s.T(), err)
	require.Equal(s.T(), "feature", newWi.Type)
}

// TestGetCountsPerIteration makes sure that the query being executed is correctly returning
// the counts of work items
func (s *workItemRepoBlackBoxTest) TestGetCountsPerIteration() {
	defer gormsupport.DeleteCreatedEntities(s.DB)()
	// create seed data
	spaceRepo := space.NewRepository(s.DB)
	spaceInstance := space.Space{
		Name: "Testing space",
	}
	spaceRepo.Create(context.Background(), &spaceInstance)
	fmt.Println("space id = ", spaceInstance.ID)
	assert.NotEqual(s.T(), uuid.UUID{}, spaceInstance.ID)

	iterationRepo := iteration.NewIterationRepository(s.DB)
	iteration1 := iteration.Iteration{
		Name:    "Sprint 1",
		SpaceID: spaceInstance.ID,
	}
	iterationRepo.Create(context.Background(), &iteration1)
	fmt.Println("iteration1 id = ", iteration1.ID)
	assert.NotEqual(s.T(), uuid.UUID{}, iteration1.ID)

	iteration2 := iteration.Iteration{
		Name:    "Sprint 2",
		SpaceID: spaceInstance.ID,
	}
	iterationRepo.Create(context.Background(), &iteration2)
	fmt.Println("iteration2 id = ", iteration2.ID)
	assert.NotEqual(s.T(), uuid.UUID{}, iteration2.ID)

	for i := 0; i < 3; i++ {
		s.repo.Create(
			context.Background(), workitem.SystemBug,
			map[string]interface{}{
				workitem.SystemTitle:     fmt.Sprintf("New issue #%d", i),
				workitem.SystemState:     workitem.SystemStateNew,
				workitem.SystemIteration: iteration1.ID.String(),
			}, "xx")
	}
	for i := 0; i < 2; i++ {
		s.repo.Create(
			context.Background(), workitem.SystemBug,
			map[string]interface{}{
				workitem.SystemTitle:     fmt.Sprintf("Closed issue #%d", i),
				workitem.SystemState:     workitem.SystemStateClosed,
				workitem.SystemIteration: iteration1.ID.String(),
			}, "xx")
	}
	countsMap, _ := s.repo.GetCountsPerIteration(context.Background(), spaceInstance.ID)
	assert.Len(s.T(), countsMap, 1)
	require.Contains(s.T(), countsMap, iteration1.ID.String())
	assert.Equal(s.T(), 5, countsMap[iteration1.ID.String()].Total)
	assert.Equal(s.T(), 2, countsMap[iteration1.ID.String()].Closed)
}
