package main

// Slaves ... Account table
type Slaves struct {
	ID          uint `gorm:"not null;unique"`
	LastAddr    string
	CreatedTime int64
	LastAuth    int64
}

// File ...
type File struct {
	ID          uint `gorm:"not null;unique"`
	name        string
	uri         string
	CreatedTime int64
	slave       uint
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
