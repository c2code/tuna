package auth

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"hexmeet.com/haishen/tuna/thirdparty/github.com/gin-contrib/cors"
	"hexmeet.com/haishen/tuna/thirdparty/github.com/gin-contrib/static"
	"hexmeet.com/haishen/tuna/utils"
)
type PingResponse struct {
	BaseResponse
	Message string `json:"message"`
}

//entry of web server
func (m Manager) webListen() {
	ginLogger := m.logger.Named("gin")

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(utils.Ginzap(ginLogger))
	router.Use(gin.Recovery())
	router.Use(cors.Default())
	router.Use(static.Serve("/", static.LocalFile("./dist", true)))
	router.Use(static.Serve("/auth", static.LocalFile("./dist", true)))

	//provide a internal access rest api
	tuna_v1:= router.Group("/")
	{
		tuna_v1.POST("/example", m.exampleRestCall)
	}

	//provide a external access rest api
	tuna_v2 := router.Group("/auth")
	{
		tuna_v2.GET("/ping", m.onPing) //to check the tuna service is accessful

		tuna_v2.GET("/log-level", m.onGetLogLevel) //get log level
		tuna_v2.POST("/log-level", m.onSetLogLevel) //set log level
	}

	portSpec := fmt.Sprintf(":%d", m.config.WebPort)

	router.Run(portSpec)
}

func (m Manager) onPing(c *gin.Context) {
	c.JSON(200, PingResponse{
		BaseResponse: BaseResponse{
			ErrCode: ErrCodeOk,
			ErrInfo: ErrInfoOk,
		},
		Message: "pong"})
}

