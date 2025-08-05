package handlers

import (
	"main/queries"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GetUsersResponse struct {
	Users []UserResponse `json:"users"`
}

type UserResponse struct {
	ID       int     `json:"id"`
	Username string  `json:"username"`
	Email    string  `json:"email"`
	UserType string  `json:"userType"`
	Nickname *string `json:"nickname,omitempty"`
}

type CreateUserResponse struct {
	User UserResponse `json:"user"`
}

type UserHandler struct {
	UserDB *queries.UserDB
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
			ID:       row.ID,
			Username: row.Username,
			Email:    row.Email,
			UserType: row.UserType,
			Nickname: nickname,
		})
	}

	c.JSON(http.StatusOK, GetUsersResponse{
		Users: users,
	})
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var user UserResponse
	if err := c.ShouldBindBodyWithJSON(&user); err != nil {
		c.JSON(400, gin.H{
			"error": "Bad Request: " + err.Error(),
		})
		return
	}

	userRow, err := h.UserDB.CreateUser(c, user.Username, user.Email, user.UserType, user.Nickname)

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
			UserType: userRow.Username,
			Nickname: nickname,
		},
	})
}
