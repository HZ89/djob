/*
 * Copyright (c) 2017.  Harrison Zhu <wcg6121@gmail.com>
 * This file is part of djob <https://github.com/HZ89/djob>.
 *
 * djob is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * djob is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with djob.  If not, see <http://www.gnu.org/licenses/>.
 */

package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"

	"github.com/HZ89/djob/errors"
	"github.com/HZ89/djob/log"
	pb "github.com/HZ89/djob/message"
)

// ApiController is interface used to operate all objs
type ApiController interface {
	ListRegions() ([]string, error)
	ListNode(*pb.Node, *pb.Search) ([]*pb.Node, int32, error)
	AddJob(*pb.Job) (*pb.Job, error)
	ModifyJob(*pb.Job) (*pb.Job, error)
	DeleteJob(*pb.Job) (*pb.Job, error)
	ListJob(in *pb.Job, search *pb.Search) ([]*pb.Job, int32, error)
	RunJob(name, region string) (*pb.Execution, error)
	GetStatus(name, region string) (*pb.JobStatus, error)
	ListExecutions(in *pb.Execution, search *pb.Search) ([]*pb.Execution, int32, error)
	Search(interface{}, *pb.Search) ([]interface{}, int32, error)
}

// tls cert key pair
type KayPair struct {
	Key  string
	Cert string
}

// APIServer body structure
type APIServer struct {
	bindIP    string            // listen to the ip
	bindPort  int               // listen to the port
	tokenList map[string]string // store auth token
	mc        ApiController     // api interface
	loger     *logrus.Entry     // custom logger
	keyPair   *KayPair          // tls key pair
	tls       bool              // use tls or not
	router    *gin.Engine       // gin obj
	server    *http.Server      // http server obj
}

// Initialize APIServer
func NewAPIServer(ip string, port int,
	tokens map[string]string, tls bool, pair *KayPair, backend ApiController) (*APIServer, error) {
	if len(tokens) == 0 {
		return nil, errors.ErrMissApiToken
	}
	// reverse tokens
	n := make(map[string]string)
	for k, v := range tokens {
		if _, exist := n[v]; exist {
			return nil, errors.ErrRepetionToken
		}
		n[v] = k
	}
	return &APIServer{
		bindIP:    ip,
		bindPort:  port,
		loger:     log.FmdLoger,
		tokenList: n,
		tls:       tls,
		keyPair:   pair,
		mc:        backend,
	}, nil
}

func (a *APIServer) prepareGin() *gin.Engine {

	if a.loger.Logger.Level.String() == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(a.logMiddleware())
	r.Use(a.tokenAuthMiddleware())
	r.Use(gin.Recovery())
	if a.tls {
		r.Use(a.tlsHeaderMiddleware())
	}
	web := r.Group("/api")
	web.Use(gzip.Gzip(gzip.DefaultCompression))
	web.GET("/region", a.listRegions)
	web.GET("/node", a.listNodes)

	jobAPI := web.Group("/job")
	jobAPI.POST("/", a.addJob)
	jobAPI.PUT("/", a.modifyJob)
	jobAPI.GET("/", a.listJobs)
	jobAPI.GET("/search", a.jobSearch)
	jobAPI.GET("/status/:region/:name", a.getJobStatus)
	jobAPI.GET("/run/:region/:name", a.runJob)
	jobAPI.DELETE("/:region/:name", a.deleteJob)

	executionAPI := web.Group("/execution")
	executionAPI.GET("/", a.listExecutions)
	executionAPI.GET("/search", a.executionSearch)

	return r
}

