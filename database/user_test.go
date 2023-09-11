package database

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/13 12:05
  @describe :
*/

func TestUser(t *testing.T) {
	var user = &User{
		ID:       1,
		Username: "root",
		Password: "root",
	}
	fmt.Println(user)
}

func TestCheckUserExistByUsername(t *testing.T) {
	type args struct {
		username string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UserExistByUsername(tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserExistByUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UserExistByUsername() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	type args struct {
		id int64
	}
	tests := []struct {
		name    string
		args    args
		want    *User
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUser(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUserByUsername(t *testing.T) {
	type args struct {
		username string
	}
	tests := []struct {
		name    string
		args    args
		want    *User
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUserByUsername(tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserByUsername() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_MarshalBinary(t *testing.T) {
	type fields struct {
		ID        int64
		Username  string
		Password  string
		BirthDate *time.Time
		CreatedAt *time.Time
		UpdatedAt *time.Time
	}
	tests := []struct {
		name     string
		fields   fields
		wantData []byte
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				ID:        tt.fields.ID,
				Username:  tt.fields.Username,
				Password:  tt.fields.Password,
				BirthDate: tt.fields.BirthDate,
				CreatedAt: *tt.fields.CreatedAt,
				UpdatedAt: *tt.fields.UpdatedAt,
			}
			gotData, err := u.MarshalBinary()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotData, tt.wantData) {
				t.Errorf("MarshalBinary() gotData = %v, want %v", gotData, tt.wantData)
			}
		})
	}
}

func TestUser_UnmarshalBinary(t *testing.T) {
	type fields struct {
		ID        int64
		Username  string
		Password  string
		BirthDate *time.Time
		CreatedAt *time.Time
		CreatorID int64
		UpdatedAt *time.Time
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "默认",
			fields: fields{
				ID:       1,
				Username: "root",
				Password: "root",
				BirthDate: func() *time.Time {
					tt := time.Now()
					return &tt
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				ID:        tt.fields.ID,
				Username:  tt.fields.Username,
				Password:  tt.fields.Password,
				BirthDate: tt.fields.BirthDate,
				CreatedAt: *tt.fields.CreatedAt,
				UpdatedAt: *tt.fields.UpdatedAt,
			}
			if err := u.UnmarshalBinary(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalBinary() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
