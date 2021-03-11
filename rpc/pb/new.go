package rpcpb

import (
	proto "github.com/golang/protobuf/proto"
)

func NewResultInfo(Cid string, Error string, ArgsType string, result []byte) *ResultInfo {
	resultInfo := &ResultInfo{
		Cid:        *proto.String(Cid),
		Error:      *proto.String(Error),
		ResultType: *proto.String(ArgsType),
		Result:     result,
	}
	return resultInfo
}
