package main

import (
	"fmt"
	"log"
	"ssh-go/Config"
	"ssh-go/SshClient"
)

func main() {
	// var mapClient map[int]*(SshClient.Client)

	data, err := Config.LoadCofig("./cmd.json")
	if err != nil {
		log.Fatal("err-->", err)
		return
	}

	for _, v := range data.Example {
		// fmt.Print(i)
		// fmt.Println(v)
		client, err := SshClient.NewSshClient(v.User, v.Pwd, v.Host, 22)
		if err == nil {
			// mapClient[i] = client
			go func(x Config.Example) {
				for _, cv := range x.Cmd {
					client.Cmd <- &SshClient.CmdRes{
						Msg:       cv,
						ResHandle: func(str string) { fmt.Println(client.IpAddress, "[res]->", str) },
					}
				}
				fmt.Println(client.IpAddress, "[end]")
			}(v)
		} else {
			fmt.Println(client.IpAddress, err)
		}

	}

	for {
	}

	// fmt.Println("hello Test")
	// client := SshClient.NewSshClient("linxiqin", "lmon.com", "192.168.188.188", 22)
	// time.Sleep(time.Second * 5)
	// client.Cmd <- &SshClient.CmdRes{
	// 	Msg:       "whoami; cd /; ls -al;",
	// 	ResHandle: func(res string) { fmt.Println(res) },
	// }
	// time.Sleep(time.Second * 5)
	// fmt.Println("ok end")
}
