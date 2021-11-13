package utils

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/rest-api/app/libraries/jsoner"
	"github.com/rest-api/app/libraries/logger"
)

type ApiResponse struct {
	Error  string      `json:"error"`
	Result interface{} `json:"result"`
	Status int         `json:"status"`
}

type HttpPackage struct {
	W        http.ResponseWriter
	R        *http.Request
	Response *ApiResponse
}

func (hp *HttpPackage) SendResponse() {
	logger.InfoLogger.Println("sending response to client")
	hp.W.WriteHeader(hp.Response.Status)
	if err := jsoner.EncodeJSON(hp.W, hp.Response); err != nil {
		logger.ErrorLogger.Println(err.Error())
		hp.W.Write([]byte(http.StatusText(hp.Response.Status)))
		hp.W.WriteHeader(hp.Response.Status)
		return
	}
	logger.InfoLogger.Println("sent json to client successfully")
}

func GetURIKeys(r *http.Request, paramKey string, expectLen int) (interface{}, error) {
	URIKeys, OK := r.URL.Query()[paramKey]
	if !OK || len(URIKeys) != expectLen {
		errStr := "url param \"key\" is not existed or range is not satisfied"
		logger.ErrorLogger.Println(errStr)
		return nil, errors.New(errStr)
	}
	return URIKeys, nil
}

func StructHandler(structVal reflect.Value) error {
	missings := []string{}
	structType := structVal.Type()
	for i := 0; i < structType.NumField(); i++ {
		key := structType.Field(i).Tag.Get("json")
		val := structVal.Field(i).Interface()
		if val == "" || val == nil {
			missings = append(missings, key)
		}
	}
	if len(missings) > 0 {
		err := fmt.Errorf("json: missing field %q", strings.Join(missings, ","))
		logger.ErrorLogger.Println(err.Error())
		return err
	}
	return nil
}

func CorsHandler(f interface{}, method string) http.HandlerFunc {
	ControllerFunc, isControllerFuncExist := f.(func(*HttpPackage))
	return func(w http.ResponseWriter, r *http.Request) {
		hp := &HttpPackage{
			W:        w,
			R:        r,
			Response: &ApiResponse{},
		}
		if r.Method != method || !isControllerFuncExist {
			err := fmt.Errorf("%v - %v not implemented", r.URL.Path, r.Method)
			logger.ErrorLogger.Printf(err.Error())
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(err.Error()))
			return
		}
		logger.InfoLogger.Printf("%v - %v Received Request", r.URL.Path, r.Method)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		ControllerFunc(hp)
	}
}

func NotFoundHandler(MatchRouteWithURLFunc func(string) bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		isURLPathExist := false
		if MatchRouteWithURLFunc(r.URL.Path) {
			isURLPathExist = true
		}

		if !isURLPathExist {
			err := fmt.Errorf("%v - %v not found", r.URL.Path, r.Method)
			logger.ErrorLogger.Printf(err.Error())
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		}
	}
}
