package main

import (
	"log"
	"ssh-go/Config"
	"ssh-go/SshClient"
)

func main() {
	var mapClient map[string]*(SshClient.Client)

	log.Fatal("hello main...,begin")

	data, err := Config.LoadCofig("./cmd.json")
	if err != nil {
		log.Fatal("err-->", err)
		return
	}

	for i, v := range data.Example {
		// fmt.Print(i)
		// fmt.Println(v)
	}

	// fmt.Println("hello Test")
	client := SshClient.NewSshClient("linxiqin", "lmon.com", "192.168.188.188", 22)
	// time.Sleep(time.Second * 5)
	// client.Cmd <- &SshClient.CmdRes{
	// 	Msg:       "whoami; cd /; ls -al;",
	// 	ResHandle: func(res string) { fmt.Println(res) },
	// }
	// time.Sleep(time.Second * 5)
	// client.IsAlive <- true
	// fmt.Println("ok end")
}
