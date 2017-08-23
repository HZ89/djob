package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"net/http"
	"time"
	pb "version.uuzu.com/zhuhuipeng/djob/message"
)

type Backend interface {
	JobModify(job *pb.Job) (*pb.Job, error)
	JobInfo(name, region string) (*pb.Job, error)
	JobDelete(name, region string) (*pb.Job, error)
	JobList(region string) ([]*pb.Job, error)
	JobRun(name, region string) error
}

type KayPair struct {
	Key  string
	Cert string
}

var jsonContentType = []string{"application/json; charset=utf-8"}

//jsonString implement output a protobuf obj as a json string and json content type
type pbjson struct {
	data interface{}
}

func (j pbjson) Render(w http.ResponseWriter) error {
	var buf bytes.Buffer
	marshaler := &jsonpb.Marshaler{}
	if err := marshaler.Marshal(&buf, j.data.(proto.Message)); err != nil {
		return err
	}
	w.Write(buf.Bytes())
	return nil
}

func (j pbjson) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = jsonContentType
	}
}

type jsonpbBinding struct{}

func (jsonpbBinding) Name() string {
	return "jsonpb"
}

func (jsonpbBinding) Bind(req *http.Request, obj interface{}) error {
	jsondec := json.NewDecoder(req.Body)
	unmarshaler := &jsonpb.Unmarshaler{AllowUnknownFields: false}
	if err := unmarshaler.UnmarshalNext(jsondec, obj.(proto.Message)); err != nil {
		return err
	}

	return nil
}

type APIServer struct {
	bindIP    string
	bindPort  int
	tokenList map[string]string
	backend   Backend
	loger     *logrus.Entry
	keyPair   *KayPair
	tls       bool
	router    *gin.Engine
	server    *http.Server
}

func NewAPIServer(ip string, port int, loger *logrus.Entry,
	tokens map[string]string, tls bool, pair *KayPair) (*APIServer, error) {
	if len(tokens) == 0 {
		return nil, errors.New("Have no tokens")
	}
	// reverse tokens
	n := make(map[string]string)
	for k, v := range tokens {
		if _, exist := n[v]; exist {
			return nil, errors.New("Have repetition token")
		}
		n[v] = k
	}
	return &APIServer{
		bindIP:    ip,
		bindPort:  port,
		loger:     loger,
		tokenList: n,
		tls:       tls,
		keyPair:   pair,
	}, nil
}

func (a *APIServer) prepareGin() *gin.Engine {
	r := gin.New()
	r.Use(a.logMiddleware())
	r.Use(a.tokenAuthMiddleware())
	r.Use(gin.Recovery())
	if a.tls {
		r.Use(a.tlsHeaderMiddleware())
	}
	web := r.Group("/web")
	web.Use(gzip.Gzip(gzip.DefaultCompression))
	web.POST("/jobs", a.modJob)
	web.GET("/:region/jobs", a.getJobList)
	web.GET("/:region/jobs/:name", a.getJob)
	web.DELETE("/:region/jobs/:name", a.deleteJob)
	return r
}

func (a *APIServer) deleteJob(c *gin.Context) {
	name := c.Params.ByName("name")
	region := c.Params.ByName("region")
	job, err := a.backend.JobDelete(name, region)
	if err != nil {
		a.respondWithError(http.StatusInternalServerError, &pb.RespJob{Status: http.StatusInternalServerError, Message: err.Error()}, c)
	}
	resp := pb.RespJob{
		Status:  0,
		Message: "succeed",
		Data:    []*pb.Job{job},
	}
	c.Render(http.StatusOK, pbjson{data: resp})
}

// TODO: add a data filter
func (a *APIServer) getJobList(c *gin.Context) {
	region := c.Params.ByName("region")
	jobs, err := a.backend.JobList(region)
	if err != nil {
		a.respondWithError(http.StatusInternalServerError, &pb.RespJob{Status: http.StatusInternalServerError, Message: err.Error()}, c)
	}
	resp := pb.RespJob{
		Status:  0,
		Message: "succeed",
		Data:    jobs,
	}
	c.Render(http.StatusOK, pbjson{data: resp})
}
func (a *APIServer) getJob(c *gin.Context) {
	name := c.Params.ByName("name")
	region := c.Params.ByName("region")
	job, err := a.backend.JobInfo(name, region)
	if err != nil {
		a.respondWithError(http.StatusInternalServerError, &pb.RespJob{Status: http.StatusInternalServerError, Message: err.Error()}, c)
	}
	resp := pb.RespJob{
		Status:  0,
		Message: "succeed",
		Data:    []*pb.Job{job},
	}
	c.Render(http.StatusOK, pbjson{data: resp})
}

func (a *APIServer) modJob(c *gin.Context) {
	var (
		job  *pb.Job
		resp *pb.RespJob
		err  error
	)
	err = c.MustBindWith(job, jsonpbBinding{})
	if err != nil {
		a.respondWithError(http.StatusBadRequest, &pb.RespJob{Status: http.StatusBadRequest, Message: err.Error()}, c)
	}
	job, err = a.backend.JobModify(job)
	if err != nil {
		a.respondWithError(http.StatusInternalServerError, &pb.RespJob{Status: http.StatusInternalServerError, Message: err.Error()}, c)
	}
	resp.Status = 0
	resp.Message = "succeed"
	resp.Data = append(resp.Data, job)

	c.Render(http.StatusOK, pbjson{data: resp})
}

func (a *APIServer) tlsHeaderMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		c.Writer.Header().Set("Strict-Transport-Security",
			"max-age=63072000; includeSubDomains")
	}
}

func (a *APIServer) logMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		entry := a.loger.WithFields(logrus.Fields{
			"client_ip":    c.ClientIP(),
			"method":       c.Request.Method,
			"path":         path,
			"latency":      latency,
			"user-agent":   c.Request.UserAgent(),
			"respond-time": end.Format(time.RFC3339),
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.String())
		} else {
			entry.Info()
		}
	}
}

func (a *APIServer) respondWithError(code int, pb interface{}, c *gin.Context) {
	resp, err := proto.Marshal(pb.(proto.Message))
	if err != nil {
		c.String(http.StatusInternalServerError, "protp decode error: %s", err.Error())
	}
	c.Render(code, pbjson{data: resp})
	c.AbortWithStatus(code)
}

func (a *APIServer) tokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("X-Auth-Token")

		if token == "" {
			a.respondWithError(http.StatusUnauthorized, "API token required", c)
			return
		}
		if _, exist := a.tokenList[token]; exist {
			c.Next()
		} else {
			a.respondWithError(http.StatusUnauthorized, "API token Error", c)
			return
		}
	}
}

func (a *APIServer) Run() error {
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
			},
		}
		a.server.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)
		return a.server.ListenAndServeTLS(a.keyPair.Cert, a.keyPair.Key)
	}
	return a.server.ListenAndServe()
}

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
