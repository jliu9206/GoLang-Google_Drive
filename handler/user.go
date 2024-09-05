package handler

import (
	"GoLang-GoogleDrive/model"
	"GoLang-GoogleDrive/util"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	PWDSALT   = "9*2/b"
	TOKENSALT = "_#tokens"
)

// SignupHandler: Create user with user name & password
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// return static html
		file, err := os.Open("./static/view/signup.html")
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
	} else if r.Method == http.MethodPost {
		r.ParseForm()

		username := r.Form.Get("username")
		password := r.Form.Get("password")

		if len(username) < 3 || len(password) < 5 {
			http.Error(w, "Internal Server Error", http.StatusBadRequest)
			return
		}

		enc_password := util.Sha1([]byte(password + PWDSALT))

		if suc := model.UserSignUp(username, enc_password); suc {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "Internal Server Error", http.StatusBadRequest)
			return
		}
	}
}

// SigninHandler: Sign in user & generate token if success
func SigninHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	encPassword := util.Sha1([]byte(password + PWDSALT))
	fmt.Print(encPassword)
	// user name & pwd verification
	pwdChecked := model.UserSignIn(username, encPassword)
	if !pwdChecked {
		http.Error(w, "Failed to authenticate (pwd)", http.StatusBadRequest)
		return
	}
	// token
	token := GenerateToken(username)
	updateOk := model.UpdateToken(username, token)
	if !updateOk {
		http.Error(w, "Failed to authenticate (token)", http.StatusBadRequest)
		return
	}
	// if ok, reroute to home
	// username, token, redirect location
	resp := util.RespMsg{
		Code: 0,
		Msg:  "Sign In OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp.JSONBytes())
}

func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// parse form
	r.ParseForm()
	username := r.Form.Get("username")
	token := r.Form.Get("token")
	// validate token
	if isTokenValid := IsTokenValid(username, token); !isTokenValid {
		http.Error(w, "Forbidden", http.StatusForbidden)
		fmt.Println("Not valid token")
		return
	}
	// query user info from db
	user, err := model.GetUserInfo(username)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		fmt.Println("user info not fetched")
		return
	}
	// return json data
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	w.Write((resp.JSONBytes()))
}

func GenerateToken(username string) string {
	timeStamp := fmt.Sprintf("%x", time.Now().Unix())

	tokenPrefix := util.MD5([]byte(username + TOKENSALT + timeStamp))

	return tokenPrefix + timeStamp[len(timeStamp)-8:]
}

func IsTokenValid(username string, token string) bool {
	if len(token) != 40 {
		return false
	}
	timeStampHex := token[len(token)-8:]
	timeStamp, err := strconv.ParseInt(timeStampHex, 16, 64)
	if err != nil {
		return false
	}

	if time.Now().Unix()-timeStamp > 3600 {
		return false
	}

	dbToken, err := model.GetTokenFromDB(username)
	if err != nil {
		fmt.Println("error getting from db")
		return false
	}
	return dbToken == token
}
