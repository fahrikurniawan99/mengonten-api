package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"mengonten-api/utils"
	"mengonten-api/worker"
)

// @Summary Get YouTube video metadata
// @Description Ambil title, duration, thumbnail, tags, is_live dari URL YouTube
// @Tags YouTube
// @Produce json
// @Security Bearer
// @Param url query string true "YouTube video URL"
// @Success 200 {object} utils.Response "Metadata retrieved"
// @Router /api/youtube/metadata [get]
func GetYouTubeMetadata(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		url := c.Query("url")
		if url == "" {
			utils.ErrorResponse(c, http.StatusBadRequest, "Parameter url wajib diisi")
			return
		}

		metadata, err := worker.ExtractYouTubeMetadata(url)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil metadata: "+err.Error())
			return
		}

		utils.SuccessResponse(c, http.StatusOK, "Metadata retrieved", metadata)
	}
}
