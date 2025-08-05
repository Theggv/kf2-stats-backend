package leaderboards

import (
	"fmt"
	"slices"
	"time"

	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/strategy"
	"github.com/theggv/kf2-stats-backend/pkg/common/util"
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
			strategy.CacheByRequestBody(func(req LeaderBoardsRequest) string {
				slices.Sort(req.ServerIds)

				return fmt.Sprintf("%v/%v/%v/%v/%v/%v",
					util.IntArrayToString(req.ServerIds, ","),
					req.OrderBy, req.Perk, req.Page,
					req.From.Format("2006-01-02"), req.To.Format("2006-01-02"),
				)
			}),
		),
		controller.getLeaderBoard)
}
