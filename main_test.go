package main

import (
	"testing"
)

func TestLastCol(t *testing.T) {
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
		{
			name: "",
			args: args{
				colNum: 51,
			},
			want: "AZ",
		},
		{
			name: "",
			args: args{
				colNum: 52,
			},
			want: "AAA",
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
