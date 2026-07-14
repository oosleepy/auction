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

//this is used to put the acces token inside the w basically to encode it into the w
type Tokenheader struct{
	Token string `json:"token"`
}


func CreateAcessToken(userid int) (string,error) {
	
	//basically claims is the part which like explains the payload of the jwt kinda 
	//so it has user id and the expires at is what is used here
	claims := CustomClaim{
		Userid: userid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour*24)),
		},
	}
	
	//creating the token kinda
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	//hashing it with the jwt key
	tokenstring,err := token.SignedString([]byte(os.Getenv("JWT_KEY")))
	
	//returning the tokenstring to be encoded and err if there was any
	return tokenstring,err

}

//this jsut checks if the accesstoekn is valid or not using the jwt key it decodes it and gives the user id if err it returns error 
// so after the veiffy it oupts the answer into the claims pointer in the arguments and fn returns the value of error if no error nil if err then err
func VerifyAccessToken(tokenstring string, claims *CustomClaim) error {
	_,err:= jwt.ParseWithClaims(tokenstring , claims, func(token *jwt.Token) (interface{} , error ){
		return []byte(os.Getenv("JWT_KEY")),nil
	})
	if err!=nil{
		return err
	}
	return nil
}

//this is the middleware layer 
//it checks the protected route and if its valid then take the user id and put it in the context so that the next function can use it with context 
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
		
		//next is basically protexted route 
		next.ServeHTTP(w, r)
	}
}
