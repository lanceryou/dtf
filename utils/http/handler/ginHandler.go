package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"reflect"
	"time"
)

func GinHandler[T, U any](
	fn func(ctx context.Context, t T) (reply U, err error),
) gin.HandlerFunc {
	rt := reflect.TypeOf(fn).In(1).Elem()
	return func(c *gin.Context) {
		r := reflect.New(rt).Interface()
		data, err := c.GetRawData()
		if err != nil {
			fmt.Printf("http data ===== %v\n", string(data))
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		err = json.Unmarshal(data, r)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		now := time.Now()
		reply, err := fn(c.Request.Context(), r.(T))
		if err != nil {
			slog.Error("time_cost", time.Since(now).String(), "err", err)
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		slog.Info("time_cost", time.Since(now).String(), "reply", reply)
		c.Writer.Header().Set("Content-type", "application/json")
		c.JSON(http.StatusOK, reply)
	}
}
