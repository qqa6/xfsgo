package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/token"
	"io/ioutil"
	"log"
	"net/http"
	"net/rpc"
	"reflect"
	"strings"
	configs "xblockchain/cmd/config"
	errors2 "xblockchain/rpc/errors"
)

var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

type ServerStarter struct {
	rpcserver  *rpc.Server
	serviceMap map[string]*service
}

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
	numCalls  uint
}

type service struct {
	name    string                 // name of service
	rcvr    reflect.Value          // receiver of methods for the service
	typ     reflect.Type           // type of the receiver
	methods map[string]*methodType // registered methods
}

func (s *service) Call(methodName string, params interface{}) (interface{}, error) {
	mtype := s.methods[methodName]
	if mtype == nil {
		return nil, errors2.New(-32601, "Method not found")
	}
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
				return nil, errors2.New(-32602, "Invalid params")
			}
			for i := 0; i < argv.NumField(); i++ {
				argv.Field(i).Set(reflect.ValueOf(paramsArr[i]))
			}
		case reflect.Map:
			paramsMap := params.(map[string]interface{})
			if len(paramsMap) != argv.NumField() {
				return nil, errors2.New(-32602, "Invalid params")
			}
			for k, v := range paramsMap {
				argvv := argv.FieldByName(k)
				if !argvv.IsValid() {
					continue
				}
				argvv.Set(reflect.ValueOf(v))
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
		e := errInter.(*errors2.JsonRPCError)
		return nil, e
	}
	return replyv.Interface(), nil
}

func NewServerStarter() (*ServerStarter, error) {
	return &ServerStarter{
		serviceMap: make(map[string]*service),
	}, nil
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return token.IsExported(t.Name()) || t.PkgPath() == ""
}
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
		if !isExportedOrBuiltinType(replyType) {
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
func (server *ServerStarter) Register(rcvr interface{}) error {
	return server.register(rcvr, "", false)
}

func (server *ServerStarter) RegisterName(name string, rcvr interface{}) error {
	return server.register(rcvr, name, true)
}

func (server *ServerStarter) register(rcvr interface{}, name string, useName bool) error {
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if useName {
		sname = name
	}
	if sname == "" {
		s := "rpc.Register: no service name for type " + s.typ.String()
		log.Print(s)
		return errors.New(s)
	}
	if !token.IsExported(sname) && !useName {
		s := "rpc.Register: type " + sname + " is not exported"
		log.Print(s)
		return errors.New(s)
	}
	s.name = sname
	s.methods = suitableMethods(s.typ)
	server.serviceMap[sname] = s

	return nil
}

func PerParseRequestData(c *http.Request) (map[string]interface{}, error) {
	if "POST" != c.Method {
		return nil, errors.New("POST method excepted")
	}
	if nil == c.Body {
		return nil, fmt.Errorf("no POST data")
	}
	body, err := ioutil.ReadAll(c.Body)
	if err != nil {
		return nil, fmt.Errorf("errors while reading request body")
	}
	var data = make(map[string]interface{})
	decoder := json.NewDecoder(bytes.NewBuffer(body))
	decoder.UseNumber()
	err = decoder.Decode(&data)
	if nil != err {
		return nil, fmt.Errorf("errors parsing json request")
	}
	return data, nil
}

func (server *ServerStarter) HandleJsonRPCRequest(data map[string]interface{}) (*int, interface{}, error) {

	if len(data) == 0 || data != nil {
		return nil, nil, errors2.New(-32700, "Parse error")
	}
	idNumber, ok := data["id"].(json.Number)
	if !ok {
		return nil, nil, errors2.New(-32600, "Invalid Request")
	}
	id64, err := idNumber.Int64()
	if err != nil {
		return nil, nil, errors2.New(-32600, "Invalid Request")
	}
	id := int(id64)
	if data["jsonrpc"] != "2.0" {
		return &id, nil, errors2.New(-32600, "Invalid Request")
	}
	method, ok := data["method"].(string)
	mpake := strings.Split(method, ".")
	if !ok || len(mpake) != 2 {
		return &id, nil, errors2.New(-32601, "Method not found")
	}
	params := data["params"]
	service := server.serviceMap[mpake[0]]
	if service == nil {
		return &id, nil, errors2.New(-32601, "Method not found")
	}
	result, err := service.Call(mpake[1], params)
	if err != nil {
		return &id, nil, err
	}
	return &id, result, nil
}

type RpcServer struct {
	common *ServerStarter
}

func (entity RpcServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resJson := make(map[string]interface{})
	data, err := PerParseRequestData(r)

	resJson["jsonrpc"] = "2.0"
	if err != nil {
		resJson["id"] = nil
		resJson["error"] = err
		jsonInfo, _ := json.Marshal(resJson)
		w.Write(jsonInfo)
		return
	}
	id, datas, err := entity.common.HandleJsonRPCRequest(data)

	resJson["id"] = id
	if err != nil {
		resJson["error"] = err
		jsonInfo, _ := json.Marshal(resJson)
		w.Write(jsonInfo)
		return
	}
	resJson["result"] = datas
	jsonInfo, _ := json.Marshal(resJson)
	w.Write(jsonInfo)

}
func (server *ServerStarter) Run() error {

	tempConfig := configs.GetConfig()
	var err error
	Path := tempConfig.Network.ProtocolType + tempConfig.Network.RPCListenAddress

	Common := RpcServer{common: server}
	switch tempConfig.Network.ProtocolType {
	case "ws://":
		ws := NewWsServer(Path, server)
		ws.Start()
	case "http://":
		http.Handle("/", Common)
		err = http.ListenAndServe(Path, nil)
	case "https://":
		http.Handle("/", Common)
		http.ListenAndServeTLS(Path, tempConfig.Network.ServerCrt, tempConfig.Network.ServerKey, nil)
	}
	return err
}
