/**
 * @description:
 * @author: jarekzha@gmail.com
 * @date: 2018/5/9
 */
package utils

import (
	"github.com/golang/protobuf/proto"
	"sync"
)

var (
	once                sync.Once
	protoBufferPoolInst *sync.Pool
)

// 获取实例
func ProtoBufferPoolInst() *sync.Pool {
	if protoBufferPoolInst == nil {
		once.Do(func() {
			protoBufferPoolInst = &sync.Pool{
				New: func() interface{} {
					return proto.NewBuffer(nil)
				},
			}
		})
	}

	return protoBufferPoolInst
}

// 获取一个新的
func GetProtoBuffer() *proto.Buffer {
	inst := ProtoBufferPoolInst()
	buffer := inst.Get().(*proto.Buffer)

	return buffer
}

// 放回池子
func PutProtoBuffer(buffer *proto.Buffer) {
	inst := ProtoBufferPoolInst()
	buffer.Reset()
	inst.Put(buffer)
}
