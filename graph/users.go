package graph

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

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

func getUser(ctx context.Context, usr *db.User, id int) (*model.User, error) {

	d := pipeline.MustDB(ctx)
	l := pipeline.MustLogger(ctx)

	dbUser := db.User{}
	query := d.Preload("Org").Preload("Parent").Preload("Role").Preload("Fields")

	switch usr.Role.Name {

	case "admin":
		query = query.Where("users.org_id = ? AND users.role_id IN (?)", usr.OrgID, []uint{2, 3, 4})

	case "owner":
		query = query.Where("users.parent_id = ?", usr.ID)

	case "worker":
		query = query.Where("users.id = ?", usr.ID)

	default:
		return nil, fmt.Errorf("unsupported role: %s", usr.Role.Name)
	}

	if err := query.First(&dbUser, id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Errorw("No users", "id", id, "user_id", usr.ID)
			return nil, nil
		} else {
			l.Errorw("Error users", "id", id, "user_id", usr.ID, "error", err)
			return nil, err
		}
	}

	user := model.User{}
	copier.Copy(&user, &dbUser)

	return &user, nil
}

func listUsers(ctx context.Context, usr *db.User) ([]*model.User, error) {

	d := pipeline.MustDB(ctx)
	l := pipeline.MustLogger(ctx)

	dbUsers := []db.User{}
	query := d.Preload("Org").Preload("Parent").Preload("Role").Preload("Fields")

	switch usr.Role.Name {

	case "admin":
		query = query.Where("users.org_id = ? AND users.role_id IN (?)", usr.OrgID, []uint{2, 3, 4})

	case "owner":
		query = query.Where("users.parent_id = ?", usr.ID)

	case "worker":
		return nil, fmt.Errorf("forbidden operation by worker")

	default:
		return nil, fmt.Errorf("unsupported role: %s", usr.Role.Name)
	}

	if err := query.Find(&dbUsers).Error; err != nil {
		l.Errorw("Error users", "user_id", usr.ID, "role", usr.Role.Name, "error", err)
		return nil, err
	}

	users := []*model.User{}

	for _, dbu := range dbUsers {

		u := model.User{}
		copier.Copy(&u, &dbu)

		users = append(users, &u)
	}

	return users, nil
}

func findUsers(ctx context.Context, usr *db.User, contains string) ([]*model.User, error) {

	d := pipeline.MustDB(ctx)
	l := pipeline.MustLogger(ctx)

	dbUsers := []db.User{}
	query := d.Preload("Org").Preload("Parent").Preload("Role").Preload("Fields")

	switch usr.Role.Name {

	case "admin":
		query = query.Where("users.org_id = ? AND users.role_id IN (?)", usr.OrgID, []uint{2, 3, 4})

	case "owner":
		query = query.Where("users.parent_id = ?", usr.ID)

	case "worker":
		return nil, fmt.Errorf("forbidden operation by worker")

	default:
		return nil, fmt.Errorf("unsupported role: %s", usr.Role.Name)
	}

	query = query.Where("users.first_name LIKE %?% OR users.last_name LIKE %?% OR users.email LIKE %?%", contains, contains, contains)

	if err := query.Find(&dbUsers).Error; err != nil {
		l.Errorw("Error users", "user_id", usr.ID, "role", usr.Role.Name, "error", err)
		return nil, err
	}

	users := []*model.User{}

	for _, dbu := range dbUsers {

		u := model.User{}
		copier.Copy(&u, &dbu)

		users = append(users, &u)
	}

	return users, nil
}
