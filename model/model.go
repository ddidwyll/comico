package model

import (
	"github.com/ddidwyll/comico/cnst"
	"github.com/ddidwyll/comico/db"

	"encoding/json"
)

type User struct {
	Id       string            `json:"id"`
	Status   byte              `json:"status"`
	Title    string            `json:"title"`
	Type     string            `json:"type"`
	Table    map[string]string `json:"Table"`
	Activity int64             `json:"activity"`
	Scribes  []string          `json:"scribes"`
	Ignores  []string          `json:"ignores"`
}

type Password struct {
	Id   string `json:"id"`
	Pass string `json:"pass"`
}

type Good struct {
	Id     string            `json:"id"`
	Author string            `json:"auth"`
	Title  string            `json:"title"`
	Type   string            `json:"type"`
	Price  string            `json:"price"`
	Text   string            `json:"text"`
	Table  map[string]string `json:"Table"`
	Images []string          `json:"images"`
}

type Post struct {
	Id     string `json:"id"`
	Author string `json:"auth"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	Text   string `json:"text"`
}

type Comment struct {
	Id     string `json:"id"`
	Author string `json:"auth"`
	Owner  string `json:"owner"`
	To     string `json:"to"`
	Type   string `json:"type"`
	Text   string `json:"text"`
}

type Modeler interface {
	id() string
	get(string) bool
	validate(string) error
	pattern(string) string
}

type MTime struct {
	Users    string `json:"users"`
	Goods    string `json:"goods"`
	Posts    string `json:"posts"`
	Comments string `json:"cmnts"`
	Files    string `json:"files"`
}

type UserAction struct {
	IP     string
	User   string
	Action string
}

func Model(t byte) Modeler {
	switch t {
	case cnst.USER:
		return new(User)
	case cnst.PASS:
		return new(Password)
	case cnst.GOOD:
		return new(Good)
	case cnst.POST:
		return new(Post)
	case cnst.CMNT:
		return new(Comment)
	}
	return nil
}

func GetOne(t byte, id string) (Modeler, error) {
	obj := Model(t)
	str, err := db.ReadOne(t, id)
	if str == "" {
		return obj, cnst.Error(cnst.NTFND, t)
	}
	if err != nil {
		return obj, err
	}
	err = json.Unmarshal([]byte(str), obj)
	return obj, err
}

func Validate(t byte, id, user string) error {
	m := Model(t)
	if !m.get(id) {
		return cnst.Error(cnst.NTFND, t)
	}
	return m.validate(user)
}

func Upsert(t, action byte, m Modeler, user string) error {
	if !db.Exist(t, m.id()) && action == cnst.UPD {
		return cnst.Error(cnst.NTFND, t)
	}
	if db.Exist(t, m.id()) && action == cnst.INS {
		return cnst.Error(cnst.EXIST, t)
	}
	if err := m.validate(user); err != nil {
		return err
	}
	str, err := json.Marshal(&m)
	if err != nil {
		return err
	}
	if err := db.Insert(t, m.id(), string(str)); err != nil {
		return err
	}
	return nil
}

func Delete(t byte, id, user string, renew bool) error {
	m := Model(t)
	if !m.get(id) {
		return cnst.Error(cnst.NTFND, t)
	}
	if err := m.validate(user); err != nil {
		return err
	}
	if renew {
		err, newId := db.Renew(t, id)
		if err == nil {
			renewFile(t, id, newId)
		}
		return err
	}
	delFile(t, id)
	return db.Delete(t, id)
}

func GetAll(t byte, user string) (string, error) {
	m := Model(t)
	pattern := m.pattern(user)
	return db.ReadAll(t, pattern)
}

func Init() {
	if !db.Exist(cnst.USER, "admin") && !db.Exist(cnst.PASS, "admin") {
		pass := Password{Id: "admin", Pass: "$2a$10$LcsC2urIFeAATYZkGEFUS.ANDi0djtA65xkiB3CO5TKzjWwPRy5ru"}
		user := User{Id: "admin", Status: 2}
		str, _ := json.Marshal(&pass)
		db.Insert(cnst.PASS, "admin", string(str))
		str, _ = json.Marshal(&user)
		db.Insert(cnst.USER, "admin", string(str))
	}
}
