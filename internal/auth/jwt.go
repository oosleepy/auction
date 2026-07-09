package auth

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaim struct{
	Userid int
	jwt.RegisteredClaims
}

type Tokenheader struct{
	Token string `json:"token"`
}


func CreateAcessToken(userid int) (string,error) {

	claims := CustomClaim{
		Userid: userid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour*24)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenstring,err := token.SignedString([]byte(os.Getenv("JWT_KEY")))
	
	return tokenstring,err

}

func VerifyAccessToken(tokenstring string, claims *CustomClaim) error {
	_,err:= jwt.ParseWithClaims(tokenstring , claims, func(token *jwt.Token) (interface{} , error ){
		return []byte(os.Getenv("JWT_KEY")),nil
	})
	if err!=nil{
		return err
	}
	return nil
}

func VerifyMiddleware( next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		authheader := r.Header.Get("Authorization")
		parts := strings.Split(authheader, " ")
		token := parts[1]

		claims := &CustomClaim{}
		err := VerifyAccessToken(token, claims)

		if err!=nil{
			http.Error(w,"Unauthorized",http.StatusUnauthorized)
			return 
		}

		ctx := context.WithValue(r.Context(), "userid", claims.Userid)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	}
}
