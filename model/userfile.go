package model

import (
	"fmt"
	"time"
)

// UserFile struct
type UserFile struct {
	UserName    string
	FileHash    string
	FileName    string
	FileSize    int64
	UploadAt    string
	LastUpdated string
}

// OnUserFileUploadFinished: update user file table after upload
func OnUserFileUploadFinished(username string, filehash string, filename string, filesize int64) bool {

	stmt, err := db.Prepare(
		"INSERT IGNORE INTO tbl_user_file (`user_name`, `file_sha1`, `file_name`, `file_size`, `upload_at`) " +
			"values (?,?,?,?,?)")
	if err != nil {
		return false
	}
	defer stmt.Close()

	if _, err = stmt.Exec(username, filehash, filename, filesize, time.Now()); err != nil {
		return false
	}
	return true
}

// QueryUserFileMetas: Retrieve all files uploaded by user
func QueryUserFileMetas(username string, limit int) ([]UserFile, error) {

	stmt, err := db.Prepare(
		"SELECT file_sha1, file_name, file_size, upload_at, last_update from " +
			"tbl_user_file where user_name=? limit ?")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(username, limit)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var userFileList []UserFile
	for rows.Next() {
		ufile := UserFile{}
		err = rows.Scan(&ufile.FileHash, &ufile.FileName, &ufile.FileSize,
			&ufile.UploadAt, &ufile.LastUpdated)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		userFileList = append(userFileList, ufile)
	}
	return userFileList, nil
}
