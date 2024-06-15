package list

import (
	"pmail/dto"
	"pmail/utils/context"
	"reflect"
	"testing"
)

func Test_genSQL(t *testing.T) {
	type args struct {
		ctx      *context.Context
		count    bool
		tagInfo  dto.SearchTag
		keyword  string
		pop3List bool
		offset   int
		limit    int
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 []any
	}{
		{
			name: "Group搜索",
			args: args{
				ctx: &context.Context{
					UserID: 1,
				},
				count:    false,
				tagInfo:  dto.SearchTag{-1, -1, 2},
				keyword:  "",
				pop3List: false,
				offset:   0,
				limit:    0,
			},
			want:  "select e.*,ue.is_read from email e left join user_email ue on e.id=ue.email_id where ue.user_id = ?  and ue.status != 3 and ue.group_id=?  order by e.id desc limit 0,10 ",
			want1: []any{1, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := genSQL(tt.args.ctx, tt.args.count, tt.args.tagInfo, tt.args.keyword, tt.args.pop3List, tt.args.offset, tt.args.limit)
			if got != tt.want {
				t.Errorf("genSQL() got = \n%v, want \n%v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("genSQL() got1 = \n%v, want \n%v", got1, tt.want1)
			}
		})
	}
}
