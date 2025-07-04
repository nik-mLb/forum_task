package transport

import (
	"encoding/json"
	"net/http"
	"nik-mLb/forum_task.com/internal/models"
	"nik-mLb/forum_task.com/internal/usecase"
	"strconv"

	"github.com/labstack/echo"
)

type ForumHandlers struct {
    uc usecase.IForumUseCase
}

func NewForumHandlers(uc usecase.IForumUseCase) *ForumHandlers {
    return &ForumHandlers{uc: uc}
}

func (h *ForumHandlers) NewForumHandlers(e *echo.Echo) {
    e.POST("/api/forum/create", h.CreateForum)
    e.GET("/api/forum/:slug/details", h.GetForum)
    e.POST("/api/forum/:slug/create", h.CreateThread)
    e.GET("/api/forum/:slug/users", h.GetForumUsers)
    e.GET("/api/forum/:slug/threads", h.GetForumThreads)
}

func (h *ForumHandlers) CreateForum(c echo.Context) error {
    var forum models.Forum
    if err := json.NewDecoder(c.Request().Body).Decode(&forum); err != nil {
        return c.JSON(http.StatusBadRequest, models.Error{Message: "Invalid request body"})
    }

    createdForum, err := h.uc.CreateForum(forum)
    if err != nil {
        switch err.Error() {
        case "user not found":
            return c.JSON(http.StatusNotFound, models.Error{Message: "User not found"})
        case "forum already exists":
            return c.JSON(http.StatusConflict, createdForum)
        default:
            return c.JSON(http.StatusInternalServerError, models.Error{Message: err.Error()})
        }
    }

    return c.JSON(http.StatusCreated, createdForum)
}

func (h *ForumHandlers) GetForum(c echo.Context) error {
    slug := c.Param("slug")
    forum, err := h.uc.GetForumBySlug(slug)
    if err != nil {
        return c.JSON(http.StatusNotFound, models.Error{Message: "Forum not found"})
    }
    return c.JSON(http.StatusOK, forum)
}

func (h *ForumHandlers) CreateThread(c echo.Context) error {
    slug := c.Param("slug")
    var thread models.Thread
    if err := json.NewDecoder(c.Request().Body).Decode(&thread); err != nil {
        return c.JSON(http.StatusBadRequest, models.Error{Message: "Invalid request body"})
    }

    createdThread, err := h.uc.CreateThread(slug, thread)
    if err != nil {
        switch err.Error() {
        case "forum not found", "user not found":
            return c.JSON(http.StatusNotFound, models.Error{Message: err.Error()})
        case "thread already exists":
            return c.JSON(http.StatusConflict, createdThread)
        default:
            return c.JSON(http.StatusInternalServerError, models.Error{Message: err.Error()})
        }
    }

    return c.JSON(http.StatusCreated, createdThread)
}

func (h *ForumHandlers) GetForumThreads(c echo.Context) error {
    slug := c.Param("slug")
    
    limit, err := strconv.Atoi(c.QueryParam("limit"))
    if err != nil || limit <= 0 {
        limit = 100
    }
    
    since := c.QueryParam("since")
    desc, _ := strconv.ParseBool(c.QueryParam("desc"))
    
    threads, err := h.uc.GetForumThreads(slug, limit, since, desc)
    if err != nil {
        return c.JSON(http.StatusNotFound, models.Error{Message: "Forum not found"})
    }
    
    return c.JSON(http.StatusOK, threads)
}

func (h *ForumHandlers) GetForumUsers(c echo.Context) error {
    slug := c.Param("slug")

    limit, err := strconv.Atoi(c.QueryParam("limit"))
    if err != nil || limit <= 0 {
        limit = 100
    }

    since := c.QueryParam("since")
    desc, _ := strconv.ParseBool(c.QueryParam("desc"))

    users, err := h.uc.GetForumUsers(slug, limit, since, desc)
    if err != nil {
        return c.JSON(http.StatusNotFound, models.Error{Message: err.Error()})
    }

    if len(users) == 0 {
        return c.JSON(http.StatusOK, []models.User{})
    }

    return c.JSON(http.StatusOK, users)
}