package auth

import (
	"net/http"

	"github.com/ArtemSilin1/HotelCrm-HTTP/internal/http-server/logger"
	"github.com/ArtemSilin1/HotelCrm-HTTP/internal/storage/data"
	"github.com/ArtemSilin1/HotelCrm-HTTP/internal/storage/models/users"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

const moduleName = "AuthModule"

func NewHandler(db *pgxpool.Pool, log *logger.Logger) *HandlerAuth {
	return &HandlerAuth{
		db:     db,
		logger: log,
	}
}

type HandlerAuth struct {
	db     *pgxpool.Pool
	logger *logger.Logger
}

func (h *HandlerAuth) InitHandler(router *gin.Engine) {
	router.POST("/user/create-acc", h.Create)
	router.POST("/user/enter-acc", h.Login)
}

//type UserRequest struct {
//	Token users.JWTToken `json:"token"`
//	User  users.Users    `json:"users"`
//}

func (h *HandlerAuth) Create(c *gin.Context) {
	var request users.Users
	//
	//userRole, tokenIsValid := h.getUserRole(&request.Token)
	//
	//if !tokenIsValid || userRole != data.Admin_manager {
	//	c.JSON(http.StatusForbidden, gin.H{"error": "token is invalid"})
	//	return
	//}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"response": data.WrongData})
		return
	}

	_, err := request.CreateUser(h.db)
	if err != nil {
		logger.New("error", moduleName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"response": data.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": "done"})
}

func (h *HandlerAuth) Login(c *gin.Context) {
	var user users.Users

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"response": data.WrongData})
		return
	}

	token, err := user.LoginUser(h.db)
	if err != nil {
		logger.New("error", moduleName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"response": data.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": token})
}

func (h *HandlerAuth) getUserRole(token *users.JWTToken) (string, bool) {
	mapClaims, err := token.VerifyToken(token.Token)
	if err != nil {
		logger.New("error", moduleName, err)
		return "", false
	}

	role, ok := mapClaims["role"].(string)
	if !ok {
		return "", false
	}

	return role, true
}
