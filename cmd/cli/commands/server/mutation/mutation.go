// graph/mutation/mutation.go
package mutation

import (
	"context"

	"gogogo/cmd/cli/commands/server/graph/model"
)

type Resolver struct{}

func (r *Resolver) CreateUser(ctx context.Context, input model.User) (*model.User, error) {
	return &model.User{
		ID:   "1",
		Name: input.Name,
	}, nil
}
