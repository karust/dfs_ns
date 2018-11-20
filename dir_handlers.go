package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
)

// Information about files in directory
func createDir(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {
	path := r.FormValue("path")
	//uid := r.FormValue("uid")

	for stID, st := range storageServers {
		// ?? Check if storage can accept this directory
		resp, err := http.PostForm("http://"+st.LastAdr+":8080/api/files/isexist",
			url.Values{"path": {path}, "size": {"0"}})
		if err != nil {
			// TODO: if dir exists send response to client
			fmt.Println("createDir:", err)
			continue
		}

		// If some server responds OK, place dir in waiting queue
		if resp.StatusCode == 200 {
			id := uint(0)
			for ; id < 100; id++ {
				if val, exist := itemsPending[id]; !exist {
					val.storageID = stID
					val.path = path
					val.isDir = true
					itemsPending[id] = val
					break
				}
			}
			// Respond client with IP of storage and ID of dir in queue
			renderOk(w, struct {
				IP string `json:"ip"`
				ID uint   `json:"id"`
			}{IP: st.LastAdr + ":8080", ID: id})
			fmt.Println("Items pending", itemsPending)
			return
		}
	}
	w.WriteHeader(417)
}

// Information about files in directory
func getDirFiles(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {
	path := r.FormValue("path")
	//uid := r.FormValue("uid")

	items := []Files{}
	db.Where("url = ?", path).Find(&items)
	//fmt.Println(items)

	resp := []FileDir{}
	for _, item := range items {
		if item.IsMain {
			resp = append(resp, FileDir{Name: item.Name, Path: item.URI, Size: item.Size,
				CrTime: item.CreatedTime, IsDir: item.IsDir})
		}
	}

	//fmt.Println(resp)
	renderOk(w, struct {
		Items []FileDir `json:"items"`
	}{Items: resp})
}

// Manage directory
func manageDir(w http.ResponseWriter, r *http.Request,
	_ httprouter.Params) {
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
				val.isDir = true
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
				val.isDir = true
				val.newName = newName
				itemsPending[id] = val
				break
			}
		}
	} else {
		w.WriteHeader(406)
		return
	}

	renderOk(w, struct {
		IP string `json:"ip"`
		ID uint   `json:"id"`
	}{IP: storageServers[item.Slave].LastAdr + ":8080", ID: id})
	return
}
