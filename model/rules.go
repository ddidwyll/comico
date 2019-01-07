package model

import (
	"comico/cnst"
	"comico/db"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"strings"
	"time"
)

// MTIME

func (t *MTime) Get() {
	t.Users, _ = db.ReadOne(cnst.LOGS, fmt.Sprint(cnst.USER))
	t.Goods, _ = db.ReadOne(cnst.LOGS, fmt.Sprint(cnst.GOOD))
	t.Posts, _ = db.ReadOne(cnst.LOGS, fmt.Sprint(cnst.POST))
	t.Comments, _ = db.ReadOne(cnst.LOGS, fmt.Sprint(cnst.CMNT))
	t.Files, _ = db.ReadOne(cnst.LOGS, fmt.Sprint(cnst.FILE))
}

// USERACTION

func (ua *UserAction) Init(ip, user, action string) *UserAction {
	ua.IP = ip
	ua.User = user
	ua.Action = action
	return ua
}

func (ua *UserAction) key() string {
	return fmt.Sprint(ua.User, ":", ua.Action, ":", ua.IP)
}

func (ua *UserAction) now() int64 {
	return time.Now().Unix()
}

func (ua *UserAction) set(sec int64) {
	db.Insert(cnst.LOGS, ua.key(), fmt.Sprint(ua.now()+sec))
}

func (ua *UserAction) get() int64 {
	str, _ := db.ReadOne(cnst.LOGS, ua.key())
	lastAction, _ := strconv.ParseInt(str, 10, 64)
	return lastAction
}

func (ua *UserAction) IsWait(sec int64) (isWait bool, wait string) {
	delta := ua.get() - ua.now()
	if delta > 0 {
		sec += 2 * delta
		isWait = true
	}
	ua.set(sec)
	return isWait, fmt.Sprint(sec)
}

// HELPERS

func length(str string) int {
	return len([]rune(str))
}

func wrongId(id string, t byte) bool {
	if db.Exist(t, id) {
		return false
	}
	now := time.Now().Unix()
	idUnix, err := strconv.ParseInt(id[0:10], 10, 64)
	delta := now - idUnix
	return err != nil || delta < -72000 || delta > 72000
}

func wrongNick(nick string) bool {
	return length(nick) > 15 || length(nick) < 4 ||
		strings.ContainsAny(nick, "@ :;.#?&$\"'\\*[]{}()/")
}

func denied(user, author string) bool {
	u := new(User)
	return !u.get(user) || ((u.Id != author) && u.Status < cnst.MDRT)
	return false
}

/* USER */

func (u *User) id() string {
	return u.Id
}

func (u *User) get(id string) bool {
	g, err := GetOne(cnst.USER, id)
	n, ok := g.(*User)
	*u = *n
	if err != nil || !ok {
		return false
	}
	return true
}

func (u *User) validate(user string) error {
	if wrongNick(u.Id) || length(u.Title) > 25 || length(u.Type) > 25 {
		return cnst.Error(cnst.WRNG, cnst.USER)
	}
	i := 0
	for key, val := range u.Table {
		if length(key) > 25 || length(val) > 120 || val == "" || i > 9 {
			return cnst.Error(cnst.WRNG, cnst.USER)
		}
		i++
	}
	for i, val := range u.Scribes {
		u.Scribes[i] = strings.ToLower(val)
		if length(val) > 25 {
			return cnst.Error(cnst.WRNG, cnst.USER)
		}
	}
	for i, val := range u.Ignores {
		u.Ignores[i] = strings.ToLower(val)
		if length(val) > 25 {
			return cnst.Error(cnst.WRNG, cnst.USER)
		}
	}
	c := new(User)
	if !c.get(user) || ((c.Id != u.Id) && c.Status < cnst.MDRT) ||
		u.Status > c.Status {
		return cnst.Error(cnst.DENY, cnst.USER)
	}
	if !db.Exist(cnst.PASS, u.Id) {
		str := "{\"id\":\"" + u.Id + "\",\"pass\":\"NA\"}"
		db.Insert(cnst.PASS, u.Id, str)
	}
	return nil
}

func Activity(user string) {
	u := new(User)
	if u.get(user) {
		u.Activity = time.Now().Unix()
		str, _ := json.Marshal(&u)
		db.Insert(cnst.USER, u.Id, string(str))
	}
}

func Tag(user, tag string, ignore bool) {
	u := new(User)
	tag = strings.ToLower(tag)
	var (
		exist  bool
		arr    []string
		target []string
	)
	if u.get(user) {
		if ignore {
			target = u.Ignores
		} else {
			target = u.Scribes
		}
		for _, val := range target {
			if val == tag {
				exist = true
			} else if val != "" {
				arr = append(arr, val)
			}
		}
		if !exist {
			arr = append(arr, tag)
		}
		if !ignore {
			u.Scribes = arr
		} else {
			u.Ignores = arr
		}
		str, _ := json.Marshal(&u)
		db.Insert(cnst.USER, u.Id, string(str))
	}
}

func (u *User) pattern(user string) string {
	return "*"
}

/* PASSWORD */

