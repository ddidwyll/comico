package cnst

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"log"
	"fmt"
)

type Config struct {
	Host   string `json:"host"`
	Port   string `json:"port"`
	Secret string `json:"secret"`
}

func GetConfig() Config {
	var config = Config{Host: "", Port: "8080", Secret: "lol"}
	file, err := os.Open("./config.json")
	defer file.Close()
	if err == nil {
		log.Println(json.NewDecoder(file).Decode(&config))
	}
	return config
}

const EXPIRE = 45 * 24 * 60 * 60

const STATIC = "./client/public/"

const (
	All = iota
	USER
	PASS
	GOOD
	POST
	CMNT
	LOGS
	MTMS
	FILE
	ERRR
)

const (
	UPD = iota
	INS
	DEL
	NTUPD
	NTINS
	NTDEL
	EXIST
	NTFND
	EMPTY
	DENY
	WRNG
	ACTVT
)

const (
	UNKN = iota
	MDRT
	ADMN
)

var maps = [][]string{
	{
		"Record updated",
		"User updated",
		"Password updated",
		"Good updated",
		"Post updated",
		"Comment updated",
	},
	{
		"Record created",
		"User created",
		"User created",
		"Good created",
		"Post created",
		"Comment created",
	},
	{
		"Record deleted",
		"User deleted",
		"Password deleted",
		"Good deleted",
		"Post deleted",
		"Comment deleted",
	},
	{
		"Record not updated.",
		"User not updated",
		"Password not updated",
		"Good not updated",
		"Post not updated",
		"Comment not updated",
	},
	{
		"Record not created",
		"User not created",
		"User not created",
		"Good not created",
		"Post not created",
		"Comment not created",
	},
	{
		"Record not deleted",
		"User not deleted",
		"Password not deleted",
		"Good not deleted",
		"Post not deleted",
		"Comment not deleted",
	},
	{
		"Entry already exists",
		"Login busy, try other",
		"Login busy, try other",
		"Good already exists",
		"Post already exists",
		"Comment already exists",
	},
	{
		"Record not exist",
		"User not exist or blocked",
		"Incorrect login or password",
		"Good not exist",
		"Post not exist",
		"Comment not exist",
	},
	{
		"Fill all required fields",
		"Fill all required fields",
		"Fill all required fields",
		"Fill all required fields",
		"Fill all required fields",
		"Fill all required fields",
	},
	{
		"This action denied for you",
		"This action denied for you",
		"This action denied for you",
		"This action denied for you",
		"This action denied for you",
		"This action denied for you",
	},
	{
		"Wrong data, contact with support",
		"Wrong data, contact with support",
		"Wrong data, contact with support",
		"Wrong data, contact with support",
		"Wrong data, contact with support",
		"Wrong data, contact with support",
	},
}

const INCLOGIN = "Incorrect login or password"

func Wait(sec string) string {
  return fmt.Sprintf("Too many tries, wait ~ %s sec", sec)
}

func Message(status, model byte) string {
	return maps[status][model]
}

func Error(status, model byte) error {
	return errors.New(maps[status][model])
}

func CommentType(db byte) (t string) {
	switch db {
	case GOOD:
		t = "goods"
	case POST:
		t = "posts"
	case USER:
		t = "users"
	}
	return
}

func CommentTypeReverse(t string) (db byte) {
	switch t {
	case "goods":
		db = GOOD
	case "posts":
		db = POST
	case "users":
		db = USER
	default:
		db = ERRR
	}
	return
}

func Status(err error, status, model byte) (int, []byte) {
	code := 200
	arr := []string{Message(status, model)}
	if err != nil {
		code = 400
		arr = append(arr, err.Error())
	}
	str := "{\"message\":\"" + strings.Join(arr, "; ") + "\"}"
	return code, []byte(str)
}
