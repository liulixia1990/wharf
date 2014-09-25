package models

import (
	"fmt"
	"reflect"
	"regexp"
	"time"

	"github.com/dockercn/docker-bucket/utils"
)

type User struct {
	Username      string //
	Password      string //
	Repositories  string //
	Organizations string //
	Email         string //Email 可以更换，全局唯一
	Fullname      string //
	Company       string //
	Location      string //
	Mobile        string //
	URL           string //
	Gravatar      string //如果是邮件地址使用 gravatar.org 的 API 显示头像，如果是上传的用户显示头像的地址。
	Actived       bool   //
	Created       int64  //
	Updated       int64  //
}

func (user *User) Has(username string) (bool, error) {
	if exist, err := LedisDB.Exists([]byte(GetObjectKey("user", username))); err != nil || exist == 0 {
		return false, err
	} else {
		return true, nil
	}
}

func (user *User) Add(username string, passwd string, email string, actived bool) error {
	//检查用户的 Key 是否存在
	if has, err := user.Has(username); err != nil {
		return err
	} else if has == true {
		//已经存在用户
		return fmt.Errorf("已经 %s 存在用户", username)
	} else {
		//检查用户名合法性，参考实现标准：
		//https://github.com/docker/docker/blob/28f09f06326848f4117baf633ec9fc542108f051/registry/registry.go#L27
		validNamespace := regexp.MustCompile(`^([a-z0-9_]{4,30})$`)
		if !validNamespace.MatchString(username) {
			return fmt.Errorf("用户名必须是 4 - 30 位之间，且只能由 a-z，0-9 和 下划线组成")
		}

		//检查密码合法性
		if len(passwd) < 5 {
			return fmt.Errorf("密码必须等于或大于 5 位字符以上")
		}

		//检查邮箱合法性
		validEmail := regexp.MustCompile(`^[a-z0-9A-Z]+([\-_\.][a-z0-9A-Z]+)*@([a-z0-9A-Z]+(-[a-z0-9A-Z]+)*\.)+[a-zA-Z]+$`)
		if !validEmail.MatchString(email) {
			return fmt.Errorf("Email 格式不合法")
		}

		key := utils.GeneralKey(username)

		user.Username = username
		user.Password = passwd
		user.Email = email
		user.Actived = actived

		user.Updated = time.Now().Unix()
		user.Created = time.Now().Unix()

		if err := user.Save(key); err != nil {
			return err
		} else {
			LedisDB.Set([]byte(GetObjectKey("user", username)), key)
		}

		return nil
	}
}

func (user *User) Save(key []byte) error {
	s := reflect.TypeOf(user).Elem()

	//循环处理 Struct 的每一个 Field
	for i := 0; i < s.NumField(); i++ {
		//获取 Field 的 Value
		value := reflect.ValueOf(user).Elem().Field(s.Field(i).Index[0])

		//判断 Field 不为空
		if utils.IsEmptyValue(value) == false {
			switch value.Kind() {
			case reflect.String:
				if _, err := LedisDB.HSet(key, []byte(s.Field(i).Name), []byte(value.String())); err != nil {
					return err
				}
			case reflect.Bool:
				if _, err := LedisDB.HSet(key, []byte(s.Field(i).Name), utils.BoolToBytes(value.Bool())); err != nil {
					return err
				}
			case reflect.Int64:
				if _, err := LedisDB.HSet(key, []byte(s.Field(i).Name), utils.Int64ToBytes(value.Int())); err != nil {
					return err
				}
			default:
				return fmt.Errorf("不支持的数据类型 %s", value.Kind().String())
			}
		}

	}

	return nil
}

func (user *User) Get(username string, passwd string, actived bool) (bool, error) {
	//检查用户的 Key 是否存在
	if has, err := user.Has(username); err != nil {
		return false, err
	} else if has == true {
		var key []byte

		//获取用户对象的 Key
		if key, err = LedisDB.Get([]byte(GetObjectKey("user", username))); err != nil {
			return false, err
		}

		//读取密码和Actived的值进行判断是否存在用户
		if results, err := LedisDB.HMget(key, []byte("Password"), []byte("Actived")); err != nil {
			return false, err
		} else {
			if password := results[0]; string(password) != passwd {
				return false, nil
			}

			if active := results[1]; utils.BytesToBool(active) != actived {
				return false, nil
			}

			return true, nil
		}

	} else {
		//没有用户的 Key 存在
		return false, nil
	}
}

type Organization struct {
	Owner        string    //用户的 Key，每个组织都由用户创建，Owner 默认是拥有所有 Repository 的读写权限
	Name         string    //
	Repositories string    //
	Privileges   string    //
	Users        string    //
	Actived      bool      //组织创建后就是默认激活的
	Created      time.Time //
	Updated      time.Time //
}

func (org *Organization) Has(name string) (bool, error) {
	if org, err := LedisDB.Exists([]byte(GetObjectKey("org", name))); err != nil {
		return false, err
	} else if org > 0 {
		return true, nil
	}

	return false, nil
}

func (org *Organization) Get(name string, actived bool) (bool, error) {
	return true, nil
}
