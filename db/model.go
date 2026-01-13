package db

import (
	"database/sql"

	"gorm.io/gorm"
)

type Org struct {
	gorm.Model
	Name       string `gorm:"not null;comment:'組織名'"`
	PostalCode string `gorm:"not null;comment:'郵便番号'"`
	Address    string `gorm:"not null;comment:'住所'"`
	Note       string `gorm:"not null;comment:'備考'"`
}

type Role struct {
	gorm.Model
	Name        string         `gorm:"not null;comment:'役割名'"`
	Description sql.NullString `gorm:"comment:'役割説明'"`
}

type User struct {
	gorm.Model
	OrgID      uint           `gorm:"not null;comment:'組織ID'"`
	Org        Org            `gorm:"foreignKey:OrgID;references:ID;"`
	ParentID   sql.NullInt64  `gorm:"comment:'親ユーザーID'"`
	Parent     *User          `gorm:"foreignKey:ParentID;references:ID;"`
	RoleID     uint           `gorm:"comment:'役割ID'"`
	Role       Role           `gorm:"foreignKey:RoleID;references:ID;"`
	Email      string         `gorm:"not null;comment:'メールアドレス'"`
	Password   string         `gorm:"not null;comment:'パスワード'"`
	Name       string         `gorm:"not null;comment:'名前'"`
	FirstName  string         `gorm:"not null;comment:'名前（名）'"`
	LastName   string         `gorm:"not null;comment:'名前（姓）'"`
	PostalCode string         `gorm:"not null;comment:'郵便番号'"`
	Address    string         `gorm:"not null;comment:'住所'"`
	Gender     sql.NullString `gorm:"comment:'性別'"`
	Birthday   string         `gorm:"not null;comment:'生年月日'"`
	Note       string         `gorm:"comment:'備考'"`
}

type Field struct {
	gorm.Model
	OrgID       uint            `gorm:"not null;comment:'組織ID'"`
	Org         Org             `gorm:"foreignKey:OrgID;references:ID;"`
	UserID      uint            `gorm:"not null;comment:'ユーザーID'"`
	User        *User           `gorm:"foreignKey:UserID;references:ID;"`
	FieldCode   sql.NullString  `gorm:"comment:'フィールドコード'"`
	Name        string          `gorm:"not null;comment:'名前'"`
	Latitude    float64         `gorm:"not null;comment:'緯度'"`
	Longitude   float64         `gorm:"not null;comment:'経度'"`
	Elevation   sql.NullFloat64 `gorm:"comment:'標高'"`
	Area        sql.NullFloat64 `gorm:"comment:'面積'"`
	Boundary    Polygon         `gorm:"type:TEXT;comment:'境界情報（GeoJSON）'"`
	PostalCode  string          `gorm:"not null;comment:'郵便番号'"`
	Address     string          `gorm:"not null;comment:'住所'"`
	FieldTypeID uint            `gorm:"comment:'フィールドタイプID'"`
	Crop        string          `gorm:"comment:'作物'"`
	Note        string          `gorm:"comment:'備考'"`
}
