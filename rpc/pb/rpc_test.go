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
package rpcpb
import (
	"github.com/golang/protobuf/proto"
	"testing"
	"time"
)
func TestRPCInfo(t *testing.T) {
	rpc := &RPCInfo{        // 使用辅助函数设置域的值
		Cid: *proto.String("123457"),
		Fn:  *proto.String("hello"),
		Expired:  *proto.Int64(time.Now().UnixNano() / 1000000),
		Reply:  *proto.Bool(true),
		ReplyTo:  *proto.String("232244"),
	}    // 进行编码
	rpc.ArgsType=[]string{"s","s"}
	rpc.Args=[][]byte{[]byte("hello"),[]byte("world")}
	data, err := proto.Marshal(rpc)
	if err != nil {
		t.Fatalf("marshaling error: ", err)
	}    // 进行解码
	newRPC := &RPCInfo{}
	err = proto.Unmarshal(data, newRPC)
	if err != nil {
		t.Fatalf("unmarshaling error: ", err)
	}    // 测试结果
	if rpc.ReplyTo != newRPC.GetReplyTo() {
		t.Fatalf("data mismatch %q != %q", rpc.GetReplyTo(), newRPC.GetReplyTo())
	}
}

func TestResultInfo(t *testing.T) {
	result := &ResultInfo{        // 使用辅助函数设置域的值
		Cid: *proto.String("123457"),
		Error:  *proto.String("hello"),
		ResultType:  *proto.String("s"),
		Result:  []byte("232244"),
	}    // 进行编码
	data, err := proto.Marshal(result)
	if err != nil {
		t.Fatalf("marshaling error: ", err)
	}    // 进行解码
	newResult := &ResultInfo{}
	err = proto.Unmarshal(data, newResult)
	if err != nil {
		t.Fatalf("unmarshaling error: ", err)
	}    // 测试结果
	if result.Cid != newResult.GetCid() {
		t.Fatalf("data mismatch %q != %q", result.GetCid(), newResult.GetCid())
	}
}
