// Copyright 2014 mqant Author. All Rights Reserved.
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
package mqtt

import (
	"bufio"
	"fmt"
	"net"
	"testing"
)

func TestConnet(t *testing.T) {
	// Set the listener
	l, err := net.Listen("tcp", ":9001")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := l.Accept()
	if err != nil {
		t.Fatal(err)
	}
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	pack, err := ReadPack(r)
	if err != nil {
		t.Fatal(err)
	}
	// Check the pack
	c := pack.variable.(*Connect)
	fmt.Println(*c.uname)
	fmt.Println(*c.upassword)
	fmt.Println(*c.id)

	// Return the connection ack
	pack = new(Pack)
	pack.msg_type = CONNACK

	ack := new(Connack)
	ack.return_code = 0
	pack.variable = ack

	if err := WritePack(pack, w); err != nil {
		t.Error(err)
		return
	}

	pack = new(Pack)
	pack.qos_level = 1
	pack.dup_flag = 0
	pack.msg_type = PUBLISH
	pub := new(Publish)
	pub.mid = 1
	s := "chat"
	pub.topic_name = &s
	pub.msg = []byte("Hello push server")
	pack.variable = pub
	if err := WritePack(pack, w); err != nil {
		t.Error(err)
		return
	}

	if _, err := ReadPack(r); err != nil {
		t.Error(err)
	}
}
