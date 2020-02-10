package lib

import (
	"fg-test/pb"
	"github.com/golang/protobuf/proto"
	"strconv"
)

var (
	uidStart uint32
)

func init() {
	uidStart = 10000
}

type Robot struct {
	UserID uint32
	//消息ID，入登录登出等
	MessageID pb.MessageID
	//协议ID
	ProtoNum int32
}

func (r *Robot) Request() (error, proto.Message) {
	//uid := atomic.AddUint32(&uidStart, 1)
	uidStr := strconv.FormatUint(uint64(r.UserID), 10)
	loginReq := &pb.CSLoginReq{
		Uid:      r.UserID,
		Account:  "test_" + uidStr,
		ProtoNum: r.ProtoNum,
		//ProtoNum: 50159572,
	}
	return nil, loginReq
}
