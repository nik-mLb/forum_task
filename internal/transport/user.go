package transport

import (
    "encoding/json"
    "net/http"
    "nik-mLb/forum_task.com/internal/models"
    "nik-mLb/forum_task.com/internal/usecase"

    "github.com/labstack/echo"
)

type UserHandlers struct {
    uc usecase.IUserUseCase
}

func NewUserHandlers(uc usecase.IUserUseCase) *UserHandlers {
    return &UserHandlers{uc: uc}
}

func (h *UserHandlers) NewUserHandlers(e *echo.Echo) {
    e.POST("/api/user/:nickname/create", h.CreateUser)
    e.GET("/api/user/:nickname/profile", h.GetUser)
    e.POST("/api/user/:nickname/profile", h.UpdateUser)
}

func (h *UserHandlers) CreateUser(c echo.Context) error {
    nickname := c.Param("nickname")
    var newUser models.NewUser
    if err := json.NewDecoder(c.Request().Body).Decode(&newUser); err != nil {
        return c.JSON(http.StatusBadRequest, models.Error{Message: "Invalid request body"})
    }

    user, err := h.uc.CreateUser(newUser, nickname)
    if err != nil {
        if err.Error() == "user already exists" {
            existingUsers, _ := h.uc.GetUserByNicknameOrEmail(nickname, newUser.Email)
            return c.JSON(http.StatusConflict, existingUsers)
        }
        return c.JSON(http.StatusInternalServerError, models.Error{Message: err.Error()})
    }

    return c.JSON(http.StatusCreated, user)
}

func (h *UserHandlers) GetUser(c echo.Context) error {
    nickname := c.Param("nickname")
    user, err := h.uc.GetUserByNickname(nickname)
    if err != nil {
        if err.Error() == "user not found" {
            return c.JSON(http.StatusNotFound, models.Error{Message: "User not found"})
        }
        return c.JSON(http.StatusInternalServerError, models.Error{Message: err.Error()})
    }

    return c.JSON(http.StatusOK, user)
}

func (h *UserHandlers) UpdateUser(c echo.Context) error {
    nickname := c.Param("nickname")
    var update models.UserUpdate
    if err := json.NewDecoder(c.Request().Body).Decode(&update); err != nil {
        return c.JSON(http.StatusBadRequest, models.Error{Message: "Invalid request body"})
    }

    user, err := h.uc.UpdateUser(nickname, update)
    if err != nil {
        switch err.Error() {
        case "user not found":
            return c.JSON(http.StatusNotFound, models.Error{Message: "User not found"})
        case "email already exists":
            return c.JSON(http.StatusConflict, models.Error{Message: "Email already exists"})
        default:
            return c.JSON(http.StatusInternalServerError, models.Error{Message: err.Error()})
        }
    }

    return c.JSON(http.StatusOK, user)
}