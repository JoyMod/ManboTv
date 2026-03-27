package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

type BrowseBootstrapHandler struct {
	logger     *zap.Logger
	httpClient *http.Client
}

func NewBrowseBootstrapHandler(logger *zap.Logger) *BrowseBootstrapHandler {
	return &BrowseBootstrapHandler{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (h *BrowseBootstrapHandler) GetBootstrap(c *gin.Context) {
	payload := buildBrowsePayload(
		c.Request.Context(),
		h.httpClient,
		h.logger,
		parseBrowseQuery(c.DefaultQuery("kind", browseDefaultKind), c.Request.URL.Query()),
	)

	c.JSON(http.StatusOK, model.Success(payload))
}

func (h *BrowseBootstrapHandler) GetBootstrapLegacy(c *gin.Context) {
	payload := buildBrowsePayload(
		c.Request.Context(),
		h.httpClient,
		h.logger,
		parseBrowseQuery(c.DefaultQuery("kind", browseDefaultKind), c.Request.URL.Query()),
	)

	c.JSON(http.StatusOK, payload)
}
