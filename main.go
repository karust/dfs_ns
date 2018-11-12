package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/julienschmidt/httprouter"
)

//Storage ...
type Storage struct {
	LastAdr string
}

var storageServers map[uint]Storage
var db *gorm.DB

func main() {
	storageServers = make(map[uint]Storage)

	db, _ = gorm.Open("sqlite3", "test.db")
	db.AutoMigrate(&Slaves{}, &File{})

	router := httprouter.New()
	router.POST("/dir/get", getDirFiles)
	router.POST("/dir/manage", manageDir)
	router.POST("/file/get", getFile)
	router.POST("/file/manage", manageFile)

	router.POST("/reg", register)
	router.POST("/auth", authStorageServ)
	http.ListenAndServe(":8080", router)
}

// Manage file
func manageFile(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {
	path := r.FormValue("path")
	fmt.Println("Path:", path)
	isDelete := r.FormValue("delete")
	fmt.Println("Delete:", isDelete)
	newName := r.FormValue("new_name")
	fmt.Println("New name:", newName)
}

func getFile(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {
	path := r.FormValue("path")
	fmt.Println("Path:", path)
	isInfo := r.FormValue("info")
	fmt.Println("Delete:", isInfo)
}

// Manage directory
func manageDir(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {
	path := r.FormValue("path")
	fmt.Println("Path:", path)
	isDelete := r.FormValue("delete")
	fmt.Println("Delete:", isDelete)
	newName := r.FormValue("new_name")
	fmt.Println("New name:", newName)
}

// Information about files in directory
func getDirFiles(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {
	path := r.FormValue("path")
	fmt.Println("Path:", path)
	isCreate := r.FormValue("create")
	fmt.Println("Create:", isCreate)
}

func register(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {

	t := time.Now().Unix()
	sl := &Slaves{CreatedTime: t, LastAddr: r.RemoteAddr}
	db.Create(sl)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": sl.ID,
	})

	tokenString, err := token.SignedString([]byte("123"))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
	} else {
		renderOk(w, struct {
			Token string `json:"token"`
		}{Token: tokenString})
	}
}

func authStorageServ(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {

	tokenString := r.Header.Get("Authorization")

	token, err := jwt.ParseWithClaims(tokenString, &Auth{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("123"), nil
	})
	if err != nil {
		fmt.Println("1)", err)
		w.WriteHeader(401)
		return
	}

	if claims, ok := token.Claims.(*Auth); ok && token.Valid {
		slave := Slaves{}
		db.Where("id = ?", claims.ID).Find(&slave)
		if slave.ID == 0 {
			w.WriteHeader(401)
			return
		}
		slave.ID = claims.ID
		slave.LastAddr = r.RemoteAddr
		slave.LastAuth = time.Now().Unix()
		fmt.Println(slave)
		db.Save(&slave)

		storage := Storage{LastAdr: r.RemoteAddr}
		storageServers[claims.ID] = storage
		w.WriteHeader(200)
		fmt.Println(storageServers)
	} else {
		fmt.Println("2)", err)
		w.WriteHeader(401)
	}
}

func renderOk(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(data)
}
