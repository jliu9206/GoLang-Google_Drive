package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"GoLang-GoogleDrive/meta"
	"GoLang-GoogleDrive/model"
	myStore "GoLang-GoogleDrive/store/ceph"
	"GoLang-GoogleDrive/util"
)

// // global variable for db
// var db *sql.DB

// func SetDB(database *sql.DB) {
// 	db = database
// }

// UploadHandler handles file upload: get -> html static page / post -> upload file stream
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// return static html
		file, err := os.Open("./static/view/upload.html")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Copy file to response
		_, err = io.Copy(w, file)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else if r.Method == "POST" {
		// read file stream
		file, head, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		fileMeta := meta.FileMeta{
			FileName:     head.Filename,
			FileLocation: "/tmp/" + head.Filename,
			UploadAt:     time.Now().Format("2006-01-02 15:04:05"),
		}
		// store file in dir
		// 1. create file
		// 2. copy file
		localFile, err := os.Create("/tmp/" + head.Filename)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer localFile.Close()

		fileMeta.FileSize, err = io.Copy(localFile, file)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		localFile.Seek(0, 0)
		// add localFile to ceph

		fileMeta.FileSha1 = util.FileSha1(localFile)
		localFile.Seek(0, 0)

		//TODO: upload file to ceph
		// userFileBucket := "userfile"
		// if isBucketExist, _ := myStore.BucketExists(userFileBucket); !isBucketExist{
		// 	myStore.CreateBucket(userFileBucket)
		// }
		cephPath := "/ceph/" + fileMeta.FileSha1
		err = myStore.UploadFileSync("userfile", cephPath, localFile)
		if err != nil {
			http.Error(w, "Internal Server Error: update user file table fails", http.StatusInternalServerError)
			return
		}
		fileMeta.FileLocation = cephPath

		// meta.UpdateFileMeta(fileMeta)
		_ = meta.UpdateFileMetaDB(fileMeta)

		// updte user file table
		r.ParseForm()
		username := r.Form.Get("username")
		if suc := model.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize); suc {
			// redirect
			http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
		} else {
			http.Error(w, "Internal Server Error: update user file table failed", http.StatusInternalServerError)
		}

	}
}

// UploadSucHandler: return message
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload ok!")
}

// GetFileMetaHandler: return file meta data by sha1
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fileHash := r.Form["filehash"][0]
	// fMeta := meta.GetFileMeta(fileHash)
	fMeta, err := meta.GetFileMetaDB(fileHash)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(fMeta)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")

	userFiles, err := model.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	// sortBy := r.Form.Get("sortBy")
	// fileMetas := meta.QueryFileMetas(limitCnt, sortBy)
	data, err := json.Marshal(userFiles)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// DownloadHandler: download file by sha1
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fsha1 := r.Form.Get("filehash")
	fmeta := meta.GetFileMeta(fsha1) //get file meta data

	file, err := os.Open(fmeta.FileLocation)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fmeta.FileName+"\"")
	w.Write(data)
}

// FileMetaUpdateHandler: Only support rename, 0
func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")
	if opType != "0" {
		http.Error(w, "Status Forbidden", http.StatusForbidden)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	currFileMeta := meta.GetFileMeta(fileSha1)
	currFileMeta.FileName = newFileName
	meta.UpdateFileMeta(currFileMeta)

	data, err := json.Marshal(currFileMeta)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// FileDeleteHandler, delete file by sha1
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fileSha1 := r.Form.Get("filehash")
	fMeta := meta.GetFileMeta(fileSha1)
	os.Remove(fMeta.FileLocation)

	meta.RemoveFileMeta(fileSha1)

	w.WriteHeader(http.StatusOK)
}

// TryFastUploadHandler: try fast upload
// 1. allow different users to upload same file at the same time
// 2. first in first in db
// 3. later upload only check db & update user-file table & delete uploaded file
// (3) asignc pipeline / timeout
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. parse form / param
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))
	// 2. query file table & check hash existence
	fileMeta, err := model.GetFileMeta(filehash)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// 3. i fnot exist, return fail
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "Fast upload fails, please use normal upload",
		}
		w.Write(resp.JSONBytes())
		return
	}
	// 4. if exists, add entry to user-file table, return ok
	if suc := model.OnUserFileUploadFinished(username, filehash, filename, int64(filesize)); suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "Fast upload ok",
		}
		w.Write(resp.JSONBytes())
		return
	} else {
		resp := util.RespMsg{
			Code: -2,
			Msg:  "Fast upload fails, please retry",
		}
		w.Write(resp.JSONBytes())
	}
}
