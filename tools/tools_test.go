package tools

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"testing"
)

func TestGetRootPath(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRootPath(); got != tt.want {
				t.Errorf("GetRootPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRunPath2(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRunPath2(); got != tt.want {
				t.Errorf("GetRunPath2() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsArray(t *testing.T) {
	type args struct {
		array []string
		arr   string
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
			if got := IsArray(tt.args.array, tt.args.arr); got != tt.want {
				t.Errorf("IsArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsFileExist(t *testing.T) {
	type args struct {
		path string
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
			got, err := IsFileExist(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsFileExist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsFileExist() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsFileNotExist(t *testing.T) {
	type args struct {
		path string
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
			got, err := IsFileNotExist(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsFileNotExist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsFileNotExist() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJsonWrite(t *testing.T) {
	type args struct {
		context *gin.Context
		status  int
		result  interface{}
		msg     string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			JsonWrite(tt.args.context, tt.args.status, tt.args.result, tt.args.msg)
		})
	}
}

func TestRandStringRunes(t *testing.T) {
	fmt.Println(RandStringRunes(36))
}
