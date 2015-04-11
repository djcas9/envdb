package main

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"strings"
	"time"
)

type UserDb struct {
	Id int64

	Enabled bool

	Name  string
	Email string `xorm:"UNIQUE NOT NULL"`

	Password string `xorm:"NOT NULL"`
	Rands    string `xorm:"VARCHAR(10)"`
	Salt     string `xorm:"VARCHAR(10)"`

	Avatar string `xorm:"VARCHAR(2048) NOT NULL"`

	Admin bool

	Created time.Time `xorm:"CREATED"`
	Updated time.Time `xorm:"UPDATED"`
}

func (u *UserDb) AvatarLink() string {
	return "//1.gravatar.com/avatar/" + u.Avatar
}

func FindAllUsers() ([]*UserDb, error) {
	var users []*UserDb
	err := x.Find(&users)
	return users, err
}

func GetUserSalt(n int, alphabets ...byte) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		if len(alphabets) == 0 {
			bytes[i] = alphanum[b%byte(len(alphanum))]
		} else {
			bytes[i] = alphabets[b%byte(len(alphabets))]
		}
	}
	return string(bytes)
}

func hashEmail(email string) string {
	// https://en.gravatar.com/site/implement/hash/
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)

	h := md5.New()
	h.Write([]byte(email))
	return hex.EncodeToString(h.Sum(nil))
}

func IsEmailUsed(email string) (bool, error) {
	if len(email) == 0 {
		return false, nil
	}
	return x.Get(&UserDb{Email: email})
}

// CreateUser creates record of a new user.
func CreateUser(u *UserDb) error {
	isExist, err := IsEmailUsed(u.Email)

	if err != nil {
		return err
	} else if isExist {
		return errors.New("A user with this email already exists.")
	}

	u.Avatar = hashEmail(u.Email)
	u.Rands = GetUserSalt(10)
	u.Salt = GetUserSalt(10)
	u.EncodePassword()

	sess := x.NewSession()
	defer sess.Close()
	if err = sess.Begin(); err != nil {
		return err
	}

	if _, err = sess.Insert(u); err != nil {
		sess.Rollback()
		return err
	} else if err = sess.Commit(); err != nil {
		return err
	}

	// Auto-set admin for user whose ID is 1.
	if u.Id == 1 {
		u.Admin = true
		_, err = x.Id(u.Id).UseBool().Update(u)
	}
	return err
}

// EncodePasswd encodes password to safe format.
func (u *UserDb) EncodePassword() {
	newPasswd := PBKDF2([]byte(u.Password), []byte(u.Salt), 10000, 50, sha256.New)
	u.Password = fmt.Sprintf("%x", newPasswd)
}

// ValidtePassword checks if given password matches the one belongs to the user.
func (u *UserDb) ValidatePassword(passwd string) bool {
	newUser := &UserDb{Password: passwd, Salt: u.Salt}
	newUser.EncodePassword()
	return u.Password == newUser.Password
}

func (self *UserDb) Update() error {
	sess := x.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return err
	}

	if _, err := sess.Id(self.Id).AllCols().Update(self); err != nil {
		sess.Rollback()
		return err
	}

	err := sess.Commit()

	if err != nil {
		return err
	}

	return nil
}

// http://code.google.com/p/go/source/browse/pbkdf2/pbkdf2.go?repo=crypto
func PBKDF2(password, salt []byte, iter, keyLen int, h func() hash.Hash) []byte {
	prf := hmac.New(h, password)
	hashLen := prf.Size()
	numBlocks := (keyLen + hashLen - 1) / hashLen

	var buf [4]byte
	dk := make([]byte, 0, numBlocks*hashLen)
	U := make([]byte, hashLen)
	for block := 1; block <= numBlocks; block++ {
		// N.B.: || means concatenation, ^ means XOR
		// for each block T_i = U_1 ^ U_2 ^ ... ^ U_iter
		// U_1 = PRF(password, salt || uint(i))
		prf.Reset()
		prf.Write(salt)
		buf[0] = byte(block >> 24)
		buf[1] = byte(block >> 16)
		buf[2] = byte(block >> 8)
		buf[3] = byte(block)
		prf.Write(buf[:4])
		dk = prf.Sum(dk)
		T := dk[len(dk)-hashLen:]
		copy(U, T)

		// U_n = PRF(password, U_(n-1))
		for n := 2; n <= iter; n++ {
			prf.Reset()
			prf.Write(U)
			U = U[:0]
			U = prf.Sum(U)
			for x := range U {
				T[x] ^= U[x]
			}
		}
	}
	return dk[:keyLen]
}

func FindUserByEmail(email string) (*UserDb, error) {
	if len(email) == 0 {
		return nil, errors.New("User not found")
	}

	user := &UserDb{Email: strings.ToLower(email)}

	has, err := x.Get(user)

	if err != nil {
		return nil, err
	}

	if has {
		return user, nil
	}

	return nil, errors.New("User not found")
}

func (self *UserDb) Delete() error {
	_, err := x.Delete(self)

	if err != nil {
		return err
	}

	return nil
}
