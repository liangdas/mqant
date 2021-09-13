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
	"fmt"
	"google.golang.org/protobuf/proto"
	"sync"
	"testing"
)

func TestSession(t *testing.T) {
	session := &SessionImp{ // 使用辅助函数设置域的值
		IP:      *proto.String("127.0.0.1"),
		Network: *proto.String("tcp"),
	} // 进行编码
	session.Settings = map[string]string{"isLogin": "true"}
	data, err := proto.Marshal(session)
	if err != nil {
		t.Fatalf("marshaling error: %v", err)
	} // 进行解码
	newSession := &SessionImp{}
	err = proto.Unmarshal(data, newSession)
	if err != nil {
		t.Fatalf("unmarshaling error: %v", err)
	} // 测试结果
	if newSession.GetSettings() == nil {
		t.Fatalf("data mismatch Settings == nil")
	} else {
		if newSession.GetSettings()["isLogin"] != "true" {
			t.Fatalf("data mismatch %q != %q", session.GetSettings()["isLogin"], newSession.GetSettings()["isLogin"])
		}
	}

}

func TestSessionagent_Serializable(t *testing.T) {
	session, err := NewSessionByMap(nil, map[string]interface{}{
		"IP": "IP",
	})
	if err != nil {
		t.Fatalf("NewSessionByMap error: %v", err)
	}
	settings := map[string]string{"a": "a"}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { //開一個協程寫map
		for j := 0; j < 1000000; j++ {
			_session := session.Clone()
			session.Serializable()
			session.SetLocalKV("ff", "sss")
			session.ImportSettings(settings)
			_session.Set("TestTopic", fmt.Sprintf("set %v", j))
			_session.SetTopic("ttt")
			_session.Serializable()
			a, ok := session.Load("a")
			if a != "a" || ok != true {
				t.Fatalf("Load error: %v", err)
			}
			cs := session.CloneSettings()
			for k, v := range settings {
				if _, ok := cs[k]; ok {
					//不用替换
				} else {
					cs[k] = v
				}
			}
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() { //開一個協程讀map
		for j := 0; j < 1000000; j++ {
			session.Clone()
			session.Serializable()
			session.SetLocalKV("ff", "sss")
			session.ImportSettings(settings)
			session.SetTopic("ttt")
			//fmt.Println("Serializable", b)
		}
		wg.Done()
	}()
	wg.Wait()
}
