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

	if err := d.Transaction(func(tx *gorm.DB) error {

		if e := tx.Create(newUser).Error; e != nil {
			return fmt.Errorf("error insert users: %w", e)
		}

		if len(input.FieldIDs) > 0 {

			fus := []db.FieldUser{}

			for _, fid := range input.FieldIDs {

				fus = append(fus, db.FieldUser{
					UserID:  newUser.ID,
					FieldID: uint(fid),
				})
			}

			if e := tx.Create(&fus).Error; e != nil {
				return fmt.Errorf("error insert field_users: %w", e)
			}
		}

		return nil
	}); err != nil {
		l.Errorw("Error transaction create user", "data", newUser, "error", err)
		return nil, err
	}

	dbUser := db.User{}
	if err := dbUser.Preload(d).First(&dbUser, newUser.ID).Error; err != nil {
		return nil, fmt.Errorf("error fetch users: %w", err)
	}

	user := model.User{}
	copier.Copy(&user, &dbUser)

	return &user, nil
}

func updateUser(ctx context.Context, usr *db.User, id uint, input model.UpdateUserInput) (*model.User, error) {

	d := pipeline.MustDB(ctx)
	l := pipeline.MustLogger(ctx)

	dbUser := db.User{}
	query := d

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

	if input.Email != nil {
		dbUser.Email = *input.Email
	}

	if input.Password != nil {

		hash, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("error hash password: %w", err)
		}

		dbUser.Password = string(hash)
	}

	if input.FarmName != nil {
		dbUser.FarmName.Scan(*input.FarmName)
	}

	if input.FirstName != nil {
		dbUser.FirstName = *input.FirstName
	}

	if input.LastName != nil {
		dbUser.LastName = *input.LastName
	}

	if input.PostalCode != nil {
		dbUser.PostalCode = *input.PostalCode
	}

	if input.Address != nil {
		dbUser.Address = *input.Address
	}

	if input.Gender != nil {
		dbUser.Gender = *input.Gender
	}

	if input.Birthday != nil {
		dbUser.Birthday = *input.Birthday
	}

	if input.Note != nil {
		dbUser.Note = *input.Note
	}

	if err := d.Save(&dbUser).Error; err != nil {
		l.Errorw("Error update users", "id", id, "data", dbUser, "error", err)
		return nil, err
	}

	if err := dbUser.Preload(d).First(&dbUser, id).Error; err != nil {
		return nil, err
	}

	user := model.User{}
	copier.Copy(&user, &dbUser)

	return &user, nil
}

func deleteUser(ctx context.Context, usr *db.User, id uint) (*model.User, error) {

	d := pipeline.MustDB(ctx)
	l := pipeline.MustLogger(ctx)

	dbUser := db.User{}
	query := d

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

	if err := query.Delete(&dbUser, id).Error; err != nil {
		l.Errorw("Error users", "id", id, "user_id", usr.ID, "error", err)
		return nil, err
	}

	user := model.User{}
	copier.Copy(&user, &dbUser)

	return &user, nil
}

func getUser(ctx context.Context, usr *db.User, id uint) (*model.User, error) {

	d := pipeline.MustDB(ctx)
	l := pipeline.MustLogger(ctx)

	dbUser := db.User{}
	query := dbUser.Preload(d)

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
	query := (&db.User{}).Preload(d)

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

func findUsers(ctx context.Context, usr *db.User, roleID *int, contains *string) ([]*model.User, error) {

	d := pipeline.MustDB(ctx)
	l := pipeline.MustLogger(ctx)

	dbUsers := []db.User{}
	query := (&db.User{}).Preload(d)

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

	if roleID != nil {
		query = query.Where("users.role_id = ?", *roleID)
	}

	if contains != nil {
		str := fmt.Sprintf("%%%s%%", *contains)
		query = query.Where("users.first_name LIKE ? OR users.last_name LIKE ? OR users.email LIKE ?", str, str, str)
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
