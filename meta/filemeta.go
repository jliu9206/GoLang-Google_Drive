package meta

import (
	"GoLang-GoogleDrive/model"
	"sort"
)

type FileMeta struct {
	FileSha1     string
	FileName     string
	FileLocation string
	FileSize     int64
	UploadAt     string
}

var fileMetaData map[string]FileMeta

func init() {
	fileMetaData = make(map[string]FileMeta)
}

// UpdateFileMeta: update/create a file meta data
func UpdateFileMeta(fmeta FileMeta) {
	fileMetaData[fmeta.FileSha1] = fmeta
}

// UpdateFileMetaDB: update/create
func UpdateFileMetaDB(fmeta FileMeta) bool {
	return model.OnFileUploadFinished(fmeta.FileSha1, fmeta.FileName, fmeta.FileSize, fmeta.FileLocation)

}

// QueryFileMetas: get multiple file metas, supports sorting by name, time, or size
func QueryFileMetas(limit int, sortBy string) []FileMeta {
	var fMetaArray []FileMeta
	for _, v := range fileMetaData {
		fMetaArray = append(fMetaArray, v)
	}
	switch sortBy {
	case "name":
		sort.Slice(fMetaArray, func(i, j int) bool {
			return fMetaArray[i].FileName < fMetaArray[j].FileName
		})
	case "size":
		sort.Slice(fMetaArray, func(i, j int) bool {
			return fMetaArray[i].FileSize < fMetaArray[j].FileSize
		})
	case "time":
		sort.Slice(fMetaArray, func(i, j int) bool {
			return fMetaArray[i].UploadAt < fMetaArray[j].UploadAt
		})
	}
	if limit > len(fMetaArray) {
		limit = len(fMetaArray)
	}
	return fMetaArray[:limit]
}

// GetFileMeta: get file meta data by sha1
func GetFileMeta(fsha1 string) FileMeta {
	return fileMetaData[fsha1]
}

// GetFileMetaDB: get file meta from DB by sha1
func GetFileMetaDB(filesha1 string) (FileMeta, error) {
	tfile, err := model.GetFileMeta(filesha1)
	if err != nil {
		return FileMeta{}, err
	}
	fmeta := FileMeta{
		FileSha1:     tfile.FileHash,
		FileName:     tfile.FileName.String,
		FileLocation: tfile.FileAddr.String,
		FileSize:     tfile.FileSize.Int64,
	}
	return fmeta, nil
}

// RemoveFileMeta: remove file by sha1
func RemoveFileMeta(fsha1 string) {
	delete(fileMetaData, fsha1)
}
