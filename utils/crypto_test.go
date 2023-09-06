package utils

import (
	"encoding/hex"
	"reflect"
	"testing"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/12 23:23
  @describe :
*/

func TestMD5(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
		{
			name: "测试结果",
			args: args{src: []byte("root")},
			want: func() []byte {
				data, _ := hex.DecodeString("63a9f0ea7bb98050796b649e85481845")
				return data
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MD5(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MD5() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMD5String(t *testing.T) {
	type args struct {
		src string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "测试结果",
			args: args{src: "root"},
			want: "63a9f0ea7bb98050796b649e85481845",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MD5String(tt.args.src); got != tt.want {
				t.Errorf("MD5String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPasswordHash(t *testing.T) {
	type args struct {
		pwd string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "测试结果",
			args: args{pwd: "root"},
			want: "b9be11166d72e9e3ae7fd407165e4bd2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PasswordHash(tt.args.pwd); got != tt.want {
				t.Errorf("PasswordHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
