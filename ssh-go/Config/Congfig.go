package Config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
)

type ConfigSSH struct {
	Example []struct {
		Host string   `json:"host"`
		User string   `json:"user"`
		Pwd  string   `json:"pwd"`
		Cmd  []string `json:"cmd"`
	} `json:"example"`
}

func LoadCofig(path string) (ConfigSSH, error) {
	var file_locker sync.Mutex //config file locker
	var config ConfigSSH

	file_locker.Lock()
	data, err := ioutil.ReadFile(path)
	file_locker.Unlock()

	if err != nil {
		fmt.Println("read json file error")
		return config, err
	}

	err = json.Unmarshal(data, &config)

	return config, err
}
