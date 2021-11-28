package SshClient

import (
	"fmt"
	"testing"
	"time"
)

func Test(t *testing.T) {
	fmt.Println("hello Test")
	client := NewSshClient("linxiqin", "lmon.com", "192.168.188.188", 22)
	time.Sleep(time.Second * 5)
	client.Cmd <- &CmdRes{
		Msg:       "whoami; cd /; ls -al;",
		ResHandle: func(res string) { fmt.Println(res) },
	}
	time.Sleep(time.Second * 5)
	client.IsAlive <- true
	fmt.Println("ok end")
}
