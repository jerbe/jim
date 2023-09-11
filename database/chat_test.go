package database

import (
	"encoding/json"
	"log"
	"testing"
	"time"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/15 01:16
  @describe :
*/

func TestAddMessage(t *testing.T) {
	type args struct {
		msg *ChatMessage
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "添加成功",
			args: args{
				msg: &ChatMessage{
					RoomID:      "00000001_00000002",
					Type:        1,
					SessionType: 1,
					SenderID:    1,
					ReceiverID:  2,
					SendStatus:  1,
					ReadStatus:  0,
					Status:      1,
					Body: ChatMessageBody{
						Text:          "纯文本消息",
						Src:           "",
						Format:        "",
						Size:          "",
						Longitude:     "",
						Latitude:      "",
						Scale:         0,
						LocationLabel: "",
					},
					CreatedAt: time.Now().UnixMilli(),
					UpdatedAt: time.Now().UnixNano(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := AddChatMessage(tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("AddChatMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkAddMessageWithTransaction(b *testing.B) {

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			msg := &ChatMessage{
				RoomID:      "0000000100000002",
				Type:        1,
				SessionType: 1,
				SenderID:    1,
				ReceiverID:  2,
				SendStatus:  1,
				ReadStatus:  0,
				Status:      1,
				Body: ChatMessageBody{
					Text:          "纯文本消息",
					Src:           "",
					Format:        "",
					Size:          "",
					Longitude:     "",
					Latitude:      "",
					Scale:         0,
					LocationLabel: "",
				},
				CreatedAt: time.Now().UnixMilli(),
				UpdatedAt: time.Now().UnixMilli(),
			}
			err := AddChatMessageTx(msg)
			if err != nil {
				log.Println(err)
			}
		}
	})
}

func BenchmarkAddMessage(b *testing.B) {

	b.SetParallelism(1000)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			msg := &ChatMessage{
				RoomID:      "0000000100000002",
				Type:        1,
				SessionType: 1,
				SenderID:    1,
				ReceiverID:  2,
				SendStatus:  1,
				ReadStatus:  0,
				Status:      1,
				Body: ChatMessageBody{
					Text:          "纯文本消息",
					Src:           "",
					Format:        "",
					Size:          "",
					Longitude:     "",
					Latitude:      "",
					Scale:         0,
					LocationLabel: "",
				},
				CreatedAt: time.Now().UnixMilli(),
				UpdatedAt: time.Now().UnixMilli(),
			}
			err := AddChatMessage(msg)
			if err != nil {
				log.Println(err)
			}
		}
	})
}

func TestAddMessageWithTransaction(t *testing.T) {

	msg := &ChatMessage{
		RoomID:      "0000000100000002",
		Type:        1,
		SessionType: 1,
		SenderID:    1,
		ReceiverID:  2,
		SendStatus:  1,
		ReadStatus:  0,
		Status:      1,
		Body: ChatMessageBody{
			Text:          "纯文本消息",
			Src:           "",
			Format:        "",
			Size:          "",
			Longitude:     "",
			Latitude:      "",
			Scale:         0,
			LocationLabel: "",
		},
		CreatedAt: time.Now().UnixMilli(),
		UpdatedAt: time.Now().UnixMilli(),
	}

	//t.Run("事务", func(t *testing.T) {
	err := AddChatMessageTx(msg)
	if err != nil {
		t.Errorf("AddChatMessageTx() error = %v, wantErr %v", err, nil)
	}

}

func TestRollbackMessage(t *testing.T) {
	t.Run("插入并回滚", func(t *testing.T) {
		msg := &ChatMessage{
			RoomID:      "0000000100000002",
			Type:        1,
			SessionType: 1,
			SenderID:    1,
			ReceiverID:  2,
			SendStatus:  1,
			ReadStatus:  0,
			Status:      1,
			Body: ChatMessageBody{
				Text:          "纯文本消息",
				Src:           "",
				Format:        "",
				Size:          "",
				Longitude:     "",
				Latitude:      "",
				Scale:         0,
				LocationLabel: "",
			},
			CreatedAt: time.Now().UnixMilli(),
			UpdatedAt: time.Now().UnixMilli(),
		}

		err := AddChatMessage(msg)
		if err != nil {
			t.Fatal(err)
		}

		if ok, err := RollbackChatMessage(msg.ID); (err != nil) != false {
			t.Errorf("RollbackChatMessage() error = %v, wantErr %v", err, false)
		} else {
			if !ok {
				t.Errorf("RollbackChatMessage() ok = %v, want %v", ok, true)
			}
		}

	})
}

func TestChatMessageList(t *testing.T) {
	t.Run("消息列表", func(t *testing.T) {

		rs, err := GetChatMessageList((&GetChatMessageListFilter{RoomID: "0000000100000002", SessionType: 1}).SetLastMessageID(100).SetLimit(1000))
		if err != nil {
			t.Fatal(err)
		}
		d, _ := json.Marshal(rs)
		log.Printf(
			"%s", d,
		)

	})
}
