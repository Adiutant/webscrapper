package http_server

import (
	"fmt"
	gin "github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	timeout "github.com/s-wijaya/gin-timeout"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	"webscrapper/models"
	"webscrapper/scrapper"
)

type httpApiServer struct {
	serverInstance *gin.Engine
	webcrawler     *scrapper.Scrapper
	config         *models.Config
	logger         *logrus.Logger
}

func (h *httpApiServer) SetAuthCheck(config *models.Config) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		user, pass, _ := ctx.Request.BasicAuth()
		if !(user == config.Login && pass == config.Password) {
			ctx.AbortWithError(http.StatusUnauthorized, fmt.Errorf("authorization error"))
			h.logger.Errorf("Unauthorized access")
			return
		}
		ctx.Next()
	}
}

func InitServer(configPrefix string) *httpApiServer {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		PrettyPrint:     true,
	})
	config := readConfig(configPrefix)
	serverMux := gin.Default()
	//serverMux.Handle("/post_info", http.TimeoutHandler(http.HandlerFunc(PostInfo), time.Duration(config.Timeout)*time.Second, "Timeout"))
	httpServer := &httpApiServer{
		serverInstance: serverMux,
		config:         config,
		logger:         logger,
	}
	authChecker := httpServer.SetAuthCheck(config)
	responseBodyTimeout := gin.H{
		"code":    http.StatusRequestTimeout,
		"message": "request timeout, response is sent from middleware"}
	httpServer.serverInstance.Use(authChecker).Use(timeout.TimeoutHandler(15*time.Minute, http.StatusRequestTimeout, responseBodyTimeout)).POST("/get_sorted_notebooks", httpServer.getNotebooks)

	return httpServer
}

func readConfig(prefix string) *models.Config {
	configFile, err := ioutil.ReadFile(fmt.Sprintf("config/%s_config.json", prefix))
	if err != nil {
		panic("Unable to read config.")
	}
	var config models.Config
	err = jsoniter.Unmarshal(configFile, &config)
	if err != nil {
		panic("Invalid config format.")
	}
	return &config

}
func (h *httpApiServer) StartServe() {
	err := h.serverInstance.Run(h.config.Port)
	if err != nil {
		panic("Unable to start serverInstance")
	}
}

//func (server httpApiServer) ShutdownServer(ctx *context.Context) {
//	err := server.serverInstance.(*ctx)
//	if err != nil {
//		panic("Error closing serverInstance")
//	}
//}

func (h *httpApiServer) getNotebooks(ctx *gin.Context) {

	var request models.Request
	if err := ctx.ShouldBindJSON(&request); err != nil {
		msg := fmt.Errorf("validation error: %s", err.Error())
		h.logger.WithFields(logrus.Fields{
			"handler": "PublishHandler",
		}).Errorf(msg.Error())
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": msg.Error()})
		return
	}
	if (request.LowPrice < 0) || (request.HighPrice < 0) || (request.LowPrice > request.HighPrice) {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	h.webcrawler = scrapper.NewScrapper()
	h.webcrawler.SetTriggers()
	done := make(chan bool)
	timer := time.NewTicker(time.Second)
	defer func() {
		done <- true
	}()
	go func() {
		for {
			select {
			case <-timer.C:
				ctx.Header("Connection", "Keep-Alive")
				ctx.String(http.StatusAccepted, "")

			case <-done:
				timer.Stop()
				return
			}
		}
	}()
	notebooks, err := h.webcrawler.StartCrawling(int(request.LowPrice), int(request.HighPrice))
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	notebooksArray := make([]models.Notebook, 0)
	for _, val := range notebooks {
		if val.Name == "" || val.Price == 0 || val.RAM == 0 || val.ScreenResolution == "" || val.Storage == 0 || val.CPUCores == 0 &&
			val.Ref == "" || val.CPUFrequency == 0 {
			continue
		}

		resolutionStrings := strings.Split(val.ScreenResolution, "х")
		x, err := strconv.ParseFloat(resolutionStrings[0], 64)
		if err != nil {
			h.logger.Error(err)
			return
		}
		y, err := strconv.ParseFloat(resolutionStrings[1], 64)
		if err != nil {
			h.logger.Error(err)
			return
		}
		val.Rating = (float64(val.CPUFrequency) * 0.2) + (float64(val.CPUCores) * 100) + (float64(val.RAM) * 0.15) + (float64(val.GPURAM) * 0.1) + (float64(val.Storage) * 0.15) + (float64(val.Price) * -0.01) + (x * y * 0.0001)
		notebooksArray = append(notebooksArray, val)

	}
	sort.Slice(notebooksArray, func(i, j int) bool {
		return notebooksArray[i].Rating > notebooksArray[j].Rating
	})
	if len(notebooksArray) == 0 {
		ctx.AbortWithStatus(http.StatusNoContent)
		h.logger.Println("No content")
		return
	}
	notebooksArray = notebooksArray[:10]

	ctx.AbortWithStatusJSON(http.StatusOK, notebooksArray)

}
