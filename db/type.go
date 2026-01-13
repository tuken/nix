package db

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Polygon struct {
	WKT string `gorm:"type:geometry(Polygon,4326)" json:"wkt"`
}

// func (p *Polygon) Scan(v any) error {

// 	if b, ok := v.([]byte); ok {

// 		p.WKT = string(b)
// 		return nil
// 	}

// 	return errors.New("invalid type for Polygon")
// }

func (p Polygon) GormDataType() string {

	return "geometry"
}

func (p Polygon) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {

	return clause.Expr{
		SQL:  "ST_GeomFromText(?)",
		Vars: []any{fmt.Sprintf("POLYGON((%s))", p.WKT)},
	}
}
