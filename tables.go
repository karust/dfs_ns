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
	ID          uint `gorm:"not null;unique"`
	Name        string
	URL         string
	URI         string `gorm:"unique_index:replica"`
	CreatedTime int64
	Size        uint
	Slave       uint `gorm:"unique_index:replica"`
	IsDir       bool
	IsMain      bool
}

// Users ... Users table
type Users struct {
	ID          uint `gorm:"not null;unique"`
	Login       string
	CreatedTime int64
	Pass        string
}
