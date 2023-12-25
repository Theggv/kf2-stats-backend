package leaderboards

import (
	"fmt"
	"time"

	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/strategy"
)

func RegisterRoutes(
	r *gin.RouterGroup,
	service *LeaderBoardsService,
	memoryStore *persist.MemoryStore,
) {
	controller := controller{
		service: service,
	}

	routes := r.Group("/leaderboards/")

	routes.POST("/",
		cache.Cache(memoryStore, 5*time.Minute,
			strategy.CacheByRequestBody[LeaderBoardsRequest](func(req LeaderBoardsRequest) string {
				return fmt.Sprintf("%v/%v/%v/%v",
					req.Type, req.Perk, req.From.Format("2006-01-02"), req.To.Format("2006-01-02"))
			}),
		),
		controller.getLeaderBoard)
}
