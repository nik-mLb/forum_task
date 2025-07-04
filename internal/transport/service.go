package transport

import (
	"net/http"
	"nik-mLb/forum_task.com/internal/models"
	"nik-mLb/forum_task.com/internal/usecase"

	"github.com/labstack/echo"
)

type ServiceHandlers struct {
	uc usecase.IServiceUseCase
}

func NewServiceHandlers(uc usecase.IServiceUseCase) *ServiceHandlers {
    return &ServiceHandlers{uc: uc}
}

func (h *ServiceHandlers) NewServiceHandlers(e *echo.Echo) {
    e.GET( "/api/service/status", h.Status)

	e.POST( "/api/service/clear", h.Clear)
}

func (h *ServiceHandlers) Clear(c echo.Context) error {
	if err := h.uc.Clear(); err != nil {
		return c.JSON(http.StatusInternalServerError, models.Error{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, nil)
}

func (h *ServiceHandlers) Status(c echo.Context) error {
	status, err := h.uc.GetStatus()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Error{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, status)
}