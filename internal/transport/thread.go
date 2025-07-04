package transport

import (
	"net/http"
	"nik-mLb/forum_task.com/internal/models"
	"nik-mLb/forum_task.com/internal/usecase"
	"strconv"

	"github.com/labstack/echo"
)

type ThreadHandlers struct {
    uc usecase.IThreadUseCase
}

func NewThreadHandlers(uc usecase.IThreadUseCase) *ThreadHandlers {
    return &ThreadHandlers{uc: uc}
}

func (h *ThreadHandlers) NewThreadHandlers(e *echo.Echo) {
	e.POST( "/api/thread/:slug_or_id/create", h.CreatePosts)
	e.GET("/api/thread/:slug_or_id/details", h.GetThreadDetails)
    e.POST("/api/thread/:slug_or_id/details", h.UpdateThread)
	e.GET("/api/thread/:slug_or_id/posts", h.GetThreadPosts)
	e.POST("/api/thread/:slug_or_id/vote", h.VoteForThread)
}

func (h *ThreadHandlers) CreatePosts(c echo.Context) error {
    slugOrID := c.Param("slug_or_id")

    var newPosts models.NewPosts
    if err := c.Bind(&newPosts); err != nil {
        return c.JSON(http.StatusBadRequest, models.Error{Message: "Invalid request body"})
    }

    createdPosts, err := h.uc.CreatePosts(slugOrID, newPosts)
    if err != nil {
        switch err.Error() {
        case "thread not found":
            return c.JSON(http.StatusNotFound, models.Error{Message: "Thread not found"})
        case "parent post not found":
            return c.JSON(http.StatusConflict, models.Error{Message: "Parent post not found"})
        case "author not found":
            return c.JSON(http.StatusNotFound, models.Error{Message: "Author not found"})
        default:
            return c.JSON(http.StatusInternalServerError, models.Error{Message: err.Error()})
        }
    }

    return c.JSON(http.StatusCreated, createdPosts)
}

func (h *ThreadHandlers) GetThreadPosts(c echo.Context) error {
    slugOrID := c.Param("slug_or_id")
    
    limit, err := strconv.Atoi(c.QueryParam("limit"))
    if err != nil || limit <= 0 {
        limit = 100 
    }
    
    since, _ := strconv.Atoi(c.QueryParam("since"))
    sort := c.QueryParam("sort")
    if sort == "" {
        sort = "flat"
    }
    desc, _ := strconv.ParseBool(c.QueryParam("desc"))
    
    posts, err := h.uc.GetThreadPosts(slugOrID, limit, since, sort, desc)
    if err != nil {
        return c.JSON(http.StatusNotFound, models.Error{Message: "Thread not found"})
    }
    
    return c.JSON(http.StatusOK, posts)
}

func (h *ThreadHandlers) GetThreadDetails(c echo.Context) error {
    slugOrID := c.Param("slug_or_id")
    
    thread, err := h.uc.GetThreadDetails(slugOrID)
    if err != nil {
        return c.JSON(http.StatusNotFound, models.Error{Message: "Thread not found"})
    }
    
    return c.JSON(http.StatusOK, thread)
}

func (h *ThreadHandlers) UpdateThread(c echo.Context) error {
    slugOrID := c.Param("slug_or_id")
    
    var threadUpdate models.ThreadUpdate
    if err := c.Bind(&threadUpdate); err != nil {
        return c.JSON(http.StatusBadRequest, models.Error{Message: "Invalid request body"})
    }
    
    thread, err := h.uc.UpdateThread(slugOrID, threadUpdate)
    if err != nil {
        return c.JSON(http.StatusNotFound, models.Error{Message: "Thread not found"})
    }
    
    return c.JSON(http.StatusOK, thread)
}

func (h *ThreadHandlers) VoteForThread(c echo.Context) error {
    slugOrID := c.Param("slug_or_id")
    
    var vote models.Vote
    if err := c.Bind(&vote); err != nil {
        return c.JSON(http.StatusBadRequest, models.Error{Message: "Invalid request body"})
    }
    
    thread, err := h.uc.VoteForThread(slugOrID, vote)
    if err != nil {
        switch err.Error() {
        case "thread not found":
            return c.JSON(http.StatusNotFound, models.Error{Message: "Thread not found"})
		case "voice must be -1 or 1":
            return c.JSON(http.StatusNotFound, models.Error{Message: "Thread not found"})
        case "user not found":
            return c.JSON(http.StatusNotFound, models.Error{Message: "User not found"})
        default:
            return c.JSON(http.StatusInternalServerError, models.Error{Message: err.Error()})
        }
    }
    
    return c.JSON(http.StatusOK, thread)
}