func (a *APIServer) listNodes(c *gin.Context) {
	var nqs pb.ApiNodeQueryString
	var s *pb.Search
	if err := c.MustBindWith(&nqs, jsonpbBinding{}); err != nil {
		a.respondWithError(http.StatusBadRequest, &pb.ApiNodeResponse{Succeed: false, Message: err.Error()}, c)
		return
	}
	if nqs.Pageing != nil {
		s = &pb.Search{
			Count:    nqs.Pageing.OutMaxPage,
			PageSize: nqs.Pageing.PageSize,
			PageNum:  nqs.Pageing.PageNum,
		}
	}
	nodes, count, err := a.mc.ListNode(nqs.Node, s)
	if err != nil {
		a.respondWithError(http.StatusInternalServerError, &pb.ApiStringResponse{Succeed: false, Message: err.Error()}, c)
		return
	}
	resp := &pb.ApiNodeResponse{Succeed: true, Message: "Succeed", Data: nodes, MaxPageNum: count}
	c.Render(http.StatusOK, pbjson{data: resp})
}

func (a *APIServer) listExecutions(c *gin.Context) {
	var eqs pb.ApiExecutionQueryString
	var s *pb.Search
	if err := c.MustBindWith(&eqs, jsonpbBinding{}); err != nil {
		a.respondWithError(http.StatusBadRequest, &pb.ApiJobResponse{Succeed: false, Message: err.Error()}, c)
		return
	}
	if eqs.Execution == nil {
		a.respondWithError(http.StatusBadRequest, &pb.ApiJobResponse{Succeed: false, Message: "no input execution obj"}, c)
		return
	}
	if eqs.Pageing != nil {
		s = &pb.Search{
			Count:    eqs.Pageing.OutMaxPage,
			PageSize: eqs.Pageing.PageSize,
			PageNum:  eqs.Pageing.PageNum,
		}
	}

	outs, count, err := a.mc.ListExecutions(eqs.Execution, s)
	if err != nil {
		a.respondWithError(http.StatusInternalServerError, &pb.ApiStringResponse{Succeed: false, Message: err.Error()}, c)
		return
	}
	resp := &pb.ApiExecutionResponse{Succeed: true, Message: "Succeed", Data: outs, MaxPageNum: count}
	c.Render(http.StatusOK, pbjson{data: resp})
}

func (a *APIServer) jobSearch(c *gin.Context) {
	a.search(&pb.Job{}, c)
}

func (a *APIServer) executionSearch(c *gin.Context) {
	a.search(&pb.Execution{}, c)
}

func (a *APIServer) search(obj interface{}, c *gin.Context) {
	var searchQuery pb.ApiSearchQueryString
	if err := c.MustBindWith(&searchQuery, jsonpbBinding{}); err != nil {
		a.respondWithError(http.StatusInternalServerError, &pb.ApiStringResponse{Succeed: false, Message: err.Error()}, c)
		return
	}
	if searchQuery.SearchCondition == nil {
		a.respondWithError(http.StatusBadRequest, &pb.ApiJobResponse{Succeed: false, Message: "no input query condition"}, c)
		return
	}
	search := &pb.Search{
		Conditions: searchQuery.SearchCondition.Conditions,
		Links:      searchQuery.SearchCondition.Links,
	}
	if searchQuery.Pageing != nil {
		search.Count = searchQuery.Pageing.OutMaxPage
		search.PageNum = searchQuery.Pageing.PageNum
		search.PageSize = searchQuery.Pageing.PageSize
	}
	outs, count, err := a.mc.Search(obj, search)

	if err != nil {
		a.respondWithError(http.StatusInternalServerError, &pb.ApiStringResponse{Succeed: false, Message: err.Error()}, c)
		return
	}

	switch obj.(type) {
	case *pb.Job:
		resp := &pb.ApiJobResponse{}
		for _, out := range outs {
			if t, ok := out.(*pb.Job); ok {
				resp.Data = append(resp.Data, t)
			} else {
				a.respondWithError(http.StatusInternalServerError, &pb.ApiJobResponse{Succeed: false, Message: errors.ErrNotExpectation.Error()}, c)
				log.FmdLoger.WithError(errors.ErrNotExpectation).Fatal("API: Search result error")
				return
			}
		}
		resp.MaxPageNum = int32(count)
		resp.Succeed = true
		resp.Message = "Succeed"
		c.Render(http.StatusOK, pbjson{data: resp})
		return
	case *pb.Execution:
		resp := &pb.ApiExecutionResponse{}
		for _, out := range outs {
			if t, ok := out.(*pb.Execution); ok {
				resp.Data = append(resp.Data, t)
			} else {
				a.respondWithError(http.StatusInternalServerError, &pb.ApiJobResponse{Succeed: false, Message: errors.ErrNotExpectation.Error()}, c)
				log.FmdLoger.WithError(errors.ErrNotExpectation).Fatal("API: Search result error")
				return
			}
		}
		resp.MaxPageNum = int32(count)
		resp.Succeed = true
		resp.Message = "Succeed"
		c.Render(http.StatusOK, pbjson{data: resp})
		return
	default:
		a.respondWithError(http.StatusInternalServerError, &pb.ApiStringResponse{Succeed: false, Message: errors.ErrType.Error()}, c)
		return
	}
}

