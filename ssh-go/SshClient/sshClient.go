package SshClient

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/ssh"
)

type CmdRes struct {
	Msg       string
	ResHandle func(str string)
}

type Client struct {
	Username  string
	Password  string
	IpAddress string
	Port      int
	Cmd       chan *CmdRes
	IsAlive   chan bool
	session   *ssh.Session
}

func NewSshClient(Username string, Password string, IpAddress string, Port int) *Client {
	client := &Client{
		Username:  Username,
		Password:  Password,
		IpAddress: IpAddress,
		Port:      Port,
		Cmd:       make(chan *CmdRes),
		IsAlive:   make(chan bool),
	}

	go client.runClient()

	return client
}

func (client *Client) runClient() {
	// 创建ssh登录配置
	config := &ssh.ClientConfig{
		Timeout:         time.Second, // ssh连接time out时间一秒钟,如果ssh验证错误会在一秒钟返回
		User:            client.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(client.Password)}, //使用密码
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),                     // 这个可以,但是不够安全
		// HostKeyCallback: hostKeyCallBackFunc(h.Host),
	}

	// dial 获取ssh client
	addr := fmt.Sprintf("%s:%d", client.IpAddress, client.Port)
	SshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Fatal("dial 创建 ssh client 失败->", err)
	}
	defer SshClient.Close()

	// 创建ssh-session
	session, err := SshClient.NewSession()
	if err != nil {
		log.Fatal("SshClient 创建ssh session失败", err)
	}
	client.session = session
	defer session.Close()

	go func() {
		for {
			cmd := <-client.Cmd
			msg, err := client.sendCmd(cmd.Msg)
			if err != nil {
				log.Fatal("远程执行cmd失败->", err)
			}
			cmd.ResHandle(msg)
		}
	}()

	for {
		select {
		case <-client.IsAlive:
			log.Fatal("[退出链接]")
			return
		}
	}

}

func (client *Client) sendCmd(cmd string) (string, error) {
	// 执行远程命令
	combo, err := client.session.CombinedOutput(cmd)
	// if err != nil {
	// 	log.Fatal("远程执行cmd失败->", err)
	// }
	// log.Println("命令输出:", string(combo))
	return string(combo), err
}

func TestSshClient() {
	fmt.Println("hello ssh Client")
}
