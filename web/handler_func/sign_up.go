package handler_func

import (
	"awesomeProject2/web/context"
	"awesomeProject2/web/session/manager"
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"net/http"
)

type User struct {
	Name     string `json:"Name"`
	Password string `json:"Password"`
}

type UserNoSecret struct {
	Name string
}

//func (u *UserNoSecret) MarshalBinary() ([]byte, error) {
//	var buf bytes.Buffer
//
//	nameLength := uint8(len(u.Name))
//	if err := binary.Write(&buf, binary.BigEndian, nameLength); err != nil {
//		return nil, err
//	}
//	if _, err := buf.WriteString(u.Name); err != nil {
//		return nil, err
//	}
//	return buf.Bytes(), nil
//}
//
//func (u *UserNoSecret) UnmarshalBinary(data []byte) error {
//	r := bytes.NewReader(data)
//	var nameLength uint8
//	if err := binary.Read(r, binary.BigEndian, &nameLength); err != nil {
//		return err
//	}
//	nameBytes := make([]byte, nameLength)
//	if _, err := io.ReadFull(r, nameBytes); err != nil {
//		return err
//	}
//	u.Name = string(nameBytes)
//	return nil
//}

var zlyUser = &User{
	Name:     "zly",
	Password: "123",
}

func SignUp(c *context.Context) {
	u := &User{}
	u.Name = c.R.FormValue("Name")
	u.Password = c.R.FormValue("Password")

	if u.Name == zlyUser.Name && u.Password == zlyUser.Password {
		id := uuid.New() // 可以使用更复杂的加密方式
		sess, err := manager.WebManager.InitSession(c, id.String())
		if err != nil {
			er := c.SystemErrorJson(err)
			if er != nil {
				log.Fatal("system error: ", er)
			}
		}
		userNoSecret := &UserNoSecret{
			Name: u.Name,
		}
		var byteVal []byte
		byteVal, err = json.Marshal(userNoSecret)
		if err != nil {
			er := c.SystemErrorJson(err)
			if er != nil {
				log.Fatal("system error: ", er)
			}
		}
		err = sess.Set(c, id.String(), byteVal)
		if err != nil {
			er := c.SystemErrorJson(err)
			if er != nil {
				log.Fatal("system error: ", er)
			}
		}
		http.Redirect(c.W, c.R, "/", http.StatusSeeOther)
	} else {
		err := c.UnauthorizedJsonDirect(c.R.URL.Path)
		if err != nil {
			er := c.SystemErrorJson(err)
			if er != nil {
				log.Fatal("system error: ", er)
			}
		}
	}
}
