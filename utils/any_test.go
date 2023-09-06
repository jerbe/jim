package utils

import (
	"fmt"
	"testing"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/30 13:26
  @describe :
*/

func TestEqual(t *testing.T) {
	type args struct {
		obj     any
		target  any
		targets []any
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "判断nil",
			args: args{
				obj: nil,
				target: func() *int {
					return nil
				}(),
				targets: nil,
			},
			want: true,
		},
		{
			name: "判断0",
			args: args{
				obj:     0,
				target:  1,
				targets: []any{0},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Equal(tt.args.obj, tt.args.target, tt.args.targets...); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIn(t *testing.T) {
	type args struct {
		obj     any
		target  any
		targets []any
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "isNil",
			args: args{
				obj:     nil,
				target:  func() *int { return nil }(),
				targets: []any{9},
			},
			want: true,
		},
		{
			name: "isOther",
			args: args{
				obj:     "nil",
				target:  func() string { return "nil" }(),
				targets: []any{9},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := In(tt.args.obj, tt.args.target, tt.args.targets...); got != tt.want {
				t.Errorf("In() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSliceUnique(t *testing.T) {
	type args struct {
		data any
		dst  any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "相同",
			args: args{
				data: []int{1, 1, 2, 2, 3, 4, 5, 6, 7},
				dst:  &([]int{1, 2, 3, 4, 3343434, 23}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SliceUnique(tt.args.data, tt.args.dst)
			fmt.Println(tt.args.dst)
			/*
				if err := SliceUnique(tt.args.data, tt.args.dst); (err != nil) != tt.wantErr {
					t.Errorf("SliceUnique() error = %v, wantErr %v", err, tt.wantErr)
				}
			*/
		})
	}
}
