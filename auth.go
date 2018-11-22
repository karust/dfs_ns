package main

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
)

func authUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	login := r.FormValue("login")
	fmt.Println("Login:", login)
	pass := r.FormValue("pass")
	fmt.Println("Pass:", pass)

	user := Users{}
	db.Where("login = ?", login).Find(&user)

	if user.Login == "" {
		db.Create(&Users{
			Login:       login,
			Pass:        pass,
			CreatedTime: time.Now().Unix()})

		db.Where("login = ?", login).Find(&user)
		renderOk(w, struct {
			UID uint `json:"uid"`
		}{UID: user.ID})
	}

	if user.Pass == pass {
		renderOk(w, struct {
			UID uint `json:"uid"`
		}{UID: user.ID})
	} else {
		w.WriteHeader(401)
	}

}

func register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

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

func authStorageServ(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

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
		//if slave.ID == 0 {
		//	w.WriteHeader(401)
		//	return
		//}
		fmt.Println("[NEW_CONN]", r.RemoteAddr)
		slave.ID = claims.ID
		slave.LastAddr, _, _ = net.SplitHostPort(r.RemoteAddr)
		slave.LastAuth = time.Now().Unix()
		//fmt.Println(slave)
		db.Save(&slave)

		storage := Storage{LastAdr: slave.LastAddr}
		storageServers[claims.ID] = storage

		renderOk(w, struct {
			UID uint `json:"uid"`
		}{UID: slave.ID})
		//w.WriteHeader(200)
		fmt.Println("[CURR_STORAGES]", storageServers)
	} else {
		fmt.Println("2)", err)
		w.WriteHeader(401)
	}
}

func ping(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	strUID := r.Header.Get("uid")
	uid, err := strconv.Atoi(strUID)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	if _, ok := storageServers[uint(uid)]; ok {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(400)
	}
}
