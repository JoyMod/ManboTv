package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

// CategoryActionRequest 分类操作请求
type CategoryActionRequest struct {
	Action string `json:"action" binding:"required"`
	Name   string `json:"name,omitempty"`
	Type   string `json:"type,omitempty"`
	Query  string `json:"query,omitempty"`
}

// SiteConfigRequest 站点配置请求
type SiteConfigRequest struct {
	SiteConfig model.SiteConfig `json:"site_config" binding:"required"`
}

// HandleCategory 分类管理
// POST /api/admin/category
func (h *AdminLegacyHandler) HandleCategory(c *gin.Context) {
	var req CategoryActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "请求参数无效"))
		return
	}

	ctx := c.Request.Context()
	categories, err := h.adminStorage.GetCustomCategories(ctx)
	if err != nil {
		h.logger.Error("获取分类失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "操作失败"))
		return
	}

	switch req.Action {
	case "add":
		if req.Name == "" || req.Type == "" || req.Query == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少必要参数"))
			return
		}
		for _, category := range categories {
			if category.Query == req.Query && category.Type == req.Type {
				c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "该分类已存在"))
				return
			}
		}

		categories = append(categories, model.CustomCategory{
			Name:     req.Name,
			Type:     req.Type,
			Query:    req.Query,
			Disabled: false,
			From:     "custom",
		})
		if err := h.adminStorage.SaveCustomCategories(ctx, categories); err != nil {
			h.logger.Error("保存分类失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "添加失败"))
			return
		}
		h.logger.Info("分类已添加", zap.String("name", req.Name))

	case "delete":
		if req.Query == "" || req.Type == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少 query 或 type 参数"))
			return
		}
		for index, category := range categories {
			if category.Query != req.Query || category.Type != req.Type {
				continue
			}
			if category.From == "config" {
				c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "不能删除配置文件中的分类"))
				return
			}

			categories = append(categories[:index], categories[index+1:]...)
			if err := h.adminStorage.SaveCustomCategories(ctx, categories); err != nil {
				h.logger.Error("保存分类失败", zap.Error(err))
				c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "删除失败"))
				return
			}
			h.logger.Info("分类已删除", zap.String("name", category.Name))
			c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
			return
		}

		c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "分类不存在"))
		return

	case "disable", "enable":
		if req.Query == "" || req.Type == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少 query 或 type 参数"))
			return
		}
		for index := range categories {
			if categories[index].Query != req.Query || categories[index].Type != req.Type {
				continue
			}
			categories[index].Disabled = req.Action == "disable"
			if err := h.adminStorage.SaveCustomCategories(ctx, categories); err != nil {
				h.logger.Error("保存分类失败", zap.Error(err))
				c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "操作失败"))
				return
			}
			h.logger.Info("分类状态已切换", zap.String("name", categories[index].Name))
			c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
			return
		}

		c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "分类不存在"))
		return

	default:
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "未知的操作类型"))
		return
	}

	c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
}

// HandleSite 站点配置
// POST /api/admin/site
func (h *AdminLegacyHandler) HandleSite(c *gin.Context) {
	var req SiteConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "请求参数无效"))
		return
	}

	config, err := h.adminStorage.GetAdminConfig(c.Request.Context())
	if err != nil {
		h.logger.Error("获取配置失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取配置失败"))
		return
	}

	config.SiteConfig = req.SiteConfig
	if err := h.adminStorage.SaveAdminConfig(c.Request.Context(), config); err != nil {
		h.logger.Error("保存配置失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "保存失败"))
		return
	}

	h.logger.Info("站点配置已更新")
	c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
}

// HandleGetSiteConfig 获取站点配置
// GET /api/admin/site
func (h *AdminLegacyHandler) HandleGetSiteConfig(c *gin.Context) {
	config, err := h.adminStorage.GetAdminConfig(c.Request.Context())
	if err != nil {
		h.logger.Error("获取配置失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取配置失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(config.SiteConfig))
}
