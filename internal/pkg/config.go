package pkg

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// Authn Contains a list of users
type Authn struct {
	Users []User `yaml:"users"`
}

// User Identifies a user including the tenant
type User struct {
	Username   string            `yaml:"username"`
	Password   string            `yaml:"password"`
	Namespace  string            `yaml:"namespace"`
	Namespaces []string          `yaml:"namespaces"`
	Labels     map[string]string `yaml:"labels"`
}

// ParseConfig read a configuration file in the path `location` and returns an Authn object
func ParseConfig(location *string) (*Authn, error) {
	data, err := ioutil.ReadFile(*location)
	if err != nil {
		return nil, err
	}
	authn := Authn{}
	err = yaml.Unmarshal([]byte(data), &authn)
	if err != nil {
		return nil, err
	}
	for i := range authn.Users {
		if authn.Users[i].Namespaces == nil {
			authn.Users[i].Namespaces = []string{}
		}
		if authn.Users[i].Labels == nil {
			authn.Users[i].Labels = map[string]string{}
		}
	}
	return &authn, nil
}
