package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"tasks-websocket/internal/config"
	"tasks-websocket/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func Authenticate(bd *sql.DB, c *gin.Context) {
	var userRequest models.User

	// get form data
	userRequest.Email = c.PostForm("email")
	userRequest.Password = c.PostForm("password")

	// if err := c.BindJSON(&userRequest); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	// 	return
	// }
	userRepo := models.NewUserRepository(bd)
	userReturned, err := userRepo.GetUserByEmail(userRequest.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Email or password is incorrect"})
		return
	}
	bcryptErr := bcrypt.CompareHashAndPassword([]byte(userReturned.Password), []byte(userRequest.Password))
	if bcryptErr != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "Email or password is incorrect"})
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    userReturned.ID,
		"email": userReturned.Email,
		"nbf":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour * 1).Unix(),
	})
	tokenString, err := token.SignedString([]byte(config.Cfg.TokenSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}
	c.SetCookie("token", tokenString, 3600, "/", "localhost", false, true)
	c.Redirect(302, "/")

}
func UpdateCardStatus(c *gin.Context, bd *sql.DB) {
	cardRepo := models.NewCardRepository(bd)
	type Status struct {
		Status string `json:"status"`
	}
	status := Status{}
	cardId := c.Param("id")
	err := c.BindJSON(&status)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	fmt.Println("CardId", cardId, "Status", status.Status)
	err = cardRepo.UpdateCardStatus(cardId, status.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating card status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Card status updated"})
}
func RenderIndexPage(c *gin.Context, bd *sql.DB) {
	userRepo := models.NewUserRepository(bd)
	boardsRepo := models.NewBoardRepository(bd)

	user, err := userRepo.GetUserById(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		log.Println(err)
		return
	}
	boards, err := boardsRepo.GetBoardsByUserId(user.ID)
	if err != nil {
		slog.Default().Error(err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "Boards not found"})
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{"Boards": boards})
}
func RenderBoardPage(c *gin.Context, bd *sql.DB) {
	cardRepo := models.NewCardRepository(bd)
	boardRepo := models.NewBoardRepository(bd)
	fmt.Println("Board ID: ", c.Param("id"))
	cards, err := cardRepo.GetCardsByBoardId(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	board, err := boardRepo.GetBoardByID(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
		return
	}
	c.HTML(http.StatusOK, "board.html", gin.H{"Cards": cards, "Board": board})
}

func RenderLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}
