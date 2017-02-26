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
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"github.com/liangdas/mqant/log"
)

const (
	Rserved = iota
	CONNECT  //1
	CONNACK	 //2

	PUBLISH  //3
	PUBACK   //4
	PUBREC   //5
	PUBREL   //6
	PUBCOMP  //7

	SUBSCRIBE //8
	SUBACK    //9

	UNSUBSCRIBE //10
	UNSUBACK    //11

	PINGREQ     //12
	PINGRESP    //13

	DISCONNECT  //14
)

var null_string = ""

type Pack struct {
	// Fixed header
	msg_type  byte
	dup_flag  byte
	qos_level byte
	retain    byte
	// Remaining Length
	length int

	// Variable header and playload
	variable interface{}
}

func (pack *Pack) GetVariable() interface{} {
	return pack.variable
}
func (pack *Pack) GetType() byte {
	return pack.msg_type
}
func (pack *Pack) SetType(typ byte) {
	pack.msg_type = typ
}
func (pack *Pack) GetDup() byte {
	return pack.dup_flag
}
func (pack *Pack) SetDup(dup byte) {
	pack.dup_flag = dup
}
func (pack *Pack) GetQos() byte {
	return pack.qos_level
}
func (pack *Pack) SetQos(qos byte) {
	pack.qos_level = qos
}

type Connect struct {
	protocol         *string
	version          byte
	keep_alive_timer int
	return_code      byte

	user_name     bool
	password      bool
	will_retain   bool
	will_qos      int
	will_flag     bool
	clean_session bool
	rserved       bool

	// Playload
	id         *string
	will_topic *string
	will_msg   *string
	uname      *string
	upassword  *string
}

func (c *Connect) GetUserName() *string {
	if !c.user_name {
		return &null_string
	} else {
		return c.uname
	}
}

func (c *Connect) GetPassword() *string {
	if !c.password {
		return &null_string
	} else {
		return c.upassword
	}
}

func (c *Connect) GetWillMsg() (bool, *string, *string) {
	if !c.will_flag {
		return false, &null_string, &null_string
	}
	return true, c.will_topic, c.will_msg
}

func (c *Connect) GetReturnCode() byte {
	return c.return_code
}

func (c *Connect) GetKeepAlive() int {
	return c.keep_alive_timer
}

func (c *Connect) IsCleanSession() bool {
	return c.clean_session
}

func (c *Connect) GetProtocol() *string {
	return c.protocol
}
func (c *Connect) GetVersion()( byte) {
	return c.version
}

type Connack struct {
	reserved    byte
	return_code byte
}

func (c *Connack) GetReturnCode() byte {
	return c.return_code
}
func (c *Connack) SetReturnCode(return_code byte) {
	c.return_code = return_code
}



type Publish struct {
	topic_name *string
	mid        int
	msg        []byte
}

func (pub *Publish) GetTopic() *string {
	return pub.topic_name
}
func (pub *Publish) SetTopic(topic *string) {
	pub.topic_name = topic
}
func (pub *Publish) GetMid() int {
	return pub.mid
}
func (pub *Publish) SetMid(id int) {
	pub.mid = id
}
func (pub *Publish) GetMsg() []byte {
	return pub.msg
}
func (pub *Publish) SetMsg(msg []byte) {
	pub.msg = msg
}

type Puback struct {
	mid int
}

func (ack *Puback) SetMid(id int) {
	ack.mid = id
}
func (ack *Puback) GetMid() int {
	return ack.mid
}

type Topics struct {
	name    *string
	Qos	byte
}

func (top *Topics) SetQos(Qos byte)  {
	top.Qos = Qos
}
func (top *Topics) GetQos() byte {
	return top.Qos
}
func (top *Topics) GetName() *string{
	return top.name
}

type Subscribe struct {
	mid     int
	topics	[]Topics
}

func (sub *Subscribe) SetMid(id int) {
	sub.mid = id
}
func (sub *Subscribe) GetMid() int {
	return sub.mid
}

func (sub *Subscribe) addTopics(top Topics){
	//n := len(sub.topics)
	//sub.topics = sub.topics[0 : n+1]
	//sub.topics[n] = top
	sub.topics=append(sub.topics,top)
}

func (sub *Subscribe) GetTopics() []Topics {
	return sub.topics
}

type Suback struct {
	mid     int
	Qos	byte	//0  2
}


type UNTopics struct {
	name    *string
}

