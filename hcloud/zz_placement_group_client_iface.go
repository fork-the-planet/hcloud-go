// Code generated by ifacemaker; DO NOT EDIT.

package hcloud

import (
	"context"
)

// IPlacementGroupClient ...
type IPlacementGroupClient interface {
	// GetByID retrieves a PlacementGroup by its ID. If the PlacementGroup does not exist, nil is returned.
	GetByID(ctx context.Context, id int64) (*PlacementGroup, *Response, error)
	// GetByName retrieves a PlacementGroup by its name. If the PlacementGroup does not exist, nil is returned.
	GetByName(ctx context.Context, name string) (*PlacementGroup, *Response, error)
	// Get retrieves a PlacementGroup by its ID if the input can be parsed as an integer, otherwise it
	// retrieves a PlacementGroup by its name. If the PlacementGroup does not exist, nil is returned.
	Get(ctx context.Context, idOrName string) (*PlacementGroup, *Response, error)
	// List returns a list of PlacementGroups for a specific page.
	//
	// Please note that filters specified in opts are not taken into account
	// when their value corresponds to their zero value or when they are empty.
	List(ctx context.Context, opts PlacementGroupListOpts) ([]*PlacementGroup, *Response, error)
	// All returns all PlacementGroups.
	All(ctx context.Context) ([]*PlacementGroup, error)
	// AllWithOpts returns all PlacementGroups for the given options.
	AllWithOpts(ctx context.Context, opts PlacementGroupListOpts) ([]*PlacementGroup, error)
	// Create creates a new PlacementGroup.
	Create(ctx context.Context, opts PlacementGroupCreateOpts) (PlacementGroupCreateResult, *Response, error)
	// Update updates a PlacementGroup.
	Update(ctx context.Context, placementGroup *PlacementGroup, opts PlacementGroupUpdateOpts) (*PlacementGroup, *Response, error)
	// Delete deletes a PlacementGroup.
	Delete(ctx context.Context, placementGroup *PlacementGroup) (*Response, error)
}
