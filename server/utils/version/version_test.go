package version

import "testing"

func TestLT(t *testing.T) {
	type args struct {
		version1 string
		version2 string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				version1: "1.0.0",
				version2: "1.0.0",
			},
			want: false,
		},
		{
			name: "test1",
			args: args{
				version1: "2.0.0",
				version2: "1.0.0",
			},
			want: false,
		},
		{
			name: "test1",
			args: args{
				version1: "1.0.0",
				version2: "2.0.0",
			},
			want: true,
		},
		{
			name: "test1",
			args: args{
				version1: "",
				version2: "1.0.0",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LT(tt.args.version1, tt.args.version2); got != tt.want {
				t.Errorf("LT() = %v, want %v", got, tt.want)
			}
		})
	}
}
