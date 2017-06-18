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
package basegate
import (
	"github.com/golang/protobuf/proto"
	"testing"
)
func TestSession(t *testing.T) {
	session := &session{        // 使用辅助函数设置域的值
		IP: *proto.String("127.0.0.1"),
		Network:  *proto.String("tcp"),
		Sessionid:  *proto.String("iii"),
		Serverid:  *proto.String("232244"),
	}    // 进行编码
	session.Settings=map[string]string{"isLogin":"true"}
	data, err := proto.Marshal(session)
	if err != nil {
		t.Fatalf("marshaling error: ", err)
	}    // 进行解码
	newSession := &session{}
	err = proto.Unmarshal(data, newSession)
	if err != nil {
		t.Fatalf("unmarshaling error: ", err)
	}    // 测试结果
	if session.Serverid != newSession.GetServerid() {
		t.Fatalf("data mismatch %q != %q", session.GetServerid(), newSession.GetServerid())
	}
	if newSession.GetSettings()==nil{
		t.Fatalf("data mismatch Settings == nil")
	}else{
		if newSession.GetSettings()["isLogin"]!="true"{
			t.Fatalf("data mismatch %q != %q", session.GetSettings()["isLogin"], newSession.GetSettings()["isLogin"])
		}
	}

}
