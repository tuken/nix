package db

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type Org struct {
	gorm.Model
	Name       string  `gorm:"not null;comment:'組織名'"`
	PostalCode string  `gorm:"not null;comment:'郵便番号'"`
	Address    string  `gorm:"not null;comment:'住所'"`
	Latitude   float64 `gorm:"not null;comment:'緯度'"`
	Longitude  float64 `gorm:"not null;comment:'経度'"`
	Note       string  `gorm:"not null;comment:'備考'"`
}

type Role struct {
	gorm.Model
	Name        string         `gorm:"not null;comment:'役割名'"`
	Description sql.NullString `gorm:"comment:'役割説明'"`
}

type User struct {
	gorm.Model
	OrgID       uint           `gorm:"not null;comment:'組織ID（orgs.id）'"`
	Org         *Org           `gorm:"foreignKey:OrgID;references:ID;"`
	ParentID    sql.NullInt64  `gorm:"comment:'親ユーザーID（roleがadmin,ownerの場合はNULL）'"`
	Parent      *User          `gorm:"foreignKey:ParentID;references:ID;"`
	RoleID      uint           `gorm:"comment:'役割ID（roles.id）'"`
	Role        *Role          `gorm:"foreignKey:RoleID;references:ID;"`
	Email       string         `gorm:"not null;comment:'メールアドレス'"`
	Password    string         `gorm:"not null;comment:'パスワード'"`
	FarmName    sql.NullString `gorm:"not null;comment:'農場名/法人名（オーナーの場合のみセットされる）'"`
	LastName    string         `gorm:"not null;comment:'姓'"`
	FirstName   string         `gorm:"not null;comment:'名'"`
	PostalCode  string         `gorm:"not null;comment:'郵便番号'"`
	Address     string         `gorm:"not null;comment:'住所'"`
	Gender      string         `gorm:"type:enum('male','female','other');default:'male';comment:'性別（male：男性、female：女性、other：その他）'"`
	Birthday    time.Time      `gorm:"not null;comment:'生年月日'"`
	Note        string         `gorm:"comment:'備考'"`
	LastLoginAt sql.NullTime   `gorm:"comment:'最終ログイン日時'"`
	Fields      []*Field       `gorm:"many2many:field_users;joinForeignKey:UserID;joinReferences:FieldID"`
}

func (u *User) Preload(db *gorm.DB) *gorm.DB {
	return db.Preload("Org").Preload("Parent").Preload("Role").Preload("Fields")
}

type FieldType struct {
	gorm.Model
	Name      string `gorm:"not null;comment:'圃場タイプ名'"`
	SortOrder uint8  `gorm:"not null;default:0;comment:'表示順'"`
}

type FieldState struct {
	gorm.Model
	Name        string         `gorm:"not null;comment:'圃場状態名'"`
	Description sql.NullString `gorm:"comment:'説明'"`
}

type Field struct {
	gorm.Model
	UserID      uint            `gorm:"not null;comment:'所有者ユーザーID（users.id）'"`
	User        *User           `gorm:"foreignKey:UserID;references:ID;"`
	FieldTypeID uint            `gorm:"comment:'圃場タイプID（field_types.id）'"`
	FieldType   *FieldType      `gorm:"foreignKey:FieldTypeID;references:ID;"`
	FieldCode   sql.NullString  `gorm:"comment:'圃場の外部連携用コード（任意）'"`
	Name        string          `gorm:"not null;comment:'圃場名'"`
	Latitude    float64         `gorm:"not null;comment:'緯度'"`
	Longitude   float64         `gorm:"not null;comment:'経度'"`
	Elevation   sql.NullFloat64 `gorm:"comment:'標高(m)'"`
	Area        sql.NullFloat64 `gorm:"comment:'面積(m²)'"`
	// Boundary    *Polygon        `gorm:"type:TEXT;comment:'圃場の境界ポリゴン（WKT形式：POLYGON）'"`
	Boundary     sql.NullString
	FieldStateID uint           `gorm:"not null;default:1;comment:'圃場状態'"`
	FieldState   *FieldState    `gorm:"foreignKey:FieldStateID;references:ID;"`
	PostalCode   string         `gorm:"not null;comment:'郵便番号'"`
	Address      string         `gorm:"not null;comment:'住所'"`
	Crop         sql.NullString `gorm:"comment:'栽培作物（米、麦、トマトなど）'"`
	// Status     string         `gorm:"type:enum('cultivated','fallow','abandoned');default:'cultivated';comment:'利用状態（cultivated：耕作中、fallow：休耕中、abandoned：耕作放棄）'"`
	Note  string  `gorm:"comment:'備考'"`
	Users []*User `gorm:"many2many:field_users;joinForeignKey:FieldID;joinReferences:UserID"`
}

