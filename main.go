package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
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
	//storageServers[16] = Storage{LastAdr: "10.1.1.147"}
	//storageServers[17] = Storage{LastAdr: "10.1.1.240"}

	itemsPending = make(map[uint]ItemsPending)

	db, _ = gorm.Open("sqlite3", "test.db")
	db.AutoMigrate(&Slaves{}, &Files{}, &Users{})

	router := httprouter.New()
	router.POST("/dir/get", getDirFiles)
	router.POST("/dir/create", createDir)
	router.POST("/dir/manage", manageDir)

	router.POST("/file/get", getFile)
	router.POST("/file/create", createFile)
	router.POST("/file/manage", manageFile)

	router.POST("/reg", register)
	router.POST("/auth/user", authUser)
	router.POST("/auth/storage", authStorageServ)
	router.POST("/storage/confirm", confirmStorage)
	router.POST("/ping", ping)

	handler := cors.Default().Handler(router)
	http.ListenAndServe(":8888", handler)
}

func confirmStorage(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {

	tokenString := r.Header.Get("Authorization")
	isReplicate := r.Header.Get("isrepl")

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
				file.URI = file.URL + item.newName
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

				fmt.Println("WWWWWWWWWWW:", isReplicate, isReplicate == "true")
				if isReplicate == "true" {
					if item.storageID == claims.ID && !item.isDir {
						file := Files{Name: split[splen-1],
							URL:         url,
							URI:         item.path,
							Size:        item.size,
							Slave:       item.storageID,
							CreatedTime: time.Now().Unix(),
							IsDir:       false,
							IsMain:      false}
						db.Create(&file)
					} else if item.storageID == claims.ID && item.isDir {
						dir := Files{Name: split[splen-1],
							URL:         url,
							URI:         item.path,
							Slave:       item.storageID,
							CreatedTime: time.Now().Unix(),
							IsDir:       true,
							IsMain:      false}
						db.Create(&dir)
					}
				} else {
					if item.storageID == claims.ID && !item.isDir {
						file := Files{Name: split[splen-1],
							URL:         url,
							URI:         item.path,
							Size:        item.size,
							Slave:       item.storageID,
							CreatedTime: time.Now().Unix(),
							IsDir:       false,
							IsMain:      true}
						db.Create(&file)
						replicate(file.ID)
					} else if item.storageID == claims.ID && item.isDir {
						dir := Files{Name: split[splen-1],
							URL:         url,
							URI:         item.path,
							Slave:       item.storageID,
							CreatedTime: time.Now().Unix(),
							IsDir:       true,
							IsMain:      true}
						db.Create(&dir)
						replicate(dir.ID)
					}
				}
				w.WriteHeader(403)
				return
			}
			delete(itemsPending, uint(fID))
			w.WriteHeader(200)
			return
		}
	}
	w.WriteHeader(401)
}

func replicate(ID uint) {
	items := []Files{}
	db.Where("id = ?", ID).Find(&items)

	storages := make(map[uint]bool)

	// Get Slaves where item stored
	for _, s := range items {
		storages[s.Slave] = true
	}

	// Find IP of server with this file
	stIP := "0.0.0.0"
	for i, s := range storageServers {
		if storages[i] == true {
			stIP = s.LastAdr
			continue
		}
	}

	if len(storages) <= 1 {
		if items[0].IsDir {
			id := uint(0)
			for ; id < 100; id++ {
				if _, exist := itemsPending[id]; !exist {
					break
				}
			}

			for i, s := range storageServers {
				if _, ok := storages[i]; ok {
					continue
				} else {
					val := itemsPending[id]
					val.storageID = i
					val.path = items[0].URI
					val.isDir = true
					itemsPending[id] = val

					ID := strconv.Itoa(int(id))
					resp, err := http.PostForm("http://"+s.LastAdr+":8080/api/files/createfolder",
						url.Values{"path": {items[0].URL}, "name": {items[0].Name}, "id": {ID}, "isrepl": {"true"}})
					if err != nil {
						fmt.Println("Response err:", err)
						continue
					}
					fmt.Println("Response", resp)
					return
				}
			}

		} else {
			id := uint(0)
			for ; id < 100; id++ {
				if _, exist := itemsPending[id]; !exist {
					break
				}
			}

			for i, s := range storageServers {
				if _, ok := storages[i]; ok {
					continue
				} else {
					location := "http://" + stIP + ":8080/api/files/download?path=" + items[0].URI
					val := itemsPending[id]
					val.storageID = i
					val.path = items[0].URI
					val.size = items[0].Size
					val.isDir = false
					itemsPending[id] = val

					ID := strconv.Itoa(int(id))
					resp, err := http.PostForm("http://"+s.LastAdr+":8080/api/files/synchronize",
						url.Values{"path": {items[0].URL}, "url": {location}, "id": {ID}, "isrepl": {"true"}})
					if err != nil {
						fmt.Println("Response err:", err)
						continue
					}
					fmt.Println("Response", resp)
					return
				}
			}
		}
	}
}

func renderOk(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(data)
}
