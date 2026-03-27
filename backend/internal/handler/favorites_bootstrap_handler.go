package handler

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

const favoritesBootstrapPageSize = 10000

type FavoritesBootstrapHandler struct {
	storage model.StorageService
	logger  *zap.Logger
}

type favoritesBootstrapFavorite struct {
	ID         string `json:"id"`
	Source     string `json:"source"`
	SourceName string `json:"source_name,omitempty"`
	Title      string `json:"title"`
	Cover      string `json:"cover"`
	Year       string `json:"year,omitempty"`
	SaveTime   int64  `json:"save_time,omitempty"`
}

type favoritesBootstrapHistory struct {
	ID            string `json:"id"`
	Source        string `json:"source"`
	SourceName    string `json:"source_name,omitempty"`
	Title         string `json:"title"`
	Cover         string `json:"cover"`
	Year          string `json:"year,omitempty"`
	Index         int    `json:"index,omitempty"`
	TotalEpisodes int    `json:"total_episodes,omitempty"`
	PlayTime      int    `json:"play_time,omitempty"`
	TotalTime     int    `json:"total_time,omitempty"`
	SaveTime      int64  `json:"save_time,omitempty"`
	LastPlayTime  int64  `json:"last_play_time,omitempty"`
}

type favoritesBootstrapResponse struct {
	Username  string                       `json:"username"`
	Favorites []favoritesBootstrapFavorite `json:"favorites"`
	History   []favoritesBootstrapHistory  `json:"history"`
}

func NewFavoritesBootstrapHandler(
	storage model.StorageService,
	logger *zap.Logger,
) *FavoritesBootstrapHandler {
	return &FavoritesBootstrapHandler{
		storage: storage,
		logger:  logger,
	}
}

func (h *FavoritesBootstrapHandler) GetBootstrap(c *gin.Context) {
	response, statusCode, code, message := h.buildBootstrapResponse(c)
	if message != "" {
		c.JSON(statusCode, model.Error(code, message))
		return
	}

	c.JSON(http.StatusOK, model.Success(response))
}

func (h *FavoritesBootstrapHandler) GetBootstrapLegacy(c *gin.Context) {
	response, statusCode, _, message := h.buildBootstrapResponse(c)
	if message != "" {
		c.JSON(statusCode, gin.H{"error": message})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *FavoritesBootstrapHandler) buildBootstrapResponse(
	c *gin.Context,
) (*favoritesBootstrapResponse, int, int, string) {
	username := resolveUsernameFromContext(c)
	if username == "" {
		return nil, http.StatusUnauthorized, model.CodeUnauthorized, "未登录"
	}

	response := &favoritesBootstrapResponse{
		Username:  username,
		Favorites: []favoritesBootstrapFavorite{},
		History:   []favoritesBootstrapHistory{},
	}

	group, ctx := errgroup.WithContext(c.Request.Context())
	group.Go(func() error {
		favorites, _, err := h.storage.GetFavorites(
			ctx,
			username,
			1,
			favoritesBootstrapPageSize,
		)
		if err != nil {
			return err
		}
		response.Favorites = mapFavoritesBootstrapFavorites(favorites)
		return nil
	})

	group.Go(func() error {
		records, _, err := h.storage.GetPlayRecords(
			ctx,
			username,
			1,
			favoritesBootstrapPageSize,
		)
		if err != nil {
			return err
		}
		response.History = mapFavoritesBootstrapHistory(records)
		return nil
	})

	if err := group.Wait(); err != nil {
		h.logger.Error("favorites bootstrap failed", zap.String("username", username), zap.Error(err))
		return nil, http.StatusInternalServerError, model.CodeInternalError, "片单数据加载失败"
	}

	return response, http.StatusOK, model.CodeSuccess, ""
}

func mapFavoritesBootstrapFavorites(
	items []*model.Favorite,
) []favoritesBootstrapFavorite {
	list := make([]favoritesBootstrapFavorite, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		list = append(list, favoritesBootstrapFavorite{
			ID:         item.Source + "+" + item.VodID,
			Source:     item.Source,
			SourceName: item.SourceName,
			Title:      item.VodName,
			Cover:      item.VodPic,
			Year:       item.Year,
			SaveTime:   item.SaveTime,
		})
	}

	sort.SliceStable(list, func(left, right int) bool {
		return list[left].SaveTime > list[right].SaveTime
	})

	return list
}

func mapFavoritesBootstrapHistory(
	items []*model.PlayRecord,
) []favoritesBootstrapHistory {
	list := make([]favoritesBootstrapHistory, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		list = append(list, favoritesBootstrapHistory{
			ID:            item.Source + "+" + item.VodID,
			Source:        item.Source,
			SourceName:    item.SourceName,
			Title:         item.VodName,
			Cover:         item.VodPic,
			Year:          item.Year,
			Index:         item.EpisodeIndex,
			TotalEpisodes: item.TotalEpisodes,
			PlayTime:      item.Progress,
			TotalTime:     item.Duration,
			SaveTime:      item.SaveTime,
			LastPlayTime:  item.UpdatedAt,
		})
	}

	sort.SliceStable(list, func(left, right int) bool {
		return list[left].LastPlayTime > list[right].LastPlayTime
	})

	return list
}
