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
	KillMe    chan bool
	IsAlive   bool
	session   *ssh.Session
	client    *ssh.Client
	Id        int
}

func NewSshClient(Username string, Password string, IpAddress string, Port int) (*Client, error) {
	client := &Client{
		Username:  Username,
		Password:  Password,
		IpAddress: IpAddress,
		Port:      Port,
		KillMe:    make(chan bool),
		Cmd:       make(chan *CmdRes),
		IsAlive:   true,
	}

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
		// log.Println("dial 创建 ssh client 失败->", err)
		client.IsAlive = false
		log.Println(client.IpAddress, "-dial 创建 ssh client 失败->", err)
		return client, err
	}
	client.client = SshClient

	// 创建ssh-session
	session, err := SshClient.NewSession()
	if err != nil {
		log.Println(client.IpAddress, "-SshClient 创建ssh session失败", err)
		client.IsAlive = false
		return client, err
	}
	client.session = session

	//泡吧
	go client.runClient()

	return client, nil
}

func (client *Client) sendCmd(cmd string) (string, error) {
	// 执行远程命令
	combo, err := client.session.CombinedOutput(cmd)
	// if err != nil {
	// 	log.Println("远程执行cmd失败->", err)
	// }
	// log.Println("命令输出:", string(combo))
	return string(combo), err
}

func (client *Client) runClient() {

	go func() {
		for {
			cmd := <-client.Cmd
			msg, err := client.sendCmd(cmd.Msg)
			if err != nil {
				log.Println(client.IpAddress, "-远程执行cmd失败->", err)
			}
			cmd.ResHandle(msg)
		}
	}()

	go func() {
		for {
			select {
			case <-time.After(time.Second * 2):
				log.Println(client.IpAddress, "-[来吧！！！]")
			}
			fmt.Println("12345531")
		}
	}()

	for {
		select {
		case <-client.KillMe:
			client.session.Close()
			client.client.Close()
			client.IsAlive = false
			log.Println(client.IpAddress, "-[退出]")
			return

		case <-time.After(time.Second * 10):
			log.Println(client.IpAddress, "-[来吧！！！]")
			client.KillMe <- true

		}
	}

}
