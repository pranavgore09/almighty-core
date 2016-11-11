package design

import (
	d "github.com/goadesign/goa/design"
	a "github.com/goadesign/goa/design/apidsl"
)

// CreateWorkItemPayload defines the structure of work item payload
var CreateWorkItemPayload = a.Type("CreateWorkItemPayload", func() {
	a.Attribute("type", d.String, "The type of the newly created work item", func() {
		a.Example("system.userstory")
		a.MinLength(1)
		a.Pattern("^[\\p{L}.]+$")
	})
	a.Attribute("fields", a.HashOf(d.String, d.Any), "The field values, must conform to the type", func() {
		a.Example(map[string]interface{}{"system.creator": "user-ref", "system.state": "new", "system.title": "Example story"})
		a.MinLength(1)
	})
	a.Required("type", "fields")
})

// UpdateWorkItemPayload has been added because the design.WorkItem could
// not be used since it mandated the presence of the ID in the payload
// which ideally should be optional. The ID should be passed on to REST URL.
var UpdateWorkItemPayload = a.Type("UpdateWorkItemPayload", func() {
	a.Attribute("type", d.String, "The type of the newly created work item", func() {
		a.Example("system.userstory")
		a.MinLength(1)
		a.Pattern("^[\\p{L}.]+$")
	})
	a.Attribute("fields", a.HashOf(d.String, d.Any), "The field values, must conform to the type", func() {
		a.Example(map[string]interface{}{"system.creator": "user-ref", "system.state": "new", "system.title": "Example story"})
		a.MinLength(1)
	})
	a.Attribute("version", d.Integer, "Version for optimistic concurrency control", func() {
		a.Example(0)
	})
	a.Required("type", "fields", "version")
})

// UpdateWorkItemJSONAPIPayload defines top level structure from jsonapi specs
// visit : http://jsonapi.org/format/#document-top-level
var UpdateWorkItemJSONAPIPayload = a.Type("UpdateWorkItemJSONAPIPayload", func() {
	a.Attribute("data", WorkItemDataForUpdate)
	a.Required("data")
})

// WorkItemDataForUpdate defines how an update payload will look like
var WorkItemDataForUpdate = a.Type("WorkItemDataForUpdate", func() {
	a.Attribute("type", d.String, func() {
		a.Enum("workitems")
	})
	a.Attribute("id", d.String, "ID of the work item which is being updated", func() {
		a.Example("42")
	})
	a.Attribute("attributes", WorkItemAttributes)
	a.Required("type", "id", "attributes")
	a.Attribute("relationships", WorkItemRelationships)
})

// WorkItemAttributes defines attributes of WI
// visit : http://jsonapi.org/format/#document-resource-objects
var WorkItemAttributes = a.Type("WorkItemAttributes", func() {
	a.Attribute("version", d.Integer, "version for optimistic concurrency control", func() {
		a.Example(5)
	})
	a.Attribute("type", d.String, "The type of the newly created work item", func() {
		a.Example("system.userstory")
		a.MinLength(1)
		a.Pattern("^[\\p{L}.]+$")
	})
	a.Attribute("fields", a.HashOf(d.String, d.Any), "The field values, must conform to the type", func() {
		a.Example(map[string]interface{}{"system.creator": "user-ref", "system.state": "new", "system.title": "Example story"})
		a.MinLength(1)
	})
	a.Required("version", "type", "fields")
})

// WorkItemRelationships defines only `assignee` as of now. To be updated
var WorkItemRelationships = a.Type("WorkItemRelationships", func() {
	a.Attribute("assignee", RelationAssignee, "This deinfes assignees of the WI")
})

// RelationAssignee is a top level structure for assignee relationship
var RelationAssignee = a.Type("RelationAssignee", func() {
	a.Attribute("data", AssigneeData)
	a.Required("data")
})

// AssigneeData defines what is needed inside Assignee Relationship
var AssigneeData = a.Type("AssigneeData", func() {
	a.Attribute("type", d.String, func() {
		a.Enum("identities")
	})
	a.Attribute("id", d.String, "UUID of the identity", func() {
		a.Example("6c5610be-30b2-4880-9fec-81e4f8e4fd76")
	})
	a.Required("type", "id")
})

