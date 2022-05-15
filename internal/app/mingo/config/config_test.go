package config

import (
	"io/ioutil"
	"testing"
)

func TestDefaults(t *testing.T) {
	err := Build([]string{"-p", "1234"}, ioutil.Discard)
	if err != nil {
		t.Errorf("build err want nil, got %v", err)
	}

	conf := GetInstance()

	if conf.listenPort != 1234 {
		t.Errorf("port want 1234, got %d", conf.listenPort)
	}

	if conf.staticDir != "" {
		t.Errorf("staticDir want \"\", got %s", conf.staticDir)
	}

	if conf != GetInstance() {
		t.Error("GetInstance not singleton")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Should not be able to call Build() twice")
		}
	}()
	Build([]string{"-p", "1234"}, ioutil.Discard)
}
