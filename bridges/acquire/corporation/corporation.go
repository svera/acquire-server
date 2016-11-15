// Package corporation contains the model Corporation and attached methods which manages corporations in game
package corporation

import acquireCorporation "github.com/svera/acquire/corporation"

// Corporation holds data related to corporations
type Corporation struct {
	*acquireCorporation.Corporation
	name  string
	index int
}

// New initialises and returns a new instance of Corporation
func New(name string, index int) *Corporation {
	return &Corporation{
		acquireCorporation.New(),
		name,
		index,
	}
}

// Name returns the corporation name
func (c *Corporation) Name() string {
	return c.name
}

// Index returns the corporation position in the corporations array
func (c *Corporation) Index() int {
	return c.index
}
