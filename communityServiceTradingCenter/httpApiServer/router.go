package httpApiServer

import (
	"encoding/json"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"moonlighting/communityServiceTradingCenter/dataManager"
	"net/http"
)

type response struct {
	Succeed bool        `json:"succeed"`
	Data    interface{} `json:"data"`
}

func sendResponse(context *gin.Context, succeed bool, data any) {
	resData, _ := json.Marshal(response{
		Succeed: succeed,
		Data:    data,
	})
	context.Data(http.StatusOK, "application/json", resData)
	context.Abort()
}

func (p *Server) route() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(cors.Default())

	v1router := r.Group("v1")

	p.routeV1Static(v1router)
	p.routeV1Api(v1router)

	return r
}

func (p *Server) routeV1Static(r *gin.RouterGroup) {

	staticRoute := r.Group("/static")
	staticRoute.Static("/", p.staticServePath)

}

func (p *Server) routeV1Api(r *gin.RouterGroup) {

	apiRoute := r.Group("/api")

	providerRoute := apiRoute.Group("/provider")
	providerRoute.POST("/query", func(context *gin.Context) {
		type localReq struct {
			Limit      int               `json:"limit"`
			Page       int               `json:"page"`
			MatchRules map[string]string `json:"matchRules"`
		}
		var req localReq
		err := context.BindJSON(&req)
		if err != nil {
			sendResponse(context, false, "parse json failed"+err.Error())
			return
		}

		list, count, totalCount, err := p.providerDataManager.QueryData(req.Limit, req.Page, req.MatchRules)
		if err != nil {
			sendResponse(context, false, "query failed : "+err.Error())
			return
		}

		resMap := make(map[string]any)

		resMap["count"] = count
		resMap["totalCount"] = totalCount
		resMap["queryList"] = list

		sendResponse(context, true, resMap)

	})

	providerRoute.POST("/insert", func(context *gin.Context) {
		type localReq struct {
			DataList []dataManager.Data `json:"dataList"`
		}

		var req localReq
		err := context.BindJSON(&req)
		if err != nil {
			sendResponse(context, false, "parse json failed : "+err.Error())
			return
		}

		err = p.providerDataManager.InsertData(req.DataList)
		if err != nil {
			sendResponse(context, false, "insert data failed : "+err.Error())
			return
		}

		sendResponse(context, true, nil)
	})

	providerRoute.POST("/delete", func(context *gin.Context) {
		type localReq struct {
			KeyList []string `json:"keyList"`
		}

		var req localReq
		err := context.BindJSON(&req)
		if err != nil {
			sendResponse(context, false, "parse json failed : "+err.Error())
			return
		}

		err = p.providerDataManager.DeleteData(req.KeyList)
		if err != nil {
			sendResponse(context, false, "delete data failed : "+err.Error())
			return
		}

		sendResponse(context, true, nil)
	})

	publishRouter := apiRoute.Group("/publish")
	publishRouter.POST("/query", func(context *gin.Context) {
		type localReq struct {
			Limit      int               `json:"limit"`
			Page       int               `json:"page"`
			MatchRules map[string]string `json:"matchRules"`
		}
		var req localReq
		err := context.BindJSON(&req)
		if err != nil {
			sendResponse(context, false, "parse json failed"+err.Error())
			return
		}

		list, count, totalCount, err := p.publishDataManager.QueryData(req.Limit, req.Page, req.MatchRules)
		if err != nil {
			sendResponse(context, false, "query failed : "+err.Error())
			return
		}

		resMap := make(map[string]any)

		resMap["count"] = count
		resMap["totalCount"] = totalCount
		resMap["queryList"] = list

		sendResponse(context, true, resMap)

	})

	publishRouter.POST("/insert", func(context *gin.Context) {
		type localReq struct {
			DataList []dataManager.Data `json:"dataList"`
		}

		var req localReq
		err := context.BindJSON(&req)
		if err != nil {
			sendResponse(context, false, "parse json failed"+err.Error())
			return
		}

		err = p.publishDataManager.InsertData(req.DataList)
		if err != nil {
			sendResponse(context, false, "insert failed : "+err.Error())
			return
		}

		sendResponse(context, true, nil)
	})

	publishRouter.POST("/delete", func(context *gin.Context) {
		type localReq struct {
			KeyList []string `json:"keyList"`
		}

		var req localReq
		err := context.BindJSON(&req)
		if err != nil {
			sendResponse(context, false, "parse json failed"+err.Error())
			return
		}

		err = p.publishDataManager.DeleteData(req.KeyList)
		if err != nil {
			sendResponse(context, false, "delete failed : "+err.Error())
			return
		}

		sendResponse(context, true, nil)
	})

	recommendRoute := apiRoute.Group("/recommend")
	recommendRoute.POST("/query", func(context *gin.Context) {
		type localReq struct {
			Limit      int               `json:"limit"`
			Page       int               `json:"page"`
			MatchRules map[string]string `json:"matchRules"`
		}
		var req localReq
		err := context.BindJSON(&req)
		if err != nil {
			sendResponse(context, false, "parse json failed"+err.Error())
			return
		}

		list, count, totalCount, err := p.recommendDataManager.QueryData(req.Limit, req.Page, req.MatchRules)
		if err != nil {
			sendResponse(context, false, "query failed : "+err.Error())
			return
		}

		resMap := make(map[string]any)

		resMap["count"] = count
		resMap["totalCount"] = totalCount
		resMap["queryList"] = list

		sendResponse(context, true, resMap)

	})

	recommendRoute.POST("/insert", func(context *gin.Context) {
		type localReq struct {
			DataList []dataManager.Data `json:"dataList"`
		}

		var req localReq
		err := context.BindJSON(&req)
		if err != nil {
			sendResponse(context, false, "parse json failed"+err.Error())
			return
		}

		err = p.recommendDataManager.InsertData(req.DataList)
		if err != nil {
			sendResponse(context, false, "insert failed : "+err.Error())
			return
		}

		sendResponse(context, true, nil)
	})

	recommendRoute.POST("/delete", func(context *gin.Context) {
		type localReq struct {
			KeyList []string `json:"keyList"`
		}

		var req localReq
		err := context.BindJSON(&req)
		if err != nil {
			sendResponse(context, false, "parse json failed"+err.Error())
			return
		}

		err = p.recommendDataManager.DeleteData(req.KeyList)
		if err != nil {
			sendResponse(context, false, "delete failed : "+err.Error())
			return
		}

		sendResponse(context, true, nil)
	})
}
