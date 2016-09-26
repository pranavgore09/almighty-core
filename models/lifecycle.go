package models

import (
	"time"
)

// The Lifecycle struct contains all the items from gorm.Model except the ID field,
// hence we can embed the Lifecycle struct into Models that needs soft delete and alike.
type Lifecycle struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// LifecycleDeleteAsIndex is based on Lifecycle.
// But it makes DeletedAt as a index with name "combine_with_delete_index"
// Any attriute with this index name will be unique_together key with DeletedAt
type LifecycleDeleteAsIndex struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"unique_index:combine_with_delete_index"`
}

// Ensure Lifecyle implements the Equaler interface
var _ Equaler = Lifecycle{}
var _ Equaler = (*Lifecycle)(nil)

// Equal returns true if two Lifecycle objects are equal; otherwise false is returned.
func (self Lifecycle) Equal(u Equaler) bool {
	other, ok := u.(Lifecycle)
	if !ok {
		return false
	}
	if !self.CreatedAt.Equal(other.CreatedAt) {
		return false
	}
	if !self.UpdatedAt.Equal(other.UpdatedAt) {
		return false
	}
	// DeletedAt can be nil so we need to do a special check here.
	if self.DeletedAt == nil && other.DeletedAt == nil {
		return true
	}
	if self.DeletedAt != nil && other.DeletedAt != nil {
		return self.DeletedAt.Equal(*other.DeletedAt)
	}
	return false
}