func (top *UNTopics) GetName() *string{
	return top.name
}

type UNSubscribe struct {
	mid     int
	topics	[]Topics
}

func (sub *UNSubscribe) SetMid(id int) {
	sub.mid = id
}
func (sub *UNSubscribe) GetMid() int {
	return sub.mid
}

func (sub *UNSubscribe) addTopics(top Topics){
	//n := len(sub.topics)
	//sub.topics = sub.topics[0 : n+1]
	//sub.topics[n] = top
	sub.topics=append(sub.topics,top)
}

func (sub *UNSubscribe) GetTopics() []Topics {
	return sub.topics
}

type UNSuback struct {
	mid     int
}




// Parse the connect flags
func parse_flags(b byte, flag *Connect) {
	if b>>7 != 0 {
		flag.user_name = true
	}
	if b = b & 127; b>>6 != 0 {
		flag.password = true
	}
	if b = b & 63; b>>5 != 0 {
		flag.will_retain = true
	}
	b = b & 31
	flag.will_qos = int(b >> 3)
	if b = b & 7; b>>2 != 0 {
		flag.will_flag = true
	}
	if b = b & 3; b>>1 != 0 {
		flag.clean_session = true
	}
	if b&1 != 0 {
		flag.rserved = true
	}
}

// Read and Write a mqtt pack
func ReadPack(r *bufio.Reader) (pack *Pack, err error) {
	// Read the fixed header
	var (
		fixed     byte
		n         int
		temp_byte byte
		count_len = 1
	)
	fixed, err = r.ReadByte()
	if err != nil {
		return
	}
	// Parse the fixed header
	pack = new(Pack)
	pack.msg_type = fixed >> 4
	fixed = fixed & 15
	pack.dup_flag = fixed >> 3
	fixed = fixed & 7
	pack.qos_level = fixed >> 1
	pack.retain = fixed & 1
	// Get the length of the pack
	temp_byte, err = r.ReadByte()
	if err != nil {
		return
	}

	// Read the high
	multiplier := 1
	for {
		count_len++
		pack.length += (int(temp_byte&127) * multiplier)
		if temp_byte>>7 != 0 && count_len < 4 {
			temp_byte, err = r.ReadByte()
			if err != nil {
				return
			}
			multiplier *= 128
		} else {
			break
		}
	}
	// Read the Variable header and the playload
	// Check the msg type
	switch pack.msg_type {
	case CONNECT:
		// Read the protocol name
		var flags byte
		var conn = new(Connect)
		pack.variable = conn
		conn.protocol, n, err = readString(r)
		if err != nil {
			break
		}
		if n > (pack.length - 4) {
			err = fmt.Errorf("out of range:%v", pack.length-n)
			break
		}
		// Read the version
		conn.version, err = r.ReadByte()
		if err != nil {
			break
		}
		flags, err = r.ReadByte()
		if err != nil {
			break
		}
		// Read the keep alive timer
		conn.keep_alive_timer, err = readInt(r, 2)
		if err != nil {
			break
		}
		parse_flags(flags, conn)
		// Read the playload
		playload_len := pack.length - 2 - n - 4
		// Read the Client Identifier
		conn.id, n, err = readString(r)
		if err != nil {
			break
		}
		if n > 23 || n < 1 {
			err = fmt.Errorf("Identifier Rejected length is:%v", n)
			conn.return_code = 2
			break
		}
		playload_len -= n
		if n < 1 && (conn.will_flag || conn.password || n < 0) {
			err = fmt.Errorf("length error : %v", playload_len)
			break
		}
		if conn.will_flag {
			// Read the will topic and the will message
			conn.will_topic, n, err = readString(r)
			if err != nil {
				break
			}
			playload_len -= n
			if playload_len < 0 {
				err = fmt.Errorf("length error : %v", playload_len)
				break
			}
			conn.will_msg, n, err = readString(r)
			if err != nil {
				break
			}
			playload_len -= n
		}
		if conn.user_name && playload_len > 0 {
			conn.uname, n, err = readString(r)
			if err != nil {
				break
			}
			playload_len -= n
			if playload_len < 0 {
				err = fmt.Errorf("length error : %v", playload_len)
				break
			}
		}
		if conn.password && playload_len > 0 {
			conn.upassword, n, err = readString(r)
			if err != nil {
				break
			}
			playload_len -= n
			if playload_len < 0 {
				err = fmt.Errorf("length error : %v", playload_len)
				break
			}
		}
	case PUBLISH:
		pub := new(Publish)
		pack.variable = pub
		// Read the topic
		pub.topic_name, n, err = readString(r)
		if err != nil {
			break
		}
		vlen := pack.length - n- 2
		if n < 1 || vlen < 2 {
			err = fmt.Errorf("length error :%v", vlen)
			break
		}
		// Read the msg id
		pub.mid, err = readInt(r, 2)
		if err != nil {
			break
		}
		vlen -= 2
		// Read the playload
		pub.msg = make([]byte, vlen)
		_, err = io.ReadFull(r, pub.msg)
	case PUBACK:
		if pack.length == 2 {
			ack := new(Puback)
			ack.mid, err = readInt(r, 2)
			if err != nil {
				break
			}
			pack.variable = ack
		} else {
			err = fmt.Errorf("Pack(%v) length(%v) != 2", pack.msg_type, pack.length)
		}
	case PUBREL:
		if pack.length == 2 {
			ack := new(Puback)
			ack.mid, err = readInt(r, 2)
			if err != nil {
				break
			}
			pack.variable = ack
		} else {
			err = fmt.Errorf("Pack(%v) length(%v) != 2", pack.msg_type, pack.length)
		}
	case PUBREC:
		if pack.length == 2 {
			ack := new(Puback)
			ack.mid, err = readInt(r, 2)
			if err != nil {
				break
			}
			pack.variable = ack
		} else {
			err = fmt.Errorf("Pack(%v) length(%v) != 2", pack.msg_type, pack.length)
		}
	case PUBCOMP:
		if pack.length == 2 {
			ack := new(Puback)
			ack.mid, err = readInt(r, 2)
			if err != nil {
				break
			}
			pack.variable = ack
		} else {
			err = fmt.Errorf("Pack(%v) length(%v) != 2", pack.msg_type, pack.length)
		}
	case SUBSCRIBE:
		sub := new(Subscribe)
		sub.topics=make([]Topics,0)
		pack.variable = sub
		// Read the msg id
		sub.mid, err = readInt(r, 2)
		if err != nil {
			break
		}
		vlen := pack.length  //The length of the payload

		for vlen>3{	//一个Top至少大于 3字节
			// Read the topic list
			top:=new(Topics)
			nlen, err := readInt(r, 2)
			if err != nil {
				break
			}
			// Read the topic name
			buf := make([]byte, nlen)
			_, err = io.ReadFull(r, buf)
			str:=string(buf)
			top.name=&str


			// Read the topic name
			tQos, err := r.ReadByte()
			if err != nil {
				break
			}
			top.Qos=tQos
			vlen = vlen- 2-nlen-1
			sub.addTopics(*top)
		}
	case SUBACK:
		if pack.length == 3 {
			ack := new(Suback)
			ack.mid, err = readInt(r, 2)
			if err != nil {
				break
			}
			ack.Qos, err = r.ReadByte()
			if err != nil {
				break
			}
			pack.variable = ack
		} else {
			err = fmt.Errorf("Pack(%v) length(%v) != 3", pack.msg_type, pack.length)
		}
	case UNSUBSCRIBE:
		sub := new(UNSubscribe)
		sub.topics=make([]Topics,0)
		pack.variable = sub
		// Read the msg id
		sub.mid, err = readInt(r, 2)
		if err != nil {
			break
		}
		vlen := pack.length  //The length of the payload

		for vlen>3{	//一个Top至少大于 3字节
			// Read the topic list
			top:=new(Topics)
			nlen, err := readInt(r, 2)
			if err != nil {
				break
			}
			// Read the topic name
			buf := make([]byte, nlen)
			_, err = io.ReadFull(r, buf)
			str:=string(buf)
			top.name=&str

			vlen = vlen- 2-nlen
			sub.addTopics(*top)
		}
	case UNSUBACK:
		if pack.length == 2 {
			ack := new(UNSuback)
			ack.mid, err = readInt(r, 2)
			if err != nil {
				break
			}
			pack.variable = ack
		} else {
			err = fmt.Errorf("Pack(%v) length(%v) != 1", pack.msg_type, pack.length)
		}
	case PINGREQ:
	// Pass
	// Nothing to do
	case DISCONNECT:
	// Pass, nothing to do.
	default:
		//将pack剩余中的数据读了
		log.Error("No Find Pack(%v) length(%v)", pack.msg_type, pack.length)
		if pack.length>0{
			buf:= make([]byte, pack.length)
			_, err = io.ReadFull(r, buf)
		}

	}

	return
}

