package ent_project

import (
	"github.com/99designs/gqlgen/graphql"
	"github.com/entkit/entkit-demo/ent-project/core/ent"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver is the resolver root.
type Resolver struct{ client *ent.Client }

// NewSchema creates a graphql executable schema.
func NewSchema(client *ent.Client) graphql.ExecutableSchema {
	return NewExecutableSchema(Config{
		Resolvers:  &Resolver{client},
		Directives: DirectiveRoot{
			//HasPermissions: HasPermission(),
		},
	})
}
