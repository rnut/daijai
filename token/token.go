package token

import (
	"daijai/models"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(user models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid":      user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString([]byte(os.Getenv("SECRET")))

}

func TokenValid(c *gin.Context) error {
	var tokenString string
	if err := ExtractToken(c, &tokenString); err != nil {
		log.Println("tokenString")
		return fmt.Errorf("unexpected token: %v", err)
	}
	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET")), nil
	})
	if err != nil {
		return err
	}
	return nil
}

func ExtractToken(c *gin.Context, token *string) error {
	authHeader := c.Request.Header.Get("Authorization")
	parts := strings.Split(authHeader, " ")
	ps := len(parts)
	switch ps {
	case 0, 1:
		return fmt.Errorf("not enough Bearer")
	case 2:
		if parts[0] != "Bearer" {
			return fmt.Errorf("non Bearer")
		} else {
			*token += parts[1]
			return nil
		}
	default:
		return fmt.Errorf("too many Bearer")
	}
}

func ExtractTokenID(c *gin.Context) (uint, error) {
	var tokenString string
	if err := ExtractToken(c, &tokenString); err != nil {
		return 0, fmt.Errorf("unexpected token")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET")), nil
	})
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		data := claims["uid"]
		// log.Printf("ExtractToken: uid: %v", data)
		uid, err := strconv.ParseUint(fmt.Sprintf("%.0f", data), 10, 32)
		if err != nil {
			log.Printf("ExtractToken: err, uid: %v", err)
			return 0, err
		}
		return uint(uid), nil
	} else {
		fmt.Println(err)
		return 0, err
	}
}

// func ExampleParse_hmac() {
// 	// sample token string taken from the New example

// 	// Parse takes the token string and a function for looking up the key. The latter is especially
// 	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
// 	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
// 	// to the callback, providing flexibility.
// 	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 		// Don't forget to validate the alg is what you expect:
// 		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
// 		}

// 		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
// 		return hmacSampleSecret, nil
// 	})

// 	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
// 		fmt.Println(claims["foo"], claims["nbf"])
// 	} else {
// 		fmt.Println(err)
// 	}

// 	// Output: bar 1.4444784e+09
// }
