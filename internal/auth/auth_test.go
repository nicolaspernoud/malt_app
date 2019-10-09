package auth

import (
	"testing"
)

func TestUser_IsMemberOf(t *testing.T) {
	type fields struct {
		Login       string
		DisplayName string
		MemberOf    []string
	}
	type args struct {
		group string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"user_is_in_group", fields{Login: "user", DisplayName: "Us Er", MemberOf: []string{"group01", "group02"}}, args{group: "group01"}, true},
		{"user_is_not_in_group", fields{Login: "user", DisplayName: "Us Er", MemberOf: []string{"group01", "group02"}}, args{group: "group03"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{
				Login:    tt.fields.Login,
				FullName: tt.fields.DisplayName,
				MemberOf: tt.fields.MemberOf,
			}
			if got := user.IsMemberOf(tt.args.group); got != tt.want {
				t.Errorf("User.IsMemberOf() = %v, want %v", got, tt.want)
			}
		})
	}
}
