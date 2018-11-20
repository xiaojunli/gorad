package plugins

import "github.com/gin-gonic/gin"

type Plugin interface {
	Init(context *gin.Context, Engine *gin.Engine)
	Process(context *gin.Context, Engine *gin.Engine)
}
