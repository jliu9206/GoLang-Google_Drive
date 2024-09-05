package model

import (
	"database/sql"
	"fmt"
)

// UserSignUp: Sign up by username & password
func UserSignUp(username string, password string) bool {
	// prepare sql query
	stmt, err := db.Prepare("INSERT IGNORE INTO tbl_user (`user_name`, `user_pwd`) VALUES (?,?)")
	if err != nil {
		fmt.Println("Failed to prepare statement, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(username, password)
	if err != nil {
		fmt.Println("Failed to execute statement, err:" + err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); err == nil && rf > 0 {
		return true
	}

	return false
}

// UserSignin: Query from db and see if pwd & name matches
func UserSignIn(username string, encpwd string) bool {
	stmt, err := db.Prepare("SELECT user_pwd FROM tbl_user WHERE user_name = ? LIMIT 1")
	if err != nil {
		fmt.Println("Failed to prepare statement, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	var userPwd string
	err = stmt.QueryRow(username).Scan(&userPwd)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("User not found")
		} else {
			fmt.Println("Failed to query statement, err:" + err.Error())
		}
		return false
	}

	return userPwd == encpwd
}

// UpdateToken: Insert / replace existing token for a user
func UpdateToken(username string, token string) bool {
	stmt, err := db.Prepare(
		"REPLACE INTO tbl_user_token (`user_name`, `user_token`) values (?,?)")

	if err != nil {
		fmt.Println("Failed to prepare statement, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, token)
	if err != nil {
		fmt.Println("Failed to exec statement, err:" + err.Error())
		return false
	}
	return true
}

type User struct {
	Username     string
	Email        string
	Phone        *string
	SignupAt     string
	LastActiveAt string
	Status       int
}

// GetUserInfo: Return user info
func GetUserInfo(username string) (User, error) {
	user := User{}
	stmt, err := db.Prepare("SELECT user_name, email, phone, signup_at, last_active, status FROM tbl_user WHERE user_name = ? LIMIT 1")
	if err != nil {
		fmt.Println("Failed to prepare statement, err:" + err.Error())
		return user, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&user.Username, &user.Email, &user.Phone, &user.SignupAt, &user.LastActiveAt, &user.Status)

	if err != nil {
		return user, err
	}
	return user, nil
}

func GetTokenFromDB(username string) (string, error) {
	var token string
	stmt, err := db.Prepare("SELECT user_token FROM tbl_user_token WHERE user_name = ? LIMIT 1")
	if err != nil {
		fmt.Println("Failed to prepare statement, err:" + err.Error())
		return token, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&token)
	if err != nil {
		return token, err
	}
	return token, nil
}
