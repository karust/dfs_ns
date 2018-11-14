package main

// Slaves ... Account table
type Slaves struct {
	ID          uint `gorm:"not null;unique"`
	LastAddr    string
	CreatedTime int64
	LastAuth    int64
}

// Files ...
type Files struct {
	ID          uint   `gorm:"not null;unique"`
	Name        string `gorm:"unique_index:path"`
	URL         string `gorm:"unique_index:path"`
	URI         string `gorm:"not null;unique"`
	CreatedTime int64
	Size        uint
	Slave       uint
	IsDir       bool
}

/*
type Group struct {
	//gorm.Model
	ID         uint
	Name       string `gorm:"not null;unique"`
	AdminID    uint   `gorm:"not null"`
	MembersNum uint32
}

type GroupMemebers struct {
	//gorm.Model
	ID  uint
	UID uint `gorm:"unique_index:UGID"`
	GID uint `gorm:"unique_index:UGID"`
}
*/
