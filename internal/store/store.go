package store

import (
	"context"
)

type (
	Store interface {
		RegisterName(context.Context, string, string) error
		UpdateName(context.Context, string, string) error
		UpsertName(context.Context, string, string) error
		LookupName(context.Context, string) (string, error)
		ReverseLookup(context.Context, string) (string, error)
		Close()
	}
)