func (p *Password) id() string {
	return p.Id
}

func (p *Password) get(id string) bool {
	g, err := GetOne(cnst.PASS, id)
	n, ok := g.(*Password)
	*p = *n
	if err != nil || !ok {
		return false
	}
	return true
}

func (p *Password) validate(user string) error {
	if wrongNick(p.Id) || length(p.Pass) > 15 ||
		length(p.Pass) < 4 {
		return cnst.Error(cnst.WRNG, cnst.USER)
	}
	if p.Id == "" || p.Pass == "" {
		return cnst.Error(cnst.EMPTY, cnst.PASS)
	}
	c, u := new(User), new(User)
	if !c.get(user) && !u.get(p.Id) {
		str := "{\"id\":\"" + p.Id + "\",\"title\":\"\"}"
		db.Insert(cnst.USER, p.Id, str)
	} else if !u.get(p.Id) {
		return cnst.Error(cnst.NTFND, cnst.USER)
	}
	if (u.Id != c.Id) &&
		(c.Status != cnst.ADMN) {
		return cnst.Error(cnst.DENY, cnst.PASS)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(p.Pass), 10)
	p.Pass = string(hash)
	return err
}

func (p *Password) Login() (*User, error) {
	u := new(User)
	if p.Id == "" || p.Pass == "" {
		return u, cnst.Error(cnst.EMPTY, cnst.PASS)
	}
	pass := []byte(p.Pass)
	if !p.get(p.Id) || !u.get(p.Id) {
		return u, cnst.Error(cnst.NTFND, cnst.USER)
	}
	err := bcrypt.CompareHashAndPassword([]byte(p.Pass), pass)
	if err != nil {
		return u, cnst.Error(cnst.NTFND, cnst.PASS)
	}
	return u, nil
}

func (p *Password) pattern(user string) string {
	return "nothing"
}

/* GOOD */

func (p *Good) id() string {
	return p.Id
}

func (g *Good) get(id string) bool {
	x, err := GetOne(cnst.GOOD, id)
	n, ok := x.(*Good)
	*g = *n
	if err != nil || !ok {
		return false
	}
	return true
}

func (g *Good) validate(user string) error {
	if wrongId(g.Id, cnst.GOOD) || length(g.Title) > 35 ||
		length(g.Price) > 25 || length(g.Type) > 25 || length(g.Text) > 1100 {
		return cnst.Error(cnst.WRNG, cnst.GOOD)
	}
	i := 0
	for key, val := range g.Table {
		if length(key) > 25 || length(val) > 120 || val == "" || i > 9 {
			return cnst.Error(cnst.WRNG, cnst.GOOD)
		}
		i++
	}
	for i, val := range g.Images {
		if length(val) > 100 || i > 9 {
			return cnst.Error(cnst.WRNG, cnst.GOOD)
		}
	}
	if g.Title == "" || g.Price == "" ||
		g.Type == "" || g.Text == "" {
		return cnst.Error(cnst.EMPTY, cnst.GOOD)
	}
	if denied(user, g.Author) {
		return cnst.Error(cnst.DENY, cnst.GOOD)
	}
	return nil
}

func (g *Good) pattern(user string) string {
	return "*"
}

/* POST */

func (p *Post) id() string {
	return p.Id
}

func (p *Post) get(id string) bool {
	g, err := GetOne(cnst.POST, id)
	n, ok := g.(*Post)
	*p = *n
	if err != nil || !ok {
		return false
	}
	return true
}

func (p *Post) validate(user string) error {
	if wrongId(p.Id, cnst.POST) || length(p.Title) > 35 ||
		length(p.Type) > 25 || length(p.Text) > 1100 {
		return cnst.Error(cnst.WRNG, cnst.POST)
	}
	if p.Title == "" || p.Type == "" || p.Text == "" {
		return cnst.Error(cnst.EMPTY, cnst.POST)
	}
	if denied(user, p.Author) {
		return cnst.Error(cnst.DENY, cnst.POST)
	}
	return nil
}

func (p *Post) pattern(user string) string {
	return "*"
}

/* COMMENT */

func (c *Comment) id() string {
	return c.Owner + ":" + c.Type + ":" + c.Id
}

func (c *Comment) get(id string) bool {
	g, err := GetOne(cnst.CMNT, id)
	n, ok := g.(*Comment)
	*c = *n
	if err != nil || !ok {
		return false
	}
	return true
}

func (c *Comment) validate(user string) error {
	t := cnst.CommentTypeReverse(c.Type)
	if (wrongId(c.Id, cnst.CMNT) && !db.Exist(cnst.CMNT, c.id())) ||
		t == cnst.ERRR || !db.Exist(t, c.Owner) ||
		length(c.To) > 15 || length(c.Text) > 600 {
		return cnst.Error(cnst.WRNG, cnst.CMNT)
	}
	if c.Text == "" {
		return cnst.Error(cnst.EMPTY, cnst.CMNT)
	}
	if denied(user, c.Author) {
		return cnst.Error(cnst.DENY, cnst.CMNT)
	}
	return nil
}

func (c *Comment) pattern(user string) string {
	return "*"
}
