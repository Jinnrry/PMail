package ssl

import (
	"reflect"
	"testing"
)

func TestGetServerParamsList(t *testing.T) {
	type args struct {
		serverName string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{name: "namesilo", args: args{serverName: "namesilo"}, want: []string{"NAMESILO_API_KEY"}, wantErr: false},
		{name: "namesiloAgain", args: args{serverName: "namesilo"}, want: []string{"NAMESILO_API_KEY"}, wantErr: false},
		{name: "auroradns", args: args{serverName: "auroradns"}, want: []string{"AURORA_API_KEY", "AURORA_SECRET"}, wantErr: false},
		{name: "alidns", args: args{serverName: "alidns"}, want: []string{"ALICLOUD_ACCESS_KEY", "ALICLOUD_SECRET_KEY"}, wantErr: false},
		{name: "null", args: args{serverName: "null"}, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetServerParamsList(tt.args.serverName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetServerParamsList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetServerParamsList() got = %v, want %v", got, tt.want)
			}
		})
	}
}