func (u *Field) Preload(db *gorm.DB) *gorm.DB {
	return db.Preload("User").Preload("FieldType").Preload("FieldState").Preload("Users")
}

type FieldUser struct {
	FieldID uint   `gorm:"not null;comment:'圃場ID（fields.id）'"`
	Field   *Field `gorm:"foreignKey:FieldID"`
	UserID  uint   `gorm:"not null;comment:'作業者ユーザーID（users.id）'"`
	User    *User  `gorm:"foreignKey:UserID"`
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
	CropItemID uint      `gorm:"not null;comment:'品目ID'"`
	CropItem   *CropItem `gorm:"foreignKey:CropItemID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Name       string    `gorm:"not null;comment:'品種名（コシヒカリ/あきたこまち/ササニシキ）'"`
	SortOrder  uint8     `gorm:"not null;default:0;comment:'表示順'"`
}

type WorkReport struct {
	gorm.Model
	UserID        uint            `gorm:"not null;comment:'作業者ユーザーID（users.id）'"`
	User          *User           `gorm:"foreignKey:UserID;references:ID;"`
	FieldID       uint            `gorm:"not null;comment:'圃場ID（fields.id）'"`
	Field         *Field          `gorm:"foreignKey:FieldID;references:ID;"`
	WorkDate      time.Time       `gorm:"not null;comment:'作業日'"`
	WorkHours     sql.NullFloat64 `gorm:"type:decimal(3,1);comment:'作業時間（時間単位）'"`
	WorkTypeID    uint            `gorm:"not null;comment:'作業タイプID（work_types.id）'"`
	WorkType      *WorkType       `gorm:"foreignKey:WorkTypeID;references:ID;"`
	CropVarietyID uint            `gorm:"not null;comment:'品種ID（crop_varieties.id）'"`
	CropVariety   *CropVariety    `gorm:"foreignKey:CropVarietyID;references:ID;"`
	WeatherCode   uint            `gorm:"comment:'天候コード（open-meteoで使用しているコード）'"`
	Weather       *Weather        `gorm:"foreignKey:WeatherCode;references:Code;"`
	WorkDetail    string          `gorm:"type:text;comment:'作業詳細'"`
	IsImage       bool            `gorm:"comment:'画像の有無（0：なし、1：あり）'"`
	Temperature   sql.NullFloat64 `gorm:"comment:'気温（℃）'"`
	Humidity      sql.NullFloat64 `gorm:"comment:'湿度（％）'"`
	CropCondition sql.NullString  `gorm:"comment:'作物状況（生育状況・病害虫・水位など）'"`
	// CreatedBy     uint            `gorm:"not null;comment:'作成者ID'"`
	// UpdatedBy     uint            `gorm:"not null;comment:'更新者ID'"`
}

func (w *WorkReport) Preload(db *gorm.DB) *gorm.DB {
	return db.Preload("User").Preload("User.Parent").Preload("Field").Preload("WorkType").Preload("CropVariety").Preload("CropVariety.CropItem").Preload("Weather")
}

type Weather struct {
	Code     uint   `gorm:"primaryKey;comment:'天気コード'"`
	Japanese string `gorm:"not null;comment:'天気の日本語表記'"`
}
