package model

import (
	"database/sql"
	"fmt"
)

type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

// global variable
var db *sql.DB

func SetDB(database *sql.DB) {
	db = database
}

// OnFileUploadFinished: if insert ok, true, else false
func OnFileUploadFinished(filehash string, filename string, filesize int64, fileaddr string) bool {
	// prepare sql query
	stmt, err := db.Prepare("INSERT IGNORE INTO tbl_file (file_sha1, file_name, file_size, file_addr, status) VALUES (?,?,?,?,1)")
	if err != nil {
		fmt.Println("Failed to prepare statement, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	//exec sql query
	ret, err := stmt.Exec(filehash, filename, filesize, fileaddr)
	if err != nil {
		fmt.Println("Failed to execute statement, err:" + err.Error())
		return false
	}

	rf, err := ret.RowsAffected()
	if err != nil {
		fmt.Println("Failed to retrieve affected rows, err:" + err.Error())
		return false
	}

	if rf <= 0 {
		fmt.Println("Failed to affect rows")
		return false
	}

	fmt.Println("File uploaded and record inserted!!")
	return true
}

func GetFileMeta(filehash string) (*TableFile, error) {
	stmt, err := db.Prepare("SELECT file_sha1, file_addr, file_name, file_size " +
		"FROM tbl_file WHERE file_sha1=? AND status=1 LIMIT 1")
	if err != nil {
		fmt.Println("Failed to prepare sql statement")
		return nil, err
	}
	defer stmt.Close()

	// create an instance to store result
	tfile := TableFile{}

	// execute & store
	err = stmt.QueryRow(filehash).Scan(&tfile.FileHash, &tfile.FileAddr, &tfile.FileName, &tfile.FileSize)
	if err != nil {
		fmt.Println("Failed to query")
		return nil, err
	}

	return &tfile, nil
}
