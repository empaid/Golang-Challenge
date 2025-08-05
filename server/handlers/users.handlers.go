package handlers

import (
	"encoding/json"
	"main/queries"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type GetUsersResponse struct {
	Users []UserResponse `json:"users"`
}

type UserResponse struct {
	ID           int     `json:"id"`
	Username     string  `json:"username"`
	Email        string  `json:"email"`
	UserType     string  `json:"userType"`
	Nickname     *string `json:"nickname,omitempty"`
	MessageCount int     `json:"messageCount,omitempty"`
}

type CreateUserRequest struct {
	Username string  `json:"username"`
	Email    string  `json:"email"`
	UserType string  `json:"userType"`
	Nickname *string `json:"nickname,omitempty"`
	Password string  `json:"password"`
}

type PatchUserRequest struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`
	UserType *string `json:"userType"`
	Nickname *string `json:"nickname,omitempty"`
}
type UserMessageResponse struct {
	ID        int    `json:"id"`
	UserId    int    `json:"userId"`
	Message   string `json:"message"`
	CreatedAt string `json:"createdAt"`
}

type CreateUserResponse struct {
	User UserResponse `json:"user"`
}

type CreateUserMessageResponse struct {
	UserMessage UserMessageResponse `json:"userMessage"`
}

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserLoginResponse struct {
	Token string `json:"token"`
}

type UserHandler struct {
	UserDB queries.UserDBStore
}

func (h *UserHandler) GetUsers(c *gin.Context) {

	userRows, err := h.UserDB.GetUsers(c)

	if err != nil {
		c.JSON(400, gin.H{
			"error": "Failed to retrieve users: " + err.Error(),
		})
		return
	}

	var users []UserResponse = []UserResponse{}
	for _, row := range userRows {

		var nickname *string = nil
		if row.Nickname.Valid {
			nickname = &row.Nickname.String
		}

		users = append(users, UserResponse{
			ID:           row.ID,
			Username:     row.Username,
			Email:        row.Email,
			UserType:     row.UserType,
			Nickname:     nickname,
			MessageCount: row.MessageCount,
		})
	}

	c.JSON(http.StatusOK, GetUsersResponse{
		Users: users,
	})
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var user CreateUserRequest
	if err := c.ShouldBindBodyWithJSON(&user); err != nil {
		c.JSON(400, gin.H{
			"error": "Bad Request: " + err.Error(),
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Bad Request: " + err.Error(),
		})
		return
	}

	userRow, err := h.UserDB.CreateUser(c, user.Username, user.Email, user.UserType, user.Nickname, hashedPassword)

	if err != nil {
		c.JSON(400, gin.H{
			"error": "Failed to create user: " + err.Error(),
		})
		return
	}

	var nickname *string = nil
	if userRow.Nickname.Valid {
		nickname = &userRow.Nickname.String
	}
	c.JSON(http.StatusOK, CreateUserResponse{
		User: UserResponse{
			ID:       userRow.ID,
			Username: userRow.Username,
			Email:    userRow.Email,
			UserType: userRow.UserType,
			Nickname: nickname,
		},
	})
}

func (h *UserHandler) PatchUser(c *gin.Context) {

	rawUserId, exists := c.Get("UserId")
	if !exists {
		c.JSON(500, gin.H{
			"error": "Missing userId ",
		})
		return
	}
	authUserId := rawUserId.(int)

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Should provide valid user id: " + err.Error(),
		})
		return
	}

	if authUserId != userID {
		c.JSON(401, gin.H{
			"error": "Not Authorized ",
		})
		return
	}

	var user PatchUserRequest
	if err := c.ShouldBindBodyWithJSON(&user); err != nil {
		c.JSON(400, gin.H{
			"error": "Bad Request: " + err.Error(),
		})
		return
	}

	var raw map[string]json.RawMessage
	if err := c.ShouldBindBodyWithJSON(&raw); err != nil {
		c.JSON(400, gin.H{"error": "invalid JSON: " + err.Error()})
		return
	}
	nicknameProvided := false
	if _, has := raw["nickname"]; has {
		nicknameProvided = true
	}

	userRow, err := h.UserDB.PatchUser(c, userID, user.Username, user.Email, user.UserType, user.Nickname, nicknameProvided)

	if err != nil {
		c.JSON(400, gin.H{
			"error": "Failed to update user: " + err.Error(),
		})
		return
	}

	var nickname *string = nil
	if userRow.Nickname.Valid {
		nickname = &userRow.Nickname.String
	}
	c.JSON(http.StatusOK, CreateUserResponse{
		User: UserResponse{
			ID:       userRow.ID,
			Username: userRow.Username,
			Email:    userRow.Email,
			UserType: userRow.UserType,
			Nickname: nickname,
		},
	})
}

func (h *UserHandler) CreateUserMessage(c *gin.Context) {
	rawUserId, exists := c.Get("UserId")
	if !exists {
		c.JSON(500, gin.H{
			"error": "Missing userId ",
		})
		return
	}
	authUserId := rawUserId.(int)

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Should provide valid user id: " + err.Error(),
		})
		return
	}

	if authUserId != userID {
		c.JSON(401, gin.H{
			"error": "Not Authorized ",
		})
		return
	}

	var userMessage UserMessageResponse
	if err := c.ShouldBindBodyWithJSON(&userMessage); err != nil {
		c.JSON(400, gin.H{
			"error": "Bad Request: " + err.Error(),
		})
		return
	}

	userMessageRow, err := h.UserDB.CreateUserMessage(c, userID, userMessage.Message)

	if err != nil {
		c.JSON(400, gin.H{
			"error": "Failed to create user message: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, CreateUserMessageResponse{
		UserMessage: UserMessageResponse{
			ID:        userMessageRow.ID,
			UserId:    userMessageRow.UserId,
			Message:   userMessageRow.Message,
			CreatedAt: userMessageRow.CreatedAt.Format(time.RFC3339),
		},
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var body UserLoginRequest
	if err := c.ShouldBindBodyWithJSON(&body); err != nil {
		c.JSON(400, gin.H{
			"error": "Bad Request: " + err.Error(),
		})
		return
	}
	userRow, err := h.UserDB.FetchUserFromEmail(c, body.Email)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Unable to find email",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword(userRow.Password, []byte(body.Password))
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Incorrect Password",
		})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userType": userRow.UserType,
		"id":       userRow.ID,
	})

	authToken, err := token.SignedString([]byte(os.Getenv("JWT_SIGNING_SECRET_KEY")))

	if err != nil {
		c.JSON(400, gin.H{
			"error": "Error while logging in" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UserLoginResponse{
		Token: authToken,
	})
}
