// Copyright 2018 The xfsgo Authors
// This file is part of the xfsgo library.
//
// The xfsgo library is free software: you can redistribute it and/or modify
// it under the terms of the MIT Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The xfsgo library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// MIT Lesser General Public License for more details.
//
// You should have received a copy of the MIT Lesser General Public License
// along with the xfsgo library. If not, see <https://mit-license.org/>.

package xfsgo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/token"
	"io"
	"io/ioutil"
	"net"
	"reflect"
	"strconv"
	"strings"
	"xfsgo/p2p/log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	jsonrpcVersion = "2.0"
)

var (
	parseError          = NewRPCError(-32700, "parse error")
	invalidRequestError = NewRPCError(-32600, "invalid request")
	methodNotFoundError = NewRPCError(-32601, "method not found")
	//invalidParamsError = NewRPCError(-32602, "invalid params")
	//internalError = NewRPCError(-32603, "invalid params")
)

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
	numCalls  uint
}

type service struct {
	name    string
	rcvr    reflect.Value
	typ     reflect.Type
	methods map[string]*methodType
}
type jsonRPCObj struct {
	jsonrpc string
	id      *int
	method  string
	params  interface{}
}
type jsonRPCRespErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RPCConfig struct {
	ListenAddr string
}

// RPCServer is an RPC server.
type RPCServer struct {
	logger     log.Logger
	config     *RPCConfig
	ginEngine  *gin.Engine
	upgrader   websocket.Upgrader
	serviceMap map[string]*service
}

func ginlogger(log log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(c.Errors) > 0 {
			log.Errorln(c.Errors.ByType(gin.ErrorTypePrivate).String())
		}
	}
}

//// NewRPCServer creates a new server instance with given configuration options.
func NewRPCServer(config *RPCConfig) *RPCServer {
	server := &RPCServer{
		logger:     log.DefaultLogger(),
		config:     config,
		serviceMap: make(map[string]*service),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}

	logger := log.DefaultLogger()
	gin.DefaultWriter = logger.Writer()
	gin.SetMode("release")
	server.ginEngine = gin.New()
	server.ginEngine.Use(ginlogger(logger))
	server.ginEngine.Use(gin.Recovery())
	return server
}

func (s *service) callMethod(mtype *methodType, params interface{}) (interface{}, error) {
	function := mtype.method.Func
	argIsValue := false
	var argv reflect.Value
	if mtype.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(mtype.ArgType.Elem())
	} else {
		argv = reflect.New(mtype.ArgType)
		argIsValue = true
	}

	if argIsValue {
		argv = argv.Elem()
	}
	if params != nil {
		tk := reflect.TypeOf(params).Kind()
		switch tk {
		case reflect.Slice:
			paramsArr, _ := params.([]interface{})
			if len(paramsArr) != argv.NumField() {
				return nil, NewRPCError(-32602, "Invalid params")
			}
			for i := 0; i < argv.NumField(); i++ {
				argv.Field(i).Set(reflect.ValueOf(paramsArr[i]))
			}
		case reflect.Map:
			paramsMap := params.(map[string]interface{})
			for i := 0; i < argv.NumField(); i++ {
				fieldInfo := argv.Type().Field(i) // a reflect.StructField
				tag := fieldInfo.Tag              // a reflect.StructTag
				name := tag.Get("json")
				if name == "" {
					name = strings.ToLower(fieldInfo.Name)
				}
				name = strings.Split(name, ",")[0] //json Tag
				if value, ok := paramsMap[name]; ok {
					if reflect.ValueOf(value).Type() == argv.FieldByName(fieldInfo.Name).Type() {
						argv.FieldByName(fieldInfo.Name).Set(reflect.ValueOf(value))
					}
				}
			}
		}
	}
	replyv := reflect.New(mtype.ReplyType.Elem())
	switch mtype.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(mtype.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(mtype.ReplyType.Elem(), 0, 0))
	}
	returnValues := function.Call([]reflect.Value{s.rcvr, argv, replyv})
	errInter := returnValues[0].Interface()
	if errInter != nil {
		e := errInter.(error)
		return nil, e
	}
	return replyv.Interface(), nil
}
func (server *RPCServer) Register(rcvr interface{}) error {
	return server.register(rcvr, "", false)
}

// RegisterName creates a service for the given receiver type under the given name and added it to the
// service collection this server provides to clients.
func (server *RPCServer) RegisterName(name string, rcvr interface{}) error {
	return server.register(rcvr, name, true)
}
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return token.IsExported(t.Name()) || t.PkgPath() == ""
}

var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

func suitableMethods(typ reflect.Type) map[string]*methodType {
	methods := make(map[string]*methodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		if method.PkgPath != "" {
			continue
		}
		if mtype.NumIn() != 3 {
			continue
		}
		argType := mtype.In(1)
		if !isExportedOrBuiltinType(argType) {
			continue
		}
		replyType := mtype.In(2)
		if replyType.Kind() != reflect.Ptr {
			continue
		}
		if mtype.NumOut() != 1 {
			continue
		}
		if returnType := mtype.Out(0); returnType != typeOfError {
			continue
		}
		methods[mname] = &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
	}
	return methods

}
func (server *RPCServer) register(rcvr interface{}, name string, useName bool) error {
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if useName {
		sname = name
	}
	if sname == "" {
		return fmt.Errorf("rpc.Register: no service name for type %s", s.typ.String())
	}
	if !token.IsExported(sname) && !useName {
		return fmt.Errorf("rpc.Register: type %s is not exported", sname)
	}
	s.name = sname
	s.methods = suitableMethods(s.typ)
	server.serviceMap[sname] = s

	return nil
}

