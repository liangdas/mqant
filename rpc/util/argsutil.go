// Copyright 2014 loolgame Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package argsutil

import (
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/proto"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/utils"
	"reflect"
	"strings"
)

var (
	NULL    = "null"    //nil   null
	BOOL    = "bool"    //bool
	INT     = "int"     //int
	LONG    = "long"    //long64
	FLOAT   = "float"   //float32
	DOUBLE  = "double"  //float64
	BYTES   = "bytes"   //[]byte
	STRING  = "string"  //string
	MAP     = "map"     //map[string]interface{}
	MAPSTR  = "mapstr"  //map[string]string{}
	TRACE   = "trace"   //log.TraceSpanImp
	Marshal = "marshal" //mqrpc.Marshaler
	Proto   = "proto"   //proto.Message
)

func ArgsTypeAnd2Bytes(app module.App, arg interface{}) (string, []byte, error) {
	if arg == nil {
		return NULL, nil, nil
	}
	switch v2 := arg.(type) {
	case []uint8:
		return BYTES, v2, nil
	}
	switch v2 := arg.(type) {
	case nil:
		return NULL, nil, nil
	case string:
		return STRING, []byte(v2), nil
	case bool:
		return BOOL, mqanttools.BoolToBytes(v2), nil
	case int32:
		return INT, mqanttools.Int32ToBytes(v2), nil
	case int64:
		return LONG, mqanttools.Int64ToBytes(v2), nil
	case float32:
		return FLOAT, mqanttools.Float32ToBytes(v2), nil
	case float64:
		return DOUBLE, mqanttools.Float64ToBytes(v2), nil
	case []byte:
		return BYTES, v2, nil
	case map[string]interface{}:
		bytes, err := mqanttools.MapToBytes(v2)
		if err != nil {
			return MAP, nil, err
		}
		return MAP, bytes, nil
	case map[string]string:
		bytes, err := mqanttools.MapToBytesString(v2)
		if err != nil {
			return MAPSTR, nil, err
		}
		return MAPSTR, bytes, nil
	case log.TraceSpanImp:
		bytes, err := json.Marshal(v2)
		if err != nil {
			return TRACE, nil, err
		}
		return TRACE, bytes, nil
	case *log.TraceSpanImp:
		bytes, err := json.Marshal(v2)
		if err != nil {
			return TRACE, nil, err
		}
		return TRACE, bytes, nil
	default:
		for _, v := range app.GetRPCSerialize() {
			ptype, vk, err := v.Serialize(arg)
			if err == nil {
				//解析成功了
				return ptype, vk, err
			}
		}

		rv := reflect.ValueOf(arg)
		if rv.Kind() != reflect.Ptr {
			//不是指针
			return "", nil, fmt.Errorf("Args2Bytes [%v] not registered to app.addrpcserialize(...) structure type or not *mqrpc.marshaler pointer type", reflect.TypeOf(arg))
		} else {
			if rv.IsNil() {
				//如果是nil则直接返回
				return NULL, nil, nil
			}
			if v2, ok := arg.(mqrpc.Marshaler); ok {
				b, err := v2.Marshal()
				if err != nil {
					return "", nil, fmt.Errorf("args [%s] marshal error %v", reflect.TypeOf(arg), err)
				}
				if v2.String() != "" {
					return fmt.Sprintf("%v@%v", Marshal, v2.String()), b, nil
				} else {
					return fmt.Sprintf("%v@%v", Marshal, reflect.TypeOf(arg)), b, nil
				}
			}
			if v2, ok := arg.(proto.Message); ok {
				b, err := proto.Marshal(v2)
				if err != nil {
					log.Error("proto.Marshal error")
					return "", nil, fmt.Errorf("args [%s] proto.Marshal error %v", reflect.TypeOf(arg), err)
				}
				return fmt.Sprintf("%v@%v", Proto, reflect.TypeOf(arg)), b, nil
			}
		}

		return "", nil, fmt.Errorf("Args2Bytes [%s] not registered to app.addrpcserialize(...) structure type", reflect.TypeOf(arg))
	}
}

func Bytes2Args(app module.App, argsType string, args []byte) (interface{}, error) {
	if strings.HasPrefix(argsType, Marshal) {
		return args, nil
	}
	if strings.HasPrefix(argsType, Proto) {
		return args, nil
	}
	switch argsType {
	case NULL:
		return nil, nil
	case STRING:
		return string(args), nil
	case BOOL:
		return mqanttools.BytesToBool(args), nil
	case INT:
		return mqanttools.BytesToInt32(args), nil
	case LONG:
		return mqanttools.BytesToInt64(args), nil
	case FLOAT:
		return mqanttools.BytesToFloat32(args), nil
	case DOUBLE:
		return mqanttools.BytesToFloat64(args), nil
	case BYTES:
		return args, nil
	case MAP:
		mps, errs := mqanttools.BytesToMap(args)
		if errs != nil {
			return nil, errs
		}
		return mps, nil
	case MAPSTR:
		mps, errs := mqanttools.BytesToMapString(args)
		if errs != nil {
			return nil, errs
		}
		return mps, nil
	case TRACE:
		trace := &log.TraceSpanImp{}
		err := json.Unmarshal(args, trace)
		if err != nil {
			return nil, err
		}
		return trace.ExtractSpan(), nil
	default:
		for _, v := range app.GetRPCSerialize() {
			vk, err := v.Deserialize(argsType, args)
			if err == nil {
				//解析成功了
				return vk, err
			}
		}
		return nil, fmt.Errorf("Bytes2Args [%s] not registered to app.addrpcserialize(...)", argsType)
	}
}
