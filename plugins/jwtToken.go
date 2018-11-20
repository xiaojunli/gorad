package plugins

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"errors"

	"strings"
)

type JWTTokenMiddleware struct {
	TokenName string
	IIS string
	Cert string
	IgnorePath string
	StaticPath string
}

/**
 * 执行
 */
func (self *JWTTokenMiddleware) Process(context *gin.Context, Engine *gin.Engine) {

}

func (self *JWTTokenMiddleware) Init(context *gin.Context, engine *gin.Engine) {
	engine.Use(func(context *gin.Context) {
		fmt.Println("----------------TOKEN-------------------")
		fmt.Println("------------->" + context.Request.URL.String())


		//解析表单
		jwtToken := context.Request.FormValue("token")

		if "" != jwtToken {
			//TODO
			fmt.Println("ooo - exec process token....")

			token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}

				//iss := config.MustString("jwt.claims.iis", "http://longim.net")

				checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(self.IgnorePath, false)
				if !checkIss {
					return token, errors.New("Invalid issuer.")
				}
				// 我们的公钥,可以在<a href="https://manage.auth0.com/" target="_blank">https://manage.auth0.com/</a> 上下载到对应的封装好的json，里面包括了签名
				k5c := "abac"
				cert := "-----BEGIN CERTIFICATE-----\n" + k5c + "\n-----END CERTIFICATE-----"

				result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))

				fmt.Println("---------------------------------")
				fmt.Println("result：", result)
				fmt.Println("---------------------------------")

				//result := []byte(cert) // 不是正确的 PUBKEY 格式 都会 报  key is of invalid type

				return result, nil
			})

			if nil != err {
				fmt.Println("ooo - validate token is fail, message: " + err.Error())
			} else {
				if !token.Valid {
					fmt.Println("-----> token is no valid!!!!")
				}
			}
		} else {
			respondWithError(401, "API token required", context)

			return
		}

		context.Next()
		fmt.Println("------------END TOKEN--------------")
	})
}

func (self *JWTTokenMiddleware) getIggnoreMatchList() (map[string][]string){
	resultMap := make(map[string][]string)

	//处理ignorePath
	if "" != self.IgnorePath {

	}

	//处理staticPath
	if "" != self.StaticPath {

	}


	return resultMap
}

/**
 * 分割字符串为数组
 */
func processPath(path string, splitChar string) ([]string) {
	var resultArray = make([]string, 0)

	if "" != path && strings.Contains(path, splitChar){
		resultArray = strings.Split(path, splitChar)
	}

	return resultArray
}

/**
 * 验证失败
 */
func respondWithError(code int, message string, context *gin.Context) {
	context.JSON(code, message)
}

