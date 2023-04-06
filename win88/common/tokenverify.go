package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

const ACCESS_SECRET = "aaasskjafjas;kjd;fklaj;kfjall"

func CreateToken(username, password, packagetag string, snid int32) (string, error) {
	var err error
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["username"] = username
	atClaims["password"] = password
	atClaims["packagetag"] = packagetag
	atClaims["snid"] = snid
	atClaims["exp"] = time.Now().Add(time.Hour * 1).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(ACCESS_SECRET))
	if err != nil {
		return "", err
	}
	return token, nil
}

func VerifyToken(tokenString string) (bool, error, string, string, string, int32) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		//return hmacSampleSecret, nil
		return []byte(ACCESS_SECRET), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println(fmt.Sprintf("%v,%v,%v,%v", claims["authorized"], claims["username"], claims["password"], claims["packagetag"], claims["exp"]))
		return true, nil, claims["username"].(string), claims["password"].(string), claims["packagetag"].(string), claims["snid"].(int32)
	} else {
		fmt.Println(err)
		return false, err, "", "", "", 0
	}
}

type TokenUserData struct {
	TelegramId string
	Password   string
	Packagetag string
	Expired    int64
}

func CreateTokenAes(tu *TokenUserData) (string, error) {
	to, err := json.Marshal(tu)
	if err != nil {
		return "", err
	}
	token := EnCrypt(to)
	return token, nil
}

func VerifyTokenAes(tokenString string) (error, *TokenUserData) {
	str := DeCrypt(tokenString)
	var tu TokenUserData
	err := json.Unmarshal([]byte(str), &tu)
	if err != nil {
		return err, nil
	}
	if tu.Expired < time.Now().Unix() {
		return errors.New("Token Expired"), nil
	}
	return err, &tu
}
