package admin

import (
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// ProbeChannel uses the account's saved endpoint, not any URL supplied to this request.
func (h *AccountHandler) ProbeChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	if h.accountTestService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Account test service unavailable")
		return
	}
	ctx := c.Request.Context()
	account, err := h.adminService.GetAccount(ctx, id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if account == nil {
		response.ErrorFrom(c, service.ErrAccountNotFound)
		return
	}
	response.Success(c, h.accountTestService.ProbeChannelConnection(ctx, account))
}
