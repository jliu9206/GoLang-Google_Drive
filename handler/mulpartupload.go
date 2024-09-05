package handler

import (
	rPool "GoLang-GoogleDrive/cache/redis"
	"GoLang-GoogleDrive/model"
	"GoLang-GoogleDrive/util"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int
	UploadID   string
	ChunkSize  int
	ChunkCount int
}

// InitializeMultipartUploadHandler: initialize multipart upload
func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. parse form
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "file size invalid", nil).JSONBytes())
		return
	}
	// 2. redis connection
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3. generate multipart info
	uploadInfo := MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   GenerateUploadID(username),
		ChunkSize:  5 * 1024 * 1024,
		ChunkCount: GenerateChunksCount(filesize),
	}
	// 4. write info into redis
	rConn.Do("HSET", "MP_"+uploadInfo.UploadID, "chunkcount", uploadInfo.ChunkCount,
		"filehash", uploadInfo.FileHash, "filesize", uploadInfo.FileSize, "chunksize", uploadInfo.ChunkSize)
	// 5. return res to client
	w.Write(util.NewRespMsg(0, "Init Multipart Upload OK", uploadInfo).JSONBytes())
}

// UploadPartHandler: upload single part of file
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	// 1. parse form
	r.ParseForm()
	// username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("chunkindex")

	// 2. create connection from redis
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 2.1 check if uploadId already has this chunk idx
	value, err := redis.String(rConn.Do("HGET", "MP_"+uploadID, "chkidx_"+chunkIndex))
	if err == redis.ErrNil {
		fmt.Printf("Chunk %s is not there yet", chunkIndex)
	} else if err != nil {
		fmt.Printf("Failed to hget redis: %v\n", err)
		w.Write(util.NewRespMsg(-1, "Failed to hget redis", nil).JSONBytes())
		return
	} else {
		if value == "1" {
			fmt.Printf("Chunk has been already uploaded")
			w.Write(util.NewRespMsg(0, "Chunk has been already uploaded", nil).JSONBytes())
			return
		}
	}

	// 3. get file chunk index for storage
	fileDir := fmt.Sprintf("/tmp/data/%s", uploadID)
	filePath := fmt.Sprintf("/tmp/data/%s/%s", uploadID, chunkIndex)

	err = os.MkdirAll(fileDir, 0744)
	if err != nil {
		fmt.Printf("Failed to create directory: %v\n", err)
		w.Write(util.NewRespMsg(-1, "Failed to create upload directory", nil).JSONBytes())
		return
	}

	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Failed to create file: %v\n", err)
		w.Write(util.NewRespMsg(-1, "Upload Part Failed", nil).JSONBytes())
		return
	}
	defer file.Close()

	// req.Body => buffer => os
	// 1. buffer can decrease the call for io&os
	// 2. batch handle
	// 3. stream handle
	// 4. large file, read at once take long time
	buffer := make([]byte, 1024*1024) // 1MB
	for {
		n, err := r.Body.Read(buffer)
		if n > 0 {
			if _, wErr := file.Write(buffer[:n]); wErr != nil {
				w.Write(util.NewRespMsg(-1, "Error writing to file", nil).JSONBytes())
				return
			}
			fmt.Printf("Chunk %s uploaded successfully\n", chunkIndex)
		}
		if err != nil {
			break
		}
	}

	// 4. update / complete status in redis
	rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)

	// 5. return res to client
	w.Write(util.NewRespMsg(0, "Part Upload ok", nil).JSONBytes())
}

// CompletUploadHandler: notice upload & merge
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. parse form
	r.ParseForm()
	uploadID := r.Form.Get("uploadid")
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize := r.Form.Get("filesize")
	filename := r.Form.Get("filename")
	// 2. connection from redis
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()
	// 3. see if all parts are uploaded by upload id (check mpupload status)
	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+uploadID))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Complete Upload Failed", nil).JSONBytes())
	}
	totalCount := 0
	chunkCount := 0
	for i := 0; i < len(data); i += 2 {
		key := string(data[i].([]byte))
		value := string(data[i+1].([]byte))
		if key == "chunkcount" {
			totalCount, _ = strconv.Atoi(value)
		} else if strings.HasPrefix(key, "chkidx_") && value == "1" {
			chunkCount++
			fmt.Printf("Total chunks: %d, Uploaded chunks: %d\n", totalCount, chunkCount)
		}
	}
	// see if all chunks/parts are uploaded ok
	if totalCount != chunkCount {
		w.Write(util.NewRespMsg(-1, "Not all parts are uploaded", nil).JSONBytes())
		return
	}
	// 4. merge all parts
	// 5. update tbl_file & tbl_user_file
	fsize, _ := strconv.Atoi(filesize)
	model.OnFileUploadFinished(filehash, filename, int64(fsize), "")
	model.OnUserFileUploadFinished(username, filehash, filename, int64(fsize))
	// 6 return res to client
	w.Write(util.NewRespMsg(0, "Complete upload ok", nil).JSONBytes())
}
func GenerateUploadID(username string) string {
	return username + fmt.Sprintf("%x", time.Now().UnixNano())
}

func GenerateChunksCount(filesize int) int {
	return int(math.Ceil(float64(filesize) / (5 * 1024 * 1024)))
}
