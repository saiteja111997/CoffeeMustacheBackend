package structures

type CuratedCartMenuItem struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	CafeID      uint       `gorm:"not null" json:"cafe_id"`
	Category    string     `gorm:"type:varchar(50);not null" json:"category"`
	SubCategory string     `gorm:"type:varchar(50)" json:"sub_category"`
	Name        string     `gorm:"type:varchar(100);not null" json:"name"`
	Description string     `gorm:"type:text" json:"description"`
	Price       float64    `gorm:"type:decimal(10,2);not null" json:"price"`
	FoodType    string     `gorm:"type:varchar(10);not null" json:"food_type"`
	Cuisine     Cuisine    `gorm:"type:varchar(50)" json:"cuisine"`     // Cuisine as enum
	CMCategory  CMCategory `gorm:"type:varchar(50)" json:"cm_category"` // CM Category as enum
	ServingSize string     `gorm:"type:varchar(50)" json:"serving_size"`
}
