package model

import (
	"comico/cnst"
	"comico/db"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"strings"
	"time"
)

// FILES

func GetFiles() string {
	str, _ := db.ReadAll(cnst.FILE, "*")
	return str
}

func setFile(id, t string) error {
	key := id + ":" + t
	now := fmt.Sprint(time.Now().Unix())
	return db.Insert(cnst.FILE, key, "\""+key+":"+now+"\"")
}

func delFile(t byte, id string) {
	if t == cnst.GOOD {
		name := cnst.STATIC + "img/goods_" + id
		os.Remove(name)
		os.Remove(name + ".jpg")
		os.Remove(name + "_sm.jpg")
	}
	if t == cnst.USER {
		name := cnst.STATIC + "img/users_" + id
		os.Remove(name + "_sm.jpg")
	}
	if t == cnst.USER || t == cnst.GOOD {
		fileKey := id + ":" + cnst.CommentType(t)
		db.Delete(cnst.FILE, fileKey)
	}
}

func renewFile(t byte, id, newId string) {
	if t == cnst.GOOD {
		name := cnst.STATIC + "img/goods_" + id
		newName := cnst.STATIC + "img/goods_" + newId
		os.Rename(name, newName)
		os.Rename(name+".jpg", newName+".jpg")
		os.Rename(name+"_sm.jpg", newName+"_sm.jpg")
		oldKey := id + ":" + cnst.CommentType(t)
		oldVal, _ := db.ReadOne(cnst.FILE, oldKey)
		arr := strings.Split(oldVal, ":")
		if len(arr) == 3 {
			newKey := newId + ":" + cnst.CommentType(t)
			db.Delete(cnst.FILE, oldKey)
			db.Insert(cnst.FILE, newKey, "\""+newKey+":"+strings.Trim(arr[2], "\"")+"\"")
		}
	}
}

func Upload(t, n, uid string, file *multipart.FileHeader) error {
	db := cnst.CommentTypeReverse(t)
	if db != cnst.USER && db != cnst.GOOD {
		return cnst.Error(cnst.WRNG, db)
	}
	if err := Validate(db, n, uid); err != nil {
		return cnst.Error(cnst.WRNG, db)
	}
	name := cnst.STATIC + "img/" + t + "_" + n
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(name)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	if err = setFile(n, t); err != nil {
		return err
	}
	if db == cnst.GOOD {
		mogrify := exec.Command("convert", "-format", "JPEG", "-resize", "1200",
			"-compress", "JPEG", "-quality", "75", name, name+".jpg")
		if _, err = mogrify.Output(); err != nil {
			return err
		}
		mogrify = exec.Command("convert", "-resize", "550", name+".jpg", name+"_sm.jpg")
		if _, err = mogrify.Output(); err != nil {
			return err
		}
	} else {
		mogrify := exec.Command("convert", "-resize", "130x130", name, name)
		if _, err = mogrify.Output(); err != nil {
			return err
		}
		os.Rename(name, name+"_sm.jpg")
	}
	return nil
}
