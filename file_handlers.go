package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// Check if storage can accept this file, then...
func createFile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	path := r.FormValue("path")
	size := r.FormValue("size")
	//uid := r.FormValue("uid")

	for stID, st := range storageServers {
		// Check if storage can accept this file
		resp, err := http.PostForm("http://"+st.LastAdr+":8080/api/files/isexist",
			url.Values{"path": {path}, "size": {"0"}})
		if err != nil {
			fmt.Println("getFile:", err)
			continue
		}
		fmt.Println(resp)
		// If some server responds OK, place file in waiting queue
		if resp.StatusCode == 200 {
			id := uint(0)
			for ; id < 100; id++ {
				if val, exist := itemsPending[id]; !exist {
					size, err := strconv.Atoi(size)
					if err != nil {
						w.WriteHeader(501)
						return
					}
					val.storageID = stID
					val.size = uint(size)
					val.path = path
					val.isDir = false
					itemsPending[id] = val
					break
				}
			}
			// Respond client with IP of storage and ID of file in queue
			renderOk(w, struct {
				IP string `json:"ip"`
				ID uint   `json:"id"`
			}{IP: st.LastAdr + ":8080", ID: id})
			//fmt.Println(itemsPending)
			return
		}
	}
	w.WriteHeader(417)
}

func getFile(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {
	path := r.FormValue("path")
	info := r.FormValue("info")
	//uid := r.FormValue("uid")
	//fmt.Println("PATH", path)
	items := []Files{}
	db.Where("uri = ?", path).Find(&items)
	fmt.Println(items)
	if len(items) == 0 || items[0].URI == "" {
		w.WriteHeader(404)
		return
	}

	for n, item := range items {
		adr := storageServers[items[n].Slave].LastAdr
		response, err := http.Get("http://" + adr + ":8080/api/status")
		if err != nil {
			fmt.Println("getFile:", err, response)
			continue
		}
		if info == "true" {
			renderOk(w, struct {
				IP   string `json:"ip"`
				Size uint   `json:"size"`
				Time int64  `json:"time"`
			}{IP: adr + ":8080", Size: item.Size, Time: item.CreatedTime})
		} else {

			renderOk(w, struct {
				IP string `json:"ip"`
			}{IP: adr + ":8080"})
		}
		return
	}
}

// Manage file
func manageFile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	path := r.FormValue("path")
	isDelete := r.FormValue("delete")
	newName := r.FormValue("new_name")
	//uid := r.FormValue("uid")

	item := Files{}
	db.Where("uri = ?", path).Find(&item)

	//adr := storageServers[item.Slave].LastAdr
	id := uint(0)
	if isDelete == "true" {
		for ; id < 100; id++ {
			if val, exist := itemsPending[id]; !exist {
				val.storageID = item.Slave
				val.path = path
				val.isDir = false
				val.isDelete = true
				itemsPending[id] = val
				break
			}
		}
	} else if newName != "" {
		for ; id < 100; id++ {
			if val, exist := itemsPending[id]; !exist {
				val.storageID = item.Slave
				val.path = path
				val.isDir = false
				val.newName = newName
				itemsPending[id] = val
				break
			}
		}
	} else {
		w.WriteHeader(406)
		return
	}
	fmt.Println("LastADDR", item.Slave, storageServers[item.Slave].LastAdr+":8080")
	renderOk(w, struct {
		IP string `json:"ip"`
		ID uint   `json:"id"`
	}{IP: storageServers[item.Slave].LastAdr + ":8080", ID: id})
	//fmt.Println(itemsPending)
	return
}
