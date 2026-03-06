package graph

import (
	"context"
	"errors"
	"fmt"

	"github.com/jinzhu/copier"
	"github.com/tuken/nix/db"
	"github.com/tuken/nix/graph/model"
	"github.com/tuken/nix/pipeline"
	"gorm.io/gorm"
)

func createField(ctx context.Context, user *db.User, input model.CreateFieldInput) (*model.Field, error) {

	d := pipeline.MustDB(ctx)
	l := pipeline.MustLogger(ctx)

	switch user.Role.Name {

	case "admin":

	case "owner":
		return nil, fmt.Errorf("forbidden operation by owner")

	case "worker":
		return nil, fmt.Errorf("forbidden operation by worker")

	default:
		return nil, fmt.Errorf("unsupported role: %s", user.Role.Name)
	}

	newField := &db.Field{
		UserID:       user.ID,
		FieldTypeID:  uint(input.FieldTypeID),
		Name:         input.Name,
		Latitude:     input.Latitude,
		Longitude:    input.Longitude,
		FieldStateID: uint(input.FieldStateID),
		PostalCode:   input.PostalCode,
		Address:      input.Address,
		Note:         input.Note,
	}

	if input.FieldCode != nil {
		newField.FieldCode.Scan(*input.FieldCode)
	}

	if input.Elevation != nil {
		newField.Elevation.Scan(*input.Elevation)
	}

	if input.Area != nil {
		newField.Area.Scan(*input.Area)
	}

	if err := d.Create(newField).Error; err != nil {
		l.Errorw("Error insert fields", "data", newField, "error", err)
		return nil, err
	}

	dbField := db.Field{}
	if err := d.Preload("User").Preload("FieldType").Preload("Users").First(&dbField, newField.ID).Error; err != nil {
		return nil, err
	}

	field := model.Field{}
	copier.Copy(&field, &dbField)

	return &field, nil
}

func getField(ctx context.Context, user *db.User, id int) (*model.Field, error) {

	d := pipeline.MustDB(ctx)
	l := pipeline.MustLogger(ctx)

	dbField := db.Field{}
	query := d.Preload("User").Preload("FieldType").Preload("Users")

	switch user.Role.Name {

	case "admin":
		query = query.Joins("INNER JOIN users u ON u.id = fields.user_id").Joins("INNER JOIN orgs o ON o.id = u.org_id AND o.id = ?", user.OrgID)

	case "owner":
		query = query.Where("fields.user_id = ?", user.ID)

	case "worker":
		fieldIDs := []uint{}
		for _, f := range user.Fields {
			fieldIDs = append(fieldIDs, f.ID)
		}

		query = query.Where("fields.id IN (?)", fieldIDs)

	default:
		return nil, fmt.Errorf("unsupported role: %s", user.Role.Name)
	}

	if err := query.First(&dbField, id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Errorw("No fields", "id", id, "user_id", user.ID)
			return nil, nil
		} else {
			l.Errorw("Error fields", "id", id, "user_id", user.ID, "error", err)
			return nil, err
		}
	}

	field := model.Field{}
	copier.Copy(&field, &dbField)

	return &field, nil
}

func listFields(ctx context.Context, user *db.User) ([]*model.Field, error) {

	d := pipeline.MustDB(ctx)
	l := pipeline.MustLogger(ctx)

	dbFields := []db.Field{}
	query := d.Preload("User").Preload("FieldType").Preload("Users")

	switch user.Role.Name {

	case "admin":
		query = query.Joins("INNER JOIN users u ON u.id = fields.user_id").Joins("INNER JOIN orgs o ON o.id = u.org_id AND o.id = ?", user.OrgID)

	case "owner":
		query = query.Where("fields.user_id = ?", user.ID)

	case "worker":
		fieldIDs := []uint{}
		for _, f := range user.Fields {
			fieldIDs = append(fieldIDs, f.ID)
		}

		query = query.Where("fields.id IN (?)", fieldIDs)

	default:
		return nil, fmt.Errorf("unsupported role: %s", user.Role.Name)
	}

	if err := query.Find(&dbFields).Error; err != nil {
		l.Errorw("Error fields", "user_id", user.ID, "role", user.Role.Name, "error", err)
		return nil, err
	}

	fields := []*model.Field{}

	for _, dbf := range dbFields {

		f := model.Field{}
		copier.Copy(&f, &dbf)

		fields = append(fields, &f)
	}

	return fields, nil
}

func findFields(ctx context.Context, user *db.User, ownerID *int, fieldTypeID *int, fieldStateID *int) ([]*model.Field, error) {

	d := pipeline.MustDB(ctx)
	l := pipeline.MustLogger(ctx)

	dbFields := []db.Field{}
	query := d.Preload("User").Preload("FieldType").Preload("Users")

	switch user.Role.Name {

	case "admin":
		query = query.Joins("INNER JOIN users u ON u.id = fields.user_id").Joins("INNER JOIN orgs o ON o.id = u.org_id AND o.id = ?", user.OrgID)

	case "owner":
		query = query.Where("fields.user_id = ?", user.ID)

	case "worker":
		fieldIDs := []uint{}
		for _, f := range user.Fields {
			fieldIDs = append(fieldIDs, f.ID)
		}

		query = query.Where("fields.id IN (?)", fieldIDs)

	default:
		return nil, fmt.Errorf("unsupported role: %s", user.Role.Name)
	}

	if ownerID != nil {
		query = query.Where("fields.user_id = ?", *ownerID)
	}

	if fieldTypeID != nil {
		query = query.Where("fields.field_type_id = ?", *fieldTypeID)
	}

	if fieldStateID != nil {
		query = query.Where("fields.field_state_id = ?", *fieldStateID)
	}

	if err := query.Find(&dbFields).Error; err != nil {
		l.Errorw("Error fields", "user_id", user.ID, "role", user.Role.Name, "error", err)
		return nil, err
	}

	fields := []*model.Field{}

	for _, dbf := range dbFields {

		f := model.Field{}
		copier.Copy(&f, &dbf)

		fields = append(fields, &f)
	}

	return fields, nil
}
