package pkg

import (
	"reflect"
	"testing"
)

func TestParseConfig(t *testing.T) {
	configInvalidLocation := "../../configs/no.config.yaml"
	configInvalidConfigFileLocation := "../../configs/bad.yaml"
	configSampleLocation := "../../configs/sample.yaml"
	configMultipleUserLocation := "../../configs/multiple.user.yaml"
	expectedSampleAuth := Authn{
		[]User{
			User{
				"Happy",
				"Prometheus",
				"tenant-1",
			},
		},
	}
	expectedMultipleUserAuth := Authn{
		[]User{
			User{
				"User-a",
				"pass-a",
				"tenant-a",
			},
			User{
				"User-b",
				"pass-b",
				"tenant-b",
			},
		},
	}
	type args struct {
		location *string
	}
	tests := []struct {
		name    string
		args    args
		want    *Authn
		wantErr bool
	}{
		{
			"Basic",
			args{
				&configSampleLocation,
			},
			&expectedSampleAuth,
			false,
		}, {
			"Multiples users",
			args{
				&configMultipleUserLocation,
			},
			&expectedMultipleUserAuth,
			false,
		}, {
			"Invalid location",
			args{
				&configInvalidLocation,
			},
			nil,
			true,
		}, {
			"Invalid yaml file",
			args{
				&configInvalidConfigFileLocation,
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConfig(tt.args.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
