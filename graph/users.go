package graph

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/jinzhu/copier"
	"github.com/tuken/nix/db"
	"github.com/tuken/nix/graph/model"
	"github.com/tuken/nix/pipeline"
)

func createUser(ctx context.Context, usr *db.User, input model.CreateUserInput) (*model.User, error) {

	d := pipeline.MustDB(ctx)
	l := pipeline.MustLogger(ctx)

	switch usr.Role.Name {

	case "admin":

	case "owner":

	case "worker":
		return nil, fmt.Errorf("forbidden operation by worker")

	default:
		return nil, fmt.Errorf("unsupported role: %s", usr.Role.Name)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hash password: %w", err)
	}

	newUser := &db.User{
		OrgID:      uint(usr.OrgID),
		RoleID:     uint(input.RoleID),
		Email:      input.Email,
		Password:   string(hash),
		FirstName:  input.FirstName,
		LastName:   input.LastName,
		PostalCode: input.PostalCode,
		Address:    input.Address,
		Gender:     input.Gender,
		Birthday:   input.Birthday,
		Note:       input.Note,
	}

	if input.ParentID != nil {
		newUser.ParentID.Scan(*input.ParentID)
	}

	if input.FarmName != nil {
		newUser.FarmName.Scan(*input.FarmName)
	}

	if err := d.Create(newUser).Error; err != nil {
		l.Errorw("Error insert users", "data", newUser, "error", err)
		return nil, fmt.Errorf("error insert users: %w", err)
	}

	dbUser := db.User{}
	if err := d.Preload("Org").Preload("Parent").Preload("Role").Preload("Fields").First(&dbUser, newUser.ID).Error; err != nil {
		return nil, fmt.Errorf("error fetch users: %w", err)
	}

	user := model.User{}
	copier.Copy(&user, &dbUser)

	return &user, nil
}