func (server *RPCServer) getServiceAndMethodType(pack string) (*service, *methodType, error) {
	mpake := strings.Split(pack, ".")
	if len(mpake) != 2 {
		return nil, nil, methodNotFoundError
	}
	mService := server.serviceMap[mpake[0]]
	if mService == nil {
		return nil, nil, methodNotFoundError
	}
	mtype := mService.methods[mpake[1]]
	if mtype == nil {
		return nil, nil, methodNotFoundError
	}
	return mService, mtype, nil
}

func (server *RPCServer) parseJsonRPCObj(jsonObjMap map[string]interface{}, obj *jsonRPCObj) error {
	idNumber, ok := jsonObjMap["id"].(json.Number)
	if !ok {
		return invalidRequestError
	}
	idNumberStr := idNumber.String()
	id, err := strconv.Atoi(idNumberStr)
	if err != nil {
		return NewRPCError(-32600, err.Error())
	}
	obj.id = &id
	version, ok := jsonObjMap["jsonrpc"].(string)
	if !ok || version != "2.0" {
		return invalidRequestError
	}
	obj.jsonrpc = version
	methodPack, ok := jsonObjMap["method"].(string)
	if !ok {
		return invalidRequestError
	}
	obj.method = methodPack
	obj.params = jsonObjMap["params"]
	return nil
}

func (server *RPCServer) jsonRPCCall(data []byte, rpcId **int, w io.Writer) error {
	var personFromJSON interface{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	decoder.Decode(&personFromJSON)

	jsonObjMap := personFromJSON.(map[string]interface{})
	rpcObj := &jsonRPCObj{}
	if err := server.parseJsonRPCObj(jsonObjMap, rpcObj); err != nil {
		*rpcId = *&rpcObj.id
		return err
	}
	*rpcId = *&rpcObj.id
	s, t, err := server.getServiceAndMethodType(rpcObj.method)
	if err != nil {
		return err
	}
	rec, err := s.callMethod(t, rpcObj.params)
	if err != nil {
		return err
	}
	outMap := make(map[string]interface{})
	outMap["jsonrpc"] = jsonrpcVersion
	outMap["id"] = rpcObj.id
	outMap["result"] = rec
	outBytes, _ := json.Marshal(outMap)
	_, _ = w.Write(outBytes)
	return nil
}

func httperr(c *gin.Context, status int, err error) {
	c.String(status, "%s", err)
	c.Abort()
}

func writeRPCError(err error, reqId *int, w io.Writer) {
	rpcErr, isRPCErr := err.(*RPCError)
	e := jsonRPCRespErr{}
	if !isRPCErr {
		e.Code = -32603
		e.Message = "internal error"
	} else {
		e.Code = rpcErr.Code
		e.Message = rpcErr.Message
	}
	outMap := make(map[string]interface{})
	outMap["jsonrpc"] = jsonrpcVersion
	outMap["id"] = reqId
	outMap["error"] = e
	outBytes, _ := json.Marshal(outMap)
	_, _ = w.Write(outBytes)
}

func isWebsocketRequest(c *gin.Context) bool {
	connection := c.GetHeader("Connection")
	upgrade := c.GetHeader("Upgrade")
	return connection == "Upgrade" && upgrade == "websocket"
}
func (server *RPCServer) handleWebsocket(c *gin.Context) error {
	conn, err := server.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return err
	}
	for {
		t, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if t != websocket.TextMessage {
			continue
		}
		//msgText := string(msg)

		bs := bytes.NewBuffer(nil)
		var rpcId *int
		if err = server.jsonRPCCall(msg, &rpcId, bs); err != nil {
			writeRPCError(err, nil, bs)
		}
		if err = conn.WriteMessage(t, bs.Bytes()); err != nil {
			continue
		}
	}
	return nil
}

//Start starts rpc server.
func (server *RPCServer) Start() error {
	server.ginEngine.Any("/", func(c *gin.Context) {
		//handle websocket request
		if isWebsocketRequest(c) {
			if err := server.handleWebsocket(c); err != nil {
				server.logger.Warnf("ws connect err")
			}
			c.Abort()
			return
		}
		if "POST" != c.Request.Method {
			httperr(c, 404, errors.New("method not allowed"))
			return
		}
		contentType := c.ContentType()
		if contentType != "application/json" {
			httperr(c, 404, errors.New("not acceptable"))
			return
		}
		if nil == c.Request.Body {
			httperr(c, 404, errors.New("body not be empty"))
			return
		}
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			httperr(c, 500, fmt.Errorf("read body err: %s", err))
			return
		}
		c.Status(200)
		c.Header("Content-Type", "application/json; charset=utf-8")
		var rpcId *int = nil

		if err = server.jsonRPCCall(body, &rpcId, c.Writer); err != nil {
			writeRPCError(err, rpcId, c.Writer)
			return
		}
		c.Abort()
	})

	ln, err := net.Listen("tcp", server.config.ListenAddr)
	if err != nil {
		return fmt.Errorf("rpc server listen err: %s", err)
	}
	server.logger.Infof("rpc service listen on: %s", ln.Addr())
	return server.ginEngine.RunListener(ln)
}