package client

import (
	"bytes"
	"redis-go/lib/logger"
	"redis-go/resp/reply"
	"strconv"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	logger.Setup(&logger.Settings{
		Path:       "logs",
		Name:       "redis-go",
		Ext:        ".log",
		TimeFormat: "2006-01-02",
	})
	client, err := MakeClient("localhost:6379")
	if err != nil {
		t.Error(err)
	}
	client.Start()

	result := client.Send([][]byte{
		[]byte("PING"),
	})
	if statusRet, ok := result.(*reply.StatusReply); ok {
		if statusRet.Status != "PONG" {
			t.Error("`ping` failed, result: " + statusRet.Status)
		}
	}

	result = client.Send([][]byte{
		[]byte("SET"),
		[]byte("a"),
		[]byte("a"),
	})
	if statusRet, ok := result.(*reply.StatusReply); ok {
		if statusRet.Status != "OK" {
			t.Error("`set` failed, result: " + statusRet.Status)
		}
	}

	result = client.Send([][]byte{
		[]byte("GET"),
		[]byte("a"),
	})
	if bulkRet, ok := result.(*reply.BulkReply); ok {
		if string(bulkRet.Arg) != "a" {
			t.Error("`get` failed, result: " + string(bulkRet.Arg))
		}
	}

	result = client.Send([][]byte{
		[]byte("DEL"),
		[]byte("a"),
	})
	if intRet, ok := result.(*reply.IntReply); ok {
		if intRet.Code != 1 {
			t.Error("`del` failed, result: " + strconv.FormatInt(intRet.Code, 10))
		}
	}

	client.doHeartbeat() // random do heartbeat
	result = client.Send([][]byte{
		[]byte("GET"),
		[]byte("a"),
	})
	if _, ok := result.(*reply.NullBulkReply); !ok {
		t.Error("`get` failed, result: " + string(result.ToBytes()))
	}

	result = client.Send([][]byte{
		[]byte("DEL"),
		[]byte("arr"),
	})

	client.Close()
}

func TestReconnect(t *testing.T) {
	logger.Setup(&logger.Settings{
		Path:       "logs",
		Name:       "godis",
		Ext:        ".log",
		TimeFormat: "2006-01-02",
	})
	client, err := MakeClient("localhost:6379")
	if err != nil {
		t.Error(err)
	}
	client.Start()

	_ = client.conn.Close()
	time.Sleep(time.Second) // wait for reconnecting
	success := false
	for i := 0; i < 3; i++ {
		result := client.Send([][]byte{
			[]byte("PING"),
		})
		if bytes.Equal(result.ToBytes(), []byte("+PONG\r\n")) {
			success = true
			break
		}
	}
	if !success {
		t.Error("reconnect error")
	}
}
