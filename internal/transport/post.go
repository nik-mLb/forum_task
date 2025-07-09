package transport

import (
	"net/http"
	"nik-mLb/forum_task.com/internal/models"
	"nik-mLb/forum_task.com/internal/usecase"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type PostHandlers struct {
    uc usecase.IPostUseCase
}

func NewPostHandlers(uc usecase.IPostUseCase) *PostHandlers {
    return &PostHandlers{uc: uc}
}

func (h *PostHandlers) NewPostHandlers(e *echo.Echo) {
	e.GET("/api/post/:id/details", h.GetPostDetails)
	e.POST("/api/post/:id/details", h.UpdatePost)
}

func (h *PostHandlers) GetPostDetails(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Error{Message: "Invalid post ID"})
	}

	related := strings.Split(c.QueryParam("related"), ",")
	
	postFull, err := h.uc.GetPostDetails(id, related)
	if err != nil {
		return c.JSON(http.StatusNotFound, models.Error{Message: "Post not found"})
	}

	return c.JSON(http.StatusOK, postFull)
}

func (h *PostHandlers) UpdatePost(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.Error{Message: "Invalid post ID"})
	}

	var postUpdate models.PostUpdate
	if err := c.Bind(&postUpdate); err != nil {
		return c.JSON(http.StatusBadRequest, models.Error{Message: "Invalid request body"})
	}

	post, err := h.uc.UpdatePost(id, postUpdate)
	if err != nil {
		return c.JSON(http.StatusNotFound, models.Error{Message: "Post not found"})
	}

	return c.JSON(http.StatusOK, post)
}