// CreateWorkItemTypePayload explains how input payload should look like
var CreateWorkItemTypePayload = a.Type("CreateWorkItemTypePayload", func() {
	a.Attribute("name", d.String, "Readable name of the type like Task, Issue, Bug, Epic etc.", func() {
		a.Example("Epic")
		a.Pattern("^[\\p{L}.]+$")
		a.MinLength(1)
	})
	a.Attribute("fields", a.HashOf(d.String, fieldDefinition), "Type fields those must be followed by respective Work Items.", func() {
		a.Example(map[string]interface{}{
			"system.administrator": map[string]interface{}{
				"Type": map[string]interface{}{
					"Kind": "string",
				},
				"Required": true,
			},
		})
		a.MinLength(1)
	})
	a.Attribute("extendedTypeName", d.String, "If newly created type extends any existing type", func() {
		a.Example("(optional field)Parent type name")
		a.MinLength(1)
		a.Pattern("^[\\p{L}.]+$")
	})
	a.Required("name", "fields")
})

// CreateTrackerAlternatePayload defines the structure of tracker payload for create
var CreateTrackerAlternatePayload = a.Type("CreateTrackerAlternatePayload", func() {
	a.Attribute("url", d.String, "URL of the tracker", func() {
		a.Example("https://api.github.com/")
		a.MinLength(1)
	})
	a.Attribute("type", d.String, "Type of the tracker", func() {
		a.Example("github")
		a.Pattern("^[\\p{L}]+$")
		a.MinLength(1)
	})
	a.Required("url", "type")
})

// UpdateTrackerAlternatePayload defines the structure of tracker payload for update
var UpdateTrackerAlternatePayload = a.Type("UpdateTrackerAlternatePayload", func() {
	a.Attribute("url", d.String, "URL of the tracker", func() {
		a.Example("https://api.github.com/")
		a.MinLength(1)
	})
	a.Attribute("type", d.String, "Type of the tracker", func() {
		a.Example("github")
		a.MinLength(1)
		a.Pattern("^[\\p{L}]+$")
	})
	a.Required("url", "type")
})

// CreateTrackerQueryAlternatePayload defines the structure of tracker query payload for create
var CreateTrackerQueryAlternatePayload = a.Type("CreateTrackerQueryAlternatePayload", func() {
	a.Attribute("query", d.String, "Search query", func() {
		a.Example("is:open is:issue user:almighty")
		a.MinLength(1)
	})
	a.Attribute("schedule", d.String, "Schedule for fetch and import", func() {
		a.Example("0 0/15 * * * *")
		a.Pattern("^[\\d]+|[\\d]+[\\/][\\d]+|\\*|\\-|\\?\\s{0,6}$")
		a.MinLength(1)
	})
	a.Attribute("trackerID", d.String, "Tracker ID", func() {
		a.Example("1")
		a.MinLength(1)
		a.Pattern("^[\\p{N}]+$")
	})
	a.Required("query", "schedule", "trackerID")
})

// UpdateTrackerQueryAlternatePayload defines the structure of tracker query payload for update
var UpdateTrackerQueryAlternatePayload = a.Type("UpdateTrackerQueryAlternatePayload", func() {
	a.Attribute("query", d.String, "Search query", func() {
		a.Example("is:open is:issue user:almighty")
		a.MinLength(1)
	})
	a.Attribute("schedule", d.String, "Schedule for fetch and import", func() {
		a.Example("0 0/15 * * * *")
		a.Pattern("^[\\d]+|[\\d]+[\\/][\\d]+|\\*|\\-|\\?\\s{0,6}$")
		a.MinLength(1)
	})
	a.Attribute("trackerID", d.String, "Tracker ID", func() {
		a.Example("1")
		a.MinLength(1)
		a.Pattern("[\\p{N}]+")
	})
	a.Required("query", "schedule", "trackerID")
})

// IdentityDataAttributes represents an identified user object attributes
var IdentityDataAttributes = a.Type("IdentityDataAttributes", func() {
	a.Attribute("fullName", d.String, "The users full name")
	a.Attribute("imageURL", d.String, "The avatar image for the user")
})

// IdentityData represents an identified user object
var IdentityData = a.Type("IdentityData", func() {
	a.Attribute("id", d.String, "unique id for the user identity")
	a.Attribute("type", d.String, "type of the user identity")
	a.Attribute("attributes", IdentityDataAttributes, "Attributes of the user identity")
})
