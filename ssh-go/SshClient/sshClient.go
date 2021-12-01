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

type ptyRequestMsg struct {
	Term     string
	Columns  uint32
	Rows     uint32
	Width    uint32
	Height   uint32
	Modelist string
}

type Client struct {
	Username   string
	Password   string
	IpAddress  string
	Port       int
	Cmd        chan *CmdRes
	KillMe     chan bool
	IsAlive    bool
	sshSession *ssh.Session
	sshClient  *ssh.Client
	sshChannel *ssh.Channel
	Id         int
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
		client.log(client.IpAddress, "-dial 创建 ssh client 失败->", err)
		return client, err
	}
	client.sshClient = SshClient

	//创建远程端shell
	channel, inRequests, err := client.sshClient.OpenChannel("session", nil)
	if err != nil {
		log.Println(client.IpAddress, "创建远程端shell->", err)
		return client, err
	}
	client.sshChannel = &channel
	go func() {
		for req := range inRequests {
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}()

	// 创建ssh-session
	session, err := SshClient.NewSession()
	if err != nil {
		client.log(client.IpAddress, "-SshClient 创建ssh session失败", err)
		client.IsAlive = false
		return client, err
	}
	client.sshSession = session

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	client.sshSession.RequestPty("xterm", 35, 150, modes)

	ok, err := channel.SendRequest("shell", true, nil)
	fmt.Println(ok, err)
	if !ok || err != nil {
		log.Println(client.IpAddress, "远程虚拟窗口打开失败->", err)
		return client, nil
	}

	//泡吧
	go client.runClient()

	return client, nil
}

func (client *Client) sendCmd(cmd string) (string, error) {
	// 执行远程命令
	combo, err := client.sshSession.CombinedOutput(cmd)
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
			client.sshSession.Close()
			client.sshClient.Close()
			client.IsAlive = false
			log.Println(client.IpAddress, "-[退出]")
			return

		case <-time.After(time.Second * 10):
			log.Println(client.IpAddress, "-[来吧！！！]")
			client.KillMe <- true

		}
	}
}

func (client *Client) log(v ...interface{}) {
	log.Println(v...)
}
