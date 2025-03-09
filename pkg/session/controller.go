package session

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/theggv/kf2-stats-backend/pkg/common/demorecord"
)

type sessionController struct {
	service *SessionService
}

// @Summary Creates a new session
// @Tags 	Session
// @Produce json
// @Param   session body    CreateSessionRequest true "Session JSON"
// @Success 201 {object} 	CreateSessionResponse
// @Router /sessions [post]
func (c *sessionController) create(ctx *gin.Context) {
	var req CreateSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	id, err := c.service.Create(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, CreateSessionResponse{
		Id: id,
	})
}

// @Summary Update session status
// @Tags 	Session
// @Produce json
// @Param   body body 		UpdateStatusRequest true "Body"
// @Success 200
// @Router /sessions/status [put]
func (c *sessionController) updateStatus(ctx *gin.Context) {
	var req UpdateStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	err := c.service.UpdateStatus(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

// @Summary Update session game data
// @Tags 	Session
// @Produce json
// @Param   body body 		UpdateGameDataRequest true "Body"
// @Success 200
// @Router /sessions/game-data [put]
func (c *sessionController) updateGameData(ctx *gin.Context) {
	var req UpdateGameDataRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	err := c.service.UpdateGameData(req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

// @Summary Upload demo
// @Tags 	Session
// @Produce json
// @Param   body body 		UpdateStatusRequest true "Body"
// @Success 201
// @Router /sessions/demo [post]
func (c *sessionController) uploadDemo(ctx *gin.Context) {
	raw, _ := ctx.GetRawData()

	err := c.service.UploadDemo(raw)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{})
}

// @Summary Get demo by session id
// @Tags 	Session
// @Produce json
// @Param   id path   	 	int true "Session id"
// @Success 200
// @Router /sessions/demo/{id} [get]
func (c *sessionController) getDemo(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Params.ByName("id"))
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	item, err := c.service.GetDemo(id)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	transform, err := demorecord.Transform(item)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		fmt.Printf("%v\n", err.Error())
		return
	}

	marshal, err := json.Marshal(transform)

	var b bytes.Buffer
	writer := gzip.NewWriter(&b)
	writer.Write(marshal)
	writer.Close()

	ctx.Header("Content-Encoding", "gzip")
	ctx.Writer.Header().Add("Vary", "Accept-Encoding")

	ctx.Data(http.StatusOK, "application/json", b.Bytes())
}
