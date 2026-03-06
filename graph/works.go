package graph

import (
	"context"
	"errors"
	"fmt"

	"github.com/jinzhu/copier"
	"github.com/tuken/nix/aws"
	"github.com/tuken/nix/conf"
	"github.com/tuken/nix/db"
	"github.com/tuken/nix/graph/model"
	"github.com/tuken/nix/pipeline"
	"gorm.io/gorm"
)

func createWWorkReport(ctx context.Context, userID uint, input model.CreateWorkReportInput) (*model.WorkReport, error) {

	d := pipeline.MustDB(ctx)
	l := pipeline.MustLogger(ctx)

	isImage := input.Image != nil

	newWorkReport := &db.WorkReport{
		UserID:        userID,
		FieldID:       uint(input.FieldID),
		WorkDate:      input.WorkDate,
		WorkTypeID:    uint(input.WorkTypeID),
		CropVarietyID: uint(input.CropVarietyID),
		WeatherCode:   uint(input.WeatherCode),
		WorkDetail:    input.WorkDetail,
		IsImage:       isImage,
	}

	if err := d.Create(newWorkReport).Error; err != nil {
		l.Errorw("Error insert work_reports", "data", newWorkReport, "error", err)
		return nil, err
	}

	s3 := aws.NewS3Client()

	if isImage {

		l.Infow("Upload", "filename", input.Image.Filename, "size", input.Image.Size, "mime", input.Image.ContentType)

		key := fmt.Sprintf("WorkReports/%d.jpg", newWorkReport.ID)

		if err := s3.Upload(input.Image.File, conf.S3BucketName, key, input.Image.ContentType); err != nil {

			isImage = false
			d.Where("id = ?", newWorkReport.ID).Update("is_image", false)

			l.Errorw("Error upload s3 object", "key", key, "mime", input.Image.ContentType, "error", err)
			return nil, err
		}
	}

	dbWorkReport := db.WorkReport{}
	if err := d.Preload("User").Preload("Field").Preload("WorkType").Preload("CropVariety").Preload("Weather").Find(&dbWorkReport, newWorkReport.ID).Error; err != nil {
		return nil, err
	}

	workReport := model.WorkReport{}
	copier.Copy(&workReport, &dbWorkReport)

	if isImage {

		url, err := s3.GetSignedURL(conf.S3BucketName, fmt.Sprintf("WorkReports/%d.jpg", newWorkReport.ID))
		if err != nil {
			l.Errorw("Error get signed URL", "key", fmt.Sprintf("WorkReports/%d.jpg", newWorkReport.ID), "error", err)
			return nil, err
		}

		workReport.ImageURL = &url
	}

	return &workReport, nil

}

func getWorkReport(ctx context.Context, userID uint, id int) (*model.WorkReport, error) {

	d := pipeline.MustDB(ctx)
	l := pipeline.MustLogger(ctx)

	dbWorkReport := db.WorkReport{}
	if err := d.Preload("User").Preload("Field").Preload("WorkType").Preload("CropVariety").Preload("Weather").First(&dbWorkReport, "id = ? AND user_id = ?", id, userID).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Errorw("No work_reports", "id", id, "user_id", userID)
			return nil, nil
		} else {
			l.Errorw("Error work_reports", "id", id, "user_id", userID, "error", err)
			return nil, err
		}
	}

	workReport := model.WorkReport{}
	copier.Copy(&workReport, &dbWorkReport)

	if dbWorkReport.IsImage {

		s3 := aws.NewS3Client()

		url, err := s3.GetSignedURL(conf.S3BucketName, fmt.Sprintf("WorkReports/%d.jpg", dbWorkReport.ID))
		if err != nil {
			l.Errorw("Error get signed URL", "key", fmt.Sprintf("WorkReports/%d.jpg", dbWorkReport.ID), "error", err)
			return nil, err
		}

		workReport.ImageURL = &url
	}

	return &workReport, nil
}

func listWorkReports(ctx context.Context, user *db.User) ([]*model.WorkReport, error) {

	d := pipeline.MustDB(ctx)
	l := pipeline.MustLogger(ctx)

	dbWorkReports := []db.WorkReport{}
	query := d.Preload("User").Preload("Field").Preload("WorkType").Preload("CropVariety").Preload("Weather")

	switch user.Role.Name {

	case "admin":
		query = query.Joins("INNER JOIN fields f ON f.id = work_reports.field_id").Joins("INNER JOIN users u ON u.id = f.user_id").Joins("INNER JOIN orgs o ON o.id = u.org_id AND o.id = ?", user.OrgID)

	case "owner":
		query = query.Joins("INNER JOIN fields f ON f.id = work_reports.field_id AND f.user_id = ?", user.ID)

	case "worker":
		fieldIDs := []uint{}
		for _, f := range user.Fields {
			fieldIDs = append(fieldIDs, f.ID)
		}

		query = query.Joins("INNER JOIN fields f ON f.id = work_reports.field_id AND f.id IN (?)", fieldIDs)

	default:
		return nil, fmt.Errorf("unsupported role: %s", user.Role.Name)
	}

	if err := query.Find(&dbWorkReports).Error; err != nil {

		l.Errorw("Error work_reports", "user_id", user.ID, "role", user.Role.Name, "error", err)
		return nil, err
	}

	workReport := []*model.WorkReport{}

	for _, dbwp := range dbWorkReports {

		wp := model.WorkReport{}
		copier.Copy(&wp, &dbwp)

		if dbwp.IsImage {

			s3 := aws.NewS3Client()

			url, err := s3.GetSignedURL(conf.S3BucketName, fmt.Sprintf("WorkReports/%d.jpg", dbwp.ID))
			if err != nil {
				l.Errorw("Error get signed URL", "key", fmt.Sprintf("WorkReports/%d.jpg", dbwp.ID), "error", err)
				return nil, err
			}

			wp.ImageURL = &url
		}

		workReport = append(workReport, &wp)
	}

	return workReport, nil
}
