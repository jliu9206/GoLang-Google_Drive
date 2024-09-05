package main

import (
	myDB "GoLang-GoogleDrive/db/mysql"
	"GoLang-GoogleDrive/handler"
	"GoLang-GoogleDrive/model"
	"fmt"
	"net/http"
)

// "user": "user1",
// "access_key": "L4ZATJRF4G1YX559NBDE",
// "secret_key": "OFSX0li3BmMXAP4ZyBURF5Dga58iLPzBnZCsUOXC"
func main() {
	username := "root"
	password := "123456"
	address := "127.0.0.1:3307"
	dbname := "gogoserver"

	myDB.Initialize(username, password, address, dbname)
	defer myDB.DBClose() // make sure main closes also ends connection
	model.SetDB(myDB.DBConn())

	// conn := myRedis.RedisPool().Get()
	// defer conn.Close()
	// if _, err := conn.Do("PING"); err != nil {
	// 	log.Fatal("Failed!!!", err)
	// }
	// fmt.Println("Redis OK!")

	http.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/file/upload", handler.HTTPInterceptor(handler.UploadHandler))
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler)
	http.HandleFunc("/file/meta", handler.HTTPInterceptor(handler.GetFileMetaHandler))
	http.HandleFunc("/file/query", handler.HTTPInterceptor(handler.FileQueryHandler))
	http.HandleFunc("/file/download", handler.HTTPInterceptor(handler.DownloadHandler))
	http.HandleFunc("/file/update", handler.HTTPInterceptor(handler.FileMetaUpdateHandler))
	http.HandleFunc("/file/delete", handler.HTTPInterceptor(handler.FileDeleteHandler))
	http.HandleFunc("/file/fastupload", handler.HTTPInterceptor(handler.TryFastUploadHandler))

	http.HandleFunc("/file/mpupload/init", handler.HTTPInterceptor(handler.InitialMultipartUploadHandler))
	http.HandleFunc("/file/mpupload/uppart", handler.HTTPInterceptor(handler.UploadPartHandler))
	http.HandleFunc("/file/mpupload/complete",
		handler.HTTPInterceptor(handler.CompleteUploadHandler))

	http.HandleFunc("/user/signup", handler.SignupHandler)
	http.HandleFunc("/user/signin", handler.SigninHandler)
	http.HandleFunc("/user/info", handler.HTTPInterceptor(handler.UserInfoHandler))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Failed to start server, err: %s", err.Error())
	}
}