func readString(r *bufio.Reader) (s *string, nn int, err error) {
	temp_string := ""
	s = &temp_string
	nn, err = readInt(r, 2)
	if err != nil {
		return
	}
	if nn > 0 {
		buf := make([]byte, nn)
		_, err = io.ReadFull(r, buf)
		if err == nil {
			*s = string(buf)
		}
	} else {
		*s = ""
	}
	return
}

func readInt(r *bufio.Reader, length int) (int, error) {
	buf := make([]byte, 4)
	_, err := io.ReadFull(r, buf[:length])
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint16(buf[:length])), nil
}

func WritePack(pack *Pack, w *bufio.Writer) error {
	if err := DelayWritePack(pack, w); err != nil {
		return err
	}
	return w.Flush()
}

func DelayWritePack(pack *Pack, w *bufio.Writer) (err error) {
	// Write the fixed header
	var fixed byte
	// Byte 1
	fixed = pack.msg_type << 4
	fixed |= (pack.dup_flag << 3)
	fixed |= (pack.qos_level << 1)
	if err = w.WriteByte(fixed); err != nil {
		return
	}
	// Byte2
	switch pack.msg_type {
	case CONNACK:
		ack := pack.variable.(*Connack)
		if err = w.WriteByte(getRemainingLength(2)[0]); err != nil {
			return
		}
		// Write the variable
		if err = writeFull(w, []byte{ack.reserved, ack.return_code}); err != nil {
			return
		}
	case PUBLISH:
		// Publish the msg to the client
		pub := pack.variable.(*Publish)
		if err = writeFull(w, getRemainingLength(4+len([]byte(*pub.topic_name))+len(pub.msg))); err != nil {
			return
		}
		if err = writeString(w, pub.topic_name); err != nil {
			return
		}
		if err = writeInt(w, pub.mid, 2); err != nil {
			return
		}
		if err = writeFull(w, pub.msg); err != nil {
			return
		}
	case PUBACK:
		ack := pack.variable.(*Puback)
		if err = w.WriteByte(getRemainingLength(2)[0]); err != nil {
			return
		}
		// Write the variable
		if err = writeInt(w, ack.mid,2); err != nil {
			return
		}
	case PUBREC:
		ack := pack.variable.(*Puback)
		if err = w.WriteByte(getRemainingLength(2)[0]); err != nil {
			return
		}
		// Write the variable
		if err = writeInt(w, ack.mid,2); err != nil {
			return
		}
	case PUBREL:
		ack := pack.variable.(*Puback)
		if err = w.WriteByte(getRemainingLength(2)[0]); err != nil {
			return
		}
		// Write the variable
		if err = writeInt(w, ack.mid,2); err != nil {
			return
		}
	case PUBCOMP:
		ack := pack.variable.(*Puback)
		if err = w.WriteByte(getRemainingLength(2)[0]); err != nil {
			return
		}
		// Write the variable
		if err = writeInt(w, ack.mid,2); err != nil {
			return
		}
	case SUBSCRIBE:
		// Subscribe the msg to the client
		sub := pack.variable.(*Subscribe)
		tnum:=0
		for _,top :=range sub.topics{
			tnum=tnum+2+len([]byte(*top.name))+1 //Qos
		}
		//The length of the payload. It can be a multibyte field.
		if err = w.WriteByte(getRemainingLength(tnum)[0]); err != nil {
			return
		}
		if err = writeInt(w, sub.mid,2); err != nil {
			return
		}
		for _,top :=range sub.topics{
			buf:=[]byte(*top.name)
			if err = writeInt(w, len(buf),2); err != nil {
				return
			}
			if err = writeFull(w,buf); err != nil {
				return
			}
			if err = w.WriteByte(top.Qos); err != nil {
				return
			}
		}
	case SUBACK:
		ack := pack.variable.(*Suback)
		if err = w.WriteByte(getRemainingLength(3)[0]); err != nil {
			return
		}
		// Write the variable
		if err = writeInt(w, ack.mid,2); err != nil {
			return
		}
		if err = w.WriteByte(ack.Qos); err != nil {
			return
		}
	case UNSUBSCRIBE:
		// Subscribe the msg to the client
		sub := pack.variable.(*UNSubscribe)
		tnum:=0
		for _,top :=range sub.topics{
			tnum=tnum+2+len([]byte(*top.name)) //
		}
		//The length of the payload. It can be a multibyte field.
		if err = w.WriteByte(getRemainingLength(tnum)[0]); err != nil {
			return
		}
		if err = writeInt(w, sub.mid,2); err != nil {
			return
		}
		for _,top :=range sub.topics{
			buf:=[]byte(*top.name)
			if err = writeInt(w, len(buf),2); err != nil {
				return
			}
			if err = writeFull(w,buf); err != nil {
				return
			}
		}
	case UNSUBACK:
		ack := pack.variable.(*UNSuback)
		if err = w.WriteByte(getRemainingLength(2)[0]); err != nil {
			return
		}
		// Write the variable
		if err = writeInt(w, ack.mid,2); err != nil {
			return
		}
	case PINGRESP:
		err = w.WriteByte(0)
	}
	return
}

