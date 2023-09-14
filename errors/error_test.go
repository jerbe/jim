package errors

import (
	"testing"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/9/14 15:37
  @describe :
*/

func TestError(t *testing.T) {
	var err error
	err = NewWithCaller("Hello world")
	err = Wrap(err)
	err = Wrap(err)

	t.Logf("%+v", err)
}

func TestIsEmptyRecord(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmptyRecord(tt.args.err); got != tt.want {
				t.Errorf("IsEmptyRecord() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNoRecord(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNoRecord(tt.args.err); got != tt.want {
				t.Errorf("IsNoRecord() = %v, want %v", got, tt.want)
			}
		})
	}
}
