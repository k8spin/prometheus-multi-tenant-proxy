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
	configMultipleNamespacesLocation := "../../configs/multiple.namespaces.yaml"
	configSampleLabelsLocation := "../../configs/sample.labels.yaml"

	expectedSampleAuth := Authn{
		[]User{
			{
				Username:   "Happy",
				Password:   "Prometheus",
				Namespace:  "default",
				Labels:     map[string][]string{},
				Namespaces: []string{},
			}, {
				Username:   "Sad",
				Password:   "Prometheus",
				Namespace:  "kube-system",
				Labels:     map[string][]string{},
				Namespaces: []string{},
			},
		},
	}
	expectedSampleLabelsAuth := Authn{
		[]User{
			{
				Username:  "Happy",
				Password:  "Prometheus",
				Namespace: "",
				Labels: map[string][]string{
					"app":  []string{"happy", "sad"},
					"team": []string{"america"},
				},
				Namespaces: []string{},
			}, {
				Username:  "Sad",
				Password:  "Prometheus",
				Namespace: "",
				Labels: map[string][]string{
					"namespace": []string{"kube-system", "monitoring"},
				},
				Namespaces: []string{},
			}, {
				Username:  "bored",
				Password:  "Prometheus",
				Namespace: "",
				Labels: map[string][]string{
					"dep": []string{"system"},
				},
				Namespaces: []string{
					"default",
					"kube-system",
				},
			},
		},
	}
	expectedMultipleUserAuth := Authn{
		[]User{
			{
				Username:   "User-a",
				Password:   "pass-a",
				Namespace:  "tenant-a",
				Labels:     map[string][]string{},
				Namespaces: []string{},
			},
			{
				Username:   "User-b",
				Password:   "pass-b",
				Namespace:  "tenant-b",
				Labels:     map[string][]string{},
				Namespaces: []string{},
			},
		},
	}
	expectedMultipleNamespaceAuth := Authn{
		[]User{
			{
				Username:   "Happy",
				Password:   "Prometheus",
				Namespace:  "default",
				Labels:     map[string][]string{},
				Namespaces: []string{},
			},
			{
				Username:   "Sad",
				Password:   "Prometheus",
				Namespace:  "kube-system",
				Labels:     map[string][]string{},
				Namespaces: []string{},
			},
			{
				Username:  "Multiple",
				Password:  "Namespaces",
				Namespace: "monitoring",
				Labels:    map[string][]string{},
				Namespaces: []string{
					"default",
					"kube-system",
					"kube-public",
				},
			},
			{
				Username:  "Multiple",
				Password:  "NamespacesWithoutNamespace",
				Namespace: "",
				Labels:    map[string][]string{},
				Namespaces: []string{
					"default",
					"kube-system",
					"kube-public",
				},
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
			"Labels",
			args{
				&configSampleLabelsLocation,
			},
			&expectedSampleLabelsAuth,
			false,
		}, {
			"Multiples users",
			args{
				&configMultipleUserLocation,
			},
			&expectedMultipleUserAuth,
			false,
		}, {
			"Multiples namespaces",
			args{
				&configMultipleNamespacesLocation,
			},
			&expectedMultipleNamespaceAuth,
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