func getRemainingLength(length int) []byte {
	b := make([]byte, 4)
	count := 0
	for {
		digit := length % 128
		length = length / 128
		if length > 0 {
			digit |= 128
			b[count] = byte(digit)
		} else {
			b[count] = byte(digit)
			break
		}
		count++
	}
	return b[:count+1]
}

func writeString(w *bufio.Writer, s *string) error {
	// Write the length of the string
	if s == nil {
		return errors.New("nil pointer")
	}
	data := []byte(*s)
	// Write the string length
	err := writeInt(w, len(data), 2)
	if err != nil {
		return err
	}
	return writeFull(w, data)
}
func writeInt(w *bufio.Writer, i, size int) error {
	b := make([]byte, size)
	binary.BigEndian.PutUint16(b, uint16(i))
	return writeFull(w, b)
}

// wirteFull write the data into the Writer's buffer
func writeFull(w *bufio.Writer, b []byte) (err error) {
	hasRead, n := 0, 0
	for n < len(b) {
		n, err = w.Write(b[hasRead:])
		if err != nil {
			break
		}
		hasRead += n
	}
	return err
}

// Pack setters

// Get a connection ack pack
func GetConnAckPack(return_code byte) *Pack {
	pack := new(Pack)
	pack.SetType(CONNACK)
	ack := new(Connack)
	pack.variable = ack
	ack.SetReturnCode(return_code)
	return pack
}

