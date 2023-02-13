package proxy

import (
	"reflect"
	"sync"
	"testing"

	"github.com/k8spin/prometheus-multi-tenant-proxy/internal/pkg"
)

func init() {
	config = &pkg.Authn{
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
	configLock = new(sync.RWMutex)
}

func Test_isAuthorized(t *testing.T) {
	type args struct {
		user string
		pass string
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 []string
	}{
		{
			"Valid User",
			args{
				"User-a",
				"pass-a",
			},
			true,
			[]string{"tenant-a"},
		}, {
			"Invalid User",
			args{
				"invalid",
				"pass-a",
			},
			false,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := isAuthorized(tt.args.user, tt.args.pass)
			if got != tt.want {
				t.Errorf("isAuthorized() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("isAuthorized() got1 = %v, want1 %v", got1, tt.want1)
			}
		})
	}
}