func (a *APIServer) listRegions(c *gin.Context) {
	var resp pb.ApiStringResponse
	var err error
	resp.Data, err = a.mc.ListRegions()
	if err != nil {
		resp.Message = err.Error()
		a.respondWithError(http.StatusInternalServerError, &resp, c)
		return
	}
	resp.Succeed = true
	resp.Message = "Succeed"
	c.Render(http.StatusOK, pbjson{data: &resp})
}

func (a *APIServer) addJob(c *gin.Context) {
	var job pb.Job
	err := c.MustBindWith(&job, jsonpbBinding{})
	if err != nil {
		a.respondWithError(http.StatusBadRequest, &pb.ApiJobResponse{Succeed: false, Message: err.Error()}, c)
		return
	}
	a.jobCRUD(&job, pb.Ops_ADD, nil, c)
}

func (a *APIServer) modifyJob(c *gin.Context) {
	var job pb.Job
	err := c.MustBindWith(&job, jsonpbBinding{})
	if err != nil {
		a.respondWithError(http.StatusBadRequest, &pb.ApiJobResponse{Succeed: false, Message: err.Error()}, c)
		return
	}
	a.jobCRUD(&job, pb.Ops_MODIFY, nil, c)
}

func (a *APIServer) listJobs(c *gin.Context) {
	var jqs pb.ApiJobQueryString
	var s *pb.Search
	if err := c.MustBindWith(&jqs, jsonpbBinding{}); err != nil {
		a.respondWithError(http.StatusBadRequest, &pb.ApiJobResponse{Succeed: false, Message: err.Error()}, c)
		return
	}
	if jqs.Job == nil {
		a.respondWithError(http.StatusBadRequest, &pb.ApiJobResponse{Succeed: false, Message: "no input job obj"}, c)
		return
	}
	if jqs.Pageing != nil {
		s = &pb.Search{
			Count:    jqs.Pageing.OutMaxPage,
			PageSize: jqs.Pageing.PageSize,
			PageNum:  jqs.Pageing.PageNum,
		}
	}

	a.jobCRUD(jqs.Job, pb.Ops_READ, s, c)
}

func (a *APIServer) deleteJob(c *gin.Context) {
	name := c.Params.ByName("name")
	region := c.Params.ByName("region")
	a.jobCRUD(&pb.Job{Name: name, Region: region}, pb.Ops_DELETE, nil, c)
}

