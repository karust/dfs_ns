package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
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

//ItemsPending ...
type ItemsPending struct {
	path      string
	size      uint
	storageID uint
	isDir     bool
	newName   string
	isDelete  bool
}

var storageServers map[uint]Storage    // StorageID : Storage
var itemsPending map[uint]ItemsPending // FileID : ItemsPending

var db *gorm.DB

func main() {
	storageServers = make(map[uint]Storage)
	storageServers[1] = Storage{LastAdr: "10.1.1.146"}
	storageServers[2] = Storage{LastAdr: "::1"}

	itemsPending = make(map[uint]ItemsPending)

	db, _ = gorm.Open("sqlite3", "test.db")
	db.AutoMigrate(&Slaves{}, &Files{})

	router := httprouter.New()
	router.POST("/dir/get", getDirFiles)
	router.POST("/dir/create", createDir)
	router.POST("/dir/manage", manageDir)

	router.POST("/file/get", getFile)
	router.POST("/file/create", createFile)
	router.POST("/file/manage", manageFile)

	router.POST("/reg", register)
	router.POST("/auth", authStorageServ)
	router.POST("/storage/confirm", confirmStorage)
	http.ListenAndServe(":8080", router)
}

func confirmStorage(w http.ResponseWriter, r *http.Request,
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
		if _, exist := storageServers[claims.ID]; exist {
			itemID := r.Header.Get("id")
			fID, err := strconv.Atoi(itemID)
			if err != nil {
				w.WriteHeader(501)
				return
			}
			if _, exist := itemsPending[uint(fID)]; !exist {
				w.WriteHeader(406)
				return
			}

			item := itemsPending[uint(fID)]
			if item.newName != "" {
				file := Files{}
				db.Where("uri = ?", item.path).Find(&file)
				file.Name = item.newName
				db.Save(&file)
			} else if item.isDelete {
				db.Delete(Files{}, "uri = ?", item.path)

			} else {
				split := strings.Split(item.path, "/")
				splen := len(split)
				url := strings.Join(split[:splen-1], "/")
				if url == "" {
					url = "/"
				}
				if item.storageID == claims.ID && !item.isDir {
					db.Create(&Files{Name: split[splen-1],
						URL:         url,
						URI:         item.path,
						Size:        item.size,
						Slave:       item.storageID,
						CreatedTime: time.Now().Unix(),
						IsDir:       false})
				} else if item.storageID == claims.ID && item.isDir {
					db.Create(&Files{Name: split[splen-1],
						URL:         url,
						URI:         item.path,
						Slave:       item.storageID,
						CreatedTime: time.Now().Unix(),
						IsDir:       true})
				} else {
					w.WriteHeader(403)
					return
				}
			}
			delete(itemsPending, uint(fID))
			w.WriteHeader(200)
			return
		}
	}
	w.WriteHeader(401)
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
		slave.LastAddr, _, _ = net.SplitHostPort(r.RemoteAddr)
		slave.LastAuth = time.Now().Unix()
		fmt.Println(slave)
		db.Save(&slave)

		storage := Storage{LastAdr: slave.LastAddr}
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
