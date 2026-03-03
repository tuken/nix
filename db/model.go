package db

import (
	"database/sql"
	"time"

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

type FieldType struct {
	gorm.Model
	Name      string `gorm:"not null;comment:'圃場タイプ名'"`
	SortOrder uint8  `gorm:"not null;default:0;comment:'表示順'"`
	Note      string `gorm:"comment:'備考'"`
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
	FieldType   *FieldType      `gorm:"foreignKey:FieldTypeID;references:ID;"`
	Crop        string          `gorm:"comment:'作物'"`
	Note        string          `gorm:"comment:'備考'"`
}

type WorkType struct {
	gorm.Model
	Name      string `gorm:"not null;comment:'作業タイプ名'"`
	SortOrder uint8  `gorm:"not null;default:0;comment:'表示順'"`
}

type CropItem struct {
	gorm.Model
	Name      string `gorm:"not null;comment:'品目名（米/大豆/麦）'"`
	SortOrder uint8  `gorm:"not null;default:0;comment:'表示順'"`
}

type CropVariety struct {
	gorm.Model
	Name      string   `gorm:"not null;comment:'品種名（コシヒカリ/あきたこまち/ササニシキ）'"`
	ItemID    uint     `gorm:"not null;comment:'品目ID'"`
	CropItem  CropItem `gorm:"foreignKey:ItemID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	SortOrder uint8    `gorm:"not null;default:0;comment:'表示順'"`
}

type WorkReport struct {
	gorm.Model
	UserID        uint            `gorm:"not null;comment:'利用者ID'"`
	User          *User           `gorm:"foreignKey:UserID;references:ID;"`
	FieldID       uint            `gorm:"not null;comment:'圃場ID'"`
	Field         *Field          `gorm:"foreignKey:FieldID;references:ID;"`
	WorkDate      time.Time       `gorm:"not null;comment:'作業日'"`
	WorkTypeID    uint            `gorm:"not null;comment:'作業タイプID'"`
	WorkType      WorkType        `gorm:"foreignKey:WorkTypeID;references:ID;"`
	CropVarietyID uint            `gorm:"not null;comment:'品種ID'"`
	CropVariety   CropVariety     `gorm:"foreignKey:CropVarietyID;references:ID;"`
	WeatherCode   uint            `gorm:"comment:'天候コード（open-meteoで使用しているコード）'"`
	IsImage       bool            `gorm:"comment:'画像の有無（0：なし、1：あり）'"`
	Temperature   sql.NullFloat64 `gorm:"comment:'気温（℃）'"`
	Humidity      sql.NullFloat64 `gorm:"comment:'湿度（％）'"`
	CropCondition sql.NullString  `gorm:"comment:'作物状況（生育状況・病害虫・水位など）'"`
	Note          sql.NullString  `gorm:"comment:'備考'"`
	CreatedBy     uint            `gorm:"not null;comment:'作成者ID'"`
	UpdatedBy     uint            `gorm:"not null;comment:'更新者ID'"`
}

// type WeatherCode struct {
// 	Code     uint   `gorm:"primaryKey;comment:'天気コード'"`
// 	Japanese string `gorm:"not null;comment:'天気の日本語表記'"`
// }
