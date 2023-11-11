package strategy

import (
	"fmt"

	cache "github.com/chenyahui/gin-cache"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func CacheByRequestBody[T interface{}](getKey func(T) string) cache.Option {
	return cache.WithCacheStrategyByRequest(func(ctx *gin.Context) (bool, cache.Strategy) {
		var req T
		if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
			return false, cache.Strategy{}
		}

		key := fmt.Sprintf("%v/%v", ctx.Request.RequestURI, getKey(req))

		return true, cache.Strategy{
			CacheKey: key,
		}
	})
}