func (a *APIServer) jobCRUD(in *pb.Job, ops pb.Ops, search *pb.Search, c *gin.Context) {

	if ops == pb.Ops_ADD || ops == pb.Ops_MODIFY {
		// this fields are forbidden to change by user
		in.SchedulerNodeName = ""
		in.ParentJobName = ""
	}
	var count int32

	resp := pb.ApiJobResponse{}
	var err error
	var out *pb.Job
	switch ops {
	case pb.Ops_READ:
		resp.Data, count, err = a.mc.ListJob(in, search)
	case pb.Ops_ADD:
		out, err = a.mc.AddJob(in)
	case pb.Ops_MODIFY:
		out, err = a.mc.ModifyJob(in)
	case pb.Ops_DELETE:
		out, err = a.mc.DeleteJob(in)
	default:
		err = errors.ErrUnknownOps
	}
	if err != nil {
		resp.Succeed = false
		resp.Message = err.Error()
		a.respondWithError(http.StatusInternalServerError, &resp, c)
		return
	}
	if out != nil {
		resp.Data = append(resp.Data, out)
	}
	resp.Succeed = true
	resp.Message = "Succeed"
	resp.MaxPageNum = count
	c.Render(http.StatusOK, pbjson{data: &resp})
}

func (a *APIServer) getJobStatus(c *gin.Context) {
	var resp pb.ApiJobStatusResponse
	name := c.Params.ByName("name")
	region := c.Params.ByName("region")
	status, err := a.mc.GetStatus(name, region)
	if err != nil {
		resp.Succeed = false
		resp.Message = err.Error()
		a.respondWithError(http.StatusInternalServerError, &resp, c)
		return
	}
	resp.Succeed = true
	resp.Message = "Succeed"
	resp.Data = append(resp.Data, status)
	c.Render(http.StatusOK, pbjson{data: &resp})
}

func (a *APIServer) runJob(c *gin.Context) {
	var resp pb.ApiExecutionResponse
	name := c.Params.ByName("name")
	region := c.Params.ByName("region")
	execution, err := a.mc.RunJob(name, region)
	if err != nil {
		resp.Succeed = false
		resp.Message = err.Error()
		a.respondWithError(http.StatusInternalServerError, &resp, c)
		return
	}
	resp.Data = append(resp.Data, execution)
	resp.Succeed = true
	resp.Message = "Succeed"
	c.Render(http.StatusOK, pbjson{data: &resp})
}

func (a *APIServer) tlsHeaderMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		c.Writer.Header().Set("Strict-Transport-Security",
			"max-age=63072000; includeSubDomains")
	}
}

// logger middleware
func (a *APIServer) logMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		entry := a.loger.WithFields(logrus.Fields{
			"client_ip":  c.ClientIP(),
			"method":     c.Request.Method,
			"path":       path,
			"status":     c.Writer.Status(),
			"latency":    latency,
			"user-agent": c.Request.UserAgent(),
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.String())
		} else {
			entry.Info()
		}
	}
}

// render error respond and quit
func (a *APIServer) respondWithError(code int, pb interface{}, c *gin.Context) {
	c.Render(code, pbjson{data: pb.(proto.Message)})
	c.AbortWithStatus(code)
}

// auth middleware used in gin
func (a *APIServer) tokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("X-Auth-Token")

		if token == "" {
			a.respondWithError(http.StatusUnauthorized, &pb.ApiJobResponse{Succeed: false, Message: "API token required"}, c)
			return
		}
		if _, exist := a.tokenList[token]; exist {
			c.Next()
		} else {
			a.respondWithError(http.StatusUnauthorized, &pb.ApiJobResponse{Succeed: false, Message: "API token Error"}, c)
			return
		}
	}
}

// Run APIServer
func (a *APIServer) Run() {
	r := a.prepareGin()
	a.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", a.bindIP, a.bindPort),
		Handler: r,
	}
	if a.tls {
		a.server.TLSConfig = &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP256, tls.CurveP384, tls.CurveP521},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		}
		a.server.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)
		go a.server.ListenAndServeTLS(a.keyPair.Cert, a.keyPair.Key)
		return
	}
	go a.server.ListenAndServe()
	return
}

// stop APIServer
func (a *APIServer) Stop(wait time.Duration) error {
	a.loger.Infof("API-server: shutdown in %d second", wait)
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}
	a.loger.Info("API-server: bye-bye")
	return nil
}
