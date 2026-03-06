package graph

import (
	"context"
	"fmt"

	"github.com/jinzhu/copier"
	"github.com/tuken/nix/aws"
	"github.com/tuken/nix/conf"
	"github.com/tuken/nix/db"
	"github.com/tuken/nix/graph/model"
	"github.com/tuken/nix/pipeline"
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
	copier.Copy(&workReport.User, &dbWorkReport.User)

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
