package main

import (
	"GoLang-GoogleDrive/store/ceph"
	"fmt"
)

func main() {

	_, err := ceph.DownloadFileSync("userfile", "/ceph/7fa4f8faae8bf10785d61e9111bf9fccfad9c504", "test_file")
	if err != nil {
		fmt.Printf("ERROR FROM TESTING: %v", err)
		return
	}
	fmt.Println("OK")

	// accessKey := "L4ZATJRF4G1YX559NBDE"
	// secretKey := "OFSX0li3BmMXAP4ZyBURF5Dga58iLPzBnZCsUOXC"
	// endpoint := "http://127.0.0.1:7480"
	// _, err := ceph.NewS3Client(accessKey, secretKey, endpoint)

	// if err != nil {
	// 	fmt.Printf("Failed to initialize s3 client for ceph: %v", err)
	// }

	// fmt.Println("OK")
	// name := "userfile"

	// exists, err := ceph.BucketExists(name)
	// if err != nil {
	// 	return
	// }
	// if !exists {
	// 	fmt.Println("OK 1")
	// }
	// err = ceph.CreateBucket(name)
	// if err != nil {
	// 	return
	// }
	// if err == nil {
	// 	fmt.Println("OK 2")
	// }

	// exists, err = ceph.BucketExists(name)
	// if exists && err == nil {
	// 	fmt.Println("OK 3")
	// }
}
