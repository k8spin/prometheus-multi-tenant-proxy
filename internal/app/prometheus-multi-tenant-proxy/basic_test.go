package proxy

import (
	"reflect"
	"testing"

	"github.com/k8spin/prometheus-multi-tenant-proxy/internal/pkg"
)

var (
	auth *BasicAuth
)

func init() {
	config := &pkg.Authn{
		Users: []pkg.User{
			{
				Username:  "User-a",
				Password:  "pass-a",
				Namespace: "tenant-a",
			},
			{
				Username:  "User-b",
				Password:  "pass-b",
				Namespace: "tenant-b",
			},
		},
	}
	auth = newBasicAuthFromConfig(config)
}

func TestBasic_isAuthorized(t *testing.T) {
	type args struct {
		user string
		pass string
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 []string
		want2 map[string]string
	}{
		{
			"Valid User",
			args{
				"User-a",
				"pass-a",
			},
			true,
			[]string{"tenant-a"},
			nil,
		}, {
			"Invalid User",
			args{
				"invalid",
				"pass-a",
			},
			false,
			nil,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := auth.isAuthorized(tt.args.user, tt.args.pass)
			if got != tt.want {
				t.Errorf("isAuthorized() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("isAuthorized() got1 = %v, want1 %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("isAuthorized() got2 = %v, want2 %v", got2, tt.want2)
			}
		})
	}
}
