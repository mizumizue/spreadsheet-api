package main

import (
	"testing"
)

func Test_LastColumnIndexToRangeChar(t *testing.T) {
	type args struct {
		colNum int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "",
			args: args{
				colNum: 0,
			},
			want: "A",
		},
		{
			name: "",
			args: args{
				colNum: 25,
			},
			want: "Z",
		},
		{
			name: "",
			args: args{
				colNum: 26,
			},
			want: "AA",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LastColumnIndexToRangeChar(tt.args.colNum); got != tt.want {
				t.Errorf("LastColumnIndexToRangeChar() = %v, want %v", got, tt.want)
			}
		})
	}
}
