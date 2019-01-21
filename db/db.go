package db

import (
	"github.com/ddidwyll/comico/cnst"
	"github.com/tidwall/buntdb"

	"fmt"
	"log"
	"strings"
	"time"
)

var expire = &buntdb.SetOptions{
	Expires: true,
	TTL:     time.Second * cnst.EXPIRE,
}

var (
	Users, usersErr         = buntdb.Open("users.db")
	Passwords, passwordsErr = buntdb.Open("passwords.db")
	Goods, goodsErr         = buntdb.Open("goods.db")
	Posts, postsErr         = buntdb.Open("posts.db")
	Comments, commentsErr   = buntdb.Open("comments.db")
	Files, filesErr         = buntdb.Open("files.db")
	Logs, _                 = buntdb.Open(":memory:")
)

func shooseDB(name byte) *buntdb.DB {
	switch name {
	case cnst.USER:
		return Users
	case cnst.PASS:
		return Passwords
	case cnst.GOOD:
		return Goods
	case cnst.POST:
		return Posts
	case cnst.CMNT:
		return Comments
	case cnst.LOGS:
		return Logs
	case cnst.FILE:
		return Files
	}
	return nil
}

func Exist(db byte, key string) bool {
	if str, _ := ReadOne(db, key); str == "" {
		return false
	}
	return true
}

func lastUpdate(db byte) {
	if db != cnst.LOGS {
		now := time.Now().Unix()
		Insert(cnst.LOGS, fmt.Sprint(db), fmt.Sprint(now))
	}
}

func Insert(db byte, key, val string) (err error) {
	shooseDB(db).Update(
		func(tx *buntdb.Tx) error {
			_, _, err = tx.Set(key, val, expire)
			return err
		})
	lastUpdate(db)
	return err
}

func ReadOne(db byte, key string) (str string, err error) {
	shooseDB(db).View(
		func(tx *buntdb.Tx) error {
			str, err = tx.Get(key)
			return err
		})
	return str, err
}

func ReadAll(db byte, pattern string) (str string, err error) {
	var arr []string
	monthAgo := fmt.Sprint(time.Now().Unix() - cnst.EXPIRE)
	err = shooseDB(db).View(
		func(tx *buntdb.Tx) error {
			return tx.DescendKeys(pattern,
				func(key, value string) bool {
					if (db == cnst.CMNT && strings.Contains(key, ":users:")) ||
						(db == cnst.FILE && strings.Contains(key, ":users:")) ||
						db == cnst.USER || key[:10] > monthAgo {
						arr = append(arr, value)
						return true
					}
					return false
				})
		})
	str = "[" + strings.Join(arr, ",") + "]"
	return str, err
}

func Delete(db byte, key string) error {
	err := shooseDB(db).Update(
		func(tx *buntdb.Tx) error {
			_, err := tx.Delete(key)
			return err
		})
	if err != nil {
		return err
	}
	lastUpdate(db)
	if db != cnst.CMNT && db != cnst.FILE {
		cmntType := cnst.CommentType(db)
		delComments(key, cmntType, "", false)
	}
	return err
}

func getComments(id, t string) (cmnts []string) {
	Comments.View(
		func(tx *buntdb.Tx) error {
			return tx.DescendKeys(id+":"+t+":*",
				func(key, value string) bool {
					cmnts = append(cmnts, key)
					return true
				})
		})
	return
}

func delComments(id, t, newId string, renew bool) {
	keys := getComments(id, t)
	for _, key := range keys {
		Comments.Update(
			func(tx *buntdb.Tx) error {
				val, _ := tx.Delete(key)
				if renew {
					val = strings.Replace(val, id, newId, 1)
					tx.Set(newId+":"+key[11:], val, expire)
				}
				return nil
			})
	}
	lastUpdate(cnst.CMNT)
}

func Renew(db byte, key string) (error, string) {
	newId := fmt.Sprint(time.Now().Unix())
	err := shooseDB(db).Update(
		func(tx *buntdb.Tx) error {
			val, err := tx.Delete(key)
			if err != nil {
				return err
			}
			val = strings.Replace(val, key, newId, 1)
			_, _, err = tx.Set(newId, val, expire)
			return err
		})
	if err != nil {
		return err, ""
	}
	lastUpdate(db)
	cmntType := cnst.CommentType(db)
	delComments(key, cmntType, newId, true)
	return err, newId
}

func Start() {
	if usersErr != nil {
		log.Fatal(usersErr)
	}
	if passwordsErr != nil {
		log.Fatal(passwordsErr)
	}
	if goodsErr != nil {
		log.Fatal(goodsErr)
	}
	if postsErr != nil {
		log.Fatal(postsErr)
	}
	if commentsErr != nil {
		log.Fatal(commentsErr)
	}
	if filesErr != nil {
		log.Fatal(filesErr)
	}
	Users.Shrink()
	Passwords.Shrink()
	Goods.Shrink()
	Posts.Shrink()
	Comments.Shrink()
	Files.Shrink()
	lastUpdate(cnst.USER)
	lastUpdate(cnst.GOOD)
	lastUpdate(cnst.POST)
	lastUpdate(cnst.CMNT)
	lastUpdate(cnst.FILE)
}

func Stop() {
	Users.Close()
	Passwords.Close()
	Goods.Close()
	Posts.Close()
	Comments.Close()
	Logs.Close()
	Files.Close()
}
