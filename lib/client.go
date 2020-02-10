package lib

import (
	"encoding/binary"
	"errors"
	"fg-test/pb"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io"
	"net"
	"time"
)

const (
	ID_TCPHEADER_LENGTH = 4            // 包头长度
	ID_DATA_MAX_LENGTH  = 10240 * 1024 // 消息体最大长度

	// 消息类型
	ID_MSG_TYPE_RECV = 1 // 数据包
)

func NewClient(c net.Conn) Client {
	return Client{Conn: c}
}

type Client struct {
	Conn    net.Conn
	Timeout time.Duration
}

func (c *Client) Send(id pb.MessageID, msg proto.Message) (int, error) {

	var (
		err  error
		body []byte
	)

	if body, err = proto.Marshal(msg); err != nil {
		return 0, err
	}

	data := make([]byte, 7+len(body))

	//组装header
	data[0] = ID_MSG_TYPE_RECV
	binary.BigEndian.PutUint16(data[1:3], uint16(id))
	binary.BigEndian.PutUint32(data[3:7], 3010)
	copy(data[7:], body)

	// 消息头+消息体
	dataLen := len(data)
	buf := make([]byte, ID_TCPHEADER_LENGTH+dataLen)
	binary.BigEndian.PutUint32(buf, uint32(dataLen))
	copy(buf[4:], data)

	retLen, err := c.Conn.Write(buf);
	if err != nil {
		return retLen, err
	}

	return retLen, err

}

func (c *Client) Receive() (int, error) {
	var err error

	//接收消息头,并解析成消息长度
	headbuf := make([]byte, ID_TCPHEADER_LENGTH)
	if _, err = io.ReadFull(c.Conn, headbuf); err != nil {
		fmt.Println("read buf err:", err.Error())
		return 0, err
	}

	dataLen := binary.BigEndian.Uint32(headbuf)
	if dataLen > ID_DATA_MAX_LENGTH || dataLen == 0 {
		return 0, errors.New("recv data length error")
	}

	databuf := make([]byte, dataLen)
	var revLen int
	if revLen, err = io.ReadFull(c.Conn, databuf); err != nil {
		return revLen, err
	}

	//接收的包的总长度
	retLenth := int(dataLen) + ID_TCPHEADER_LENGTH
	return retLenth, err

}
