package handlers

import (
	"fmt"
	"tasks-websocket/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func EnsureAuthenticated() gin.HandlerFunc {
	return func(c *gin.Context) {
		stringToken, err := c.Cookie("token")
		if err != nil {
			c.Redirect(302, "/login")
		}
		token, err := jwt.Parse(stringToken, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// 	// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return []byte(config.Cfg.TokenSecret), nil
		})
		if err != nil {
			c.Redirect(302, "/login")
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["id"])
			c.Next()
		}

	}
}