// Get a publis pack
func GetPubPack(qos byte, dup byte, mid int, topic *string, msg []byte) *Pack {
	pack := new(Pack)
	pack.SetQos(qos)
	pack.SetDup(dup)
	pack.SetType(PUBLISH)

	pub := new(Publish)
	pub.SetMid(mid)
	pub.SetTopic(topic)
	pub.SetMsg(msg)
	pack.variable = pub
	return pack
}

// Get a connection ack pack
func GetPubAckPack(mid int) *Pack {
	pack := new(Pack)
	pack.SetType(PUBACK)
	ack := new(Puback)
	ack.SetMid(mid)
	pack.variable = ack
	return pack
}
// Get a connection ack pack
func GetPubRECPack(mid int) *Pack {
	pack := new(Pack)
	pack.SetType(PUBREC)
	ack := new(Puback)
	ack.SetMid(mid)
	pack.variable = ack
	return pack
}
// Get a connection ack pack
func GetPubRELPack(mid int) *Pack {
	pack := new(Pack)
	pack.SetType(PUBREL)
	pack.SetQos(1)
	pack.SetDup(0)
	ack := new(Puback)
	ack.SetMid(mid)
	pack.variable = ack
	return pack
}
// Get a connection ack pack
func GetPubCOMPPack(mid int) *Pack {
	pack := new(Pack)
	pack.SetType(PUBCOMP)
	ack := new(Puback)
	ack.SetMid(mid)
	pack.variable = ack
	return pack
}
// Get a connection ack pack
func GetSubAckPack(mid int) *Pack {
	pack := new(Pack)
	pack.SetType(SUBACK)
	ack := new(Suback)
	ack.mid=mid
	ack.Qos=0
	pack.variable = ack
	return pack
}
func GetUNSubAckPack(mid int) *Pack {
	pack := new(Pack)
	pack.SetType(UNSUBACK)
	ack := new(UNSuback)
	ack.mid=mid
	pack.variable = ack
	return pack
}
// Get a request for ping pack
func GetPingResp(qos byte, dup byte) *Pack {
	pack := new(Pack)
	pack.SetQos(qos)
	pack.SetDup(dup)
	pack.SetType(PINGRESP)
	return pack
}
