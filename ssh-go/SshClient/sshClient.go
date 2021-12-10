package SshClient

import (
	"bufio"
	"fmt"
	"log"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/ssh"
)

type CmdRes struct {
	Msg       string
	ResHandle func(str string)
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
	sshChannel ssh.Channel
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
	client.sshChannel = channel
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
		client.log("SshClient 创建ssh session失败", err)
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
	if !ok || err != nil {
		client.log("远程虚拟窗口打开失败->", err)
		return client, nil
	}
	//泡吧
	go client.runClient()

	time.Sleep(time.Second * 5)

	return client, nil
}

// func (client *Client) sendCmd(cmd string) (string, error) {
// 	// 执行远程命令
// 	combo, err := client.sshSession.CombinedOutput(cmd)
// 	// if err != nil {
// 	// 	log.Println("远程执行cmd失败->", err)
// 	// }
// 	// log.Println("命令输出:", string(combo))
// 	return string(combo), err
// }

func (client *Client) runClient() {

	//获取输入
	go func() {
		for {
			cmd := <-client.Cmd
			goCmd := fmt.Sprint(cmd.Msg, "\n")
			fmt.Println(goCmd)
			// client.log([]byte(goCmd))
			// msg, err := client.sendCmd(cmd.Msg)
			_, err := client.sshChannel.Write([]byte(goCmd))
			if err != nil {
				log.Println(client.IpAddress, "-远程执行cmd失败->", err)
			}
			// cmd.ResHandle(msg)
		}
	}()

	//第二个协程将远程主机的返回结果返回给用户
	go func() {
		br := bufio.NewReader(client.sshChannel)
		buf := []byte{}
		t := time.NewTimer(time.Microsecond * 100)
		defer t.Stop()
		// 构建一个信道, 一端将数据远程主机的数据写入, 一段读取数据写入ws
		r := make(chan rune)

		// 另起一个协程, 一个死循环不断的读取ssh channel的数据, 并传给r信道直到连接断开
		go func() {
			defer client.sshSession.Close()
			defer client.sshClient.Close()

			for {
				x, size, err := br.ReadRune()
				if err != nil {
					client.log(err)
					client.KillMe <- true
					return
				}
				if size > 0 {
					r <- x
				}
			}
		}()

		// 主循环
		for {
			select {
			// 每隔100微秒, 只要buf的长度不为0就将数据
			case <-t.C:
				if len(buf) != 0 {
					client.log(buf)
					buf = []byte{}
				}
				t.Reset(time.Microsecond * 100)
			// 前面已经将ssh channel里读取的数据写入创建的通道r, 这里读取数据, 不断增加buf的长度, 在设定的 100 microsecond后由上面判定长度是否返送数据
			case d := <-r:
				if d != utf8.RuneError {
					p := make([]byte, utf8.RuneLen(d))
					utf8.EncodeRune(p, d)
					buf = append(buf, p...)
				} else {
					buf = append(buf, []byte("@")...)
				}
			}
		}
	}()

	for {
		select {
		case <-client.KillMe:
			client.IsAlive = false
			client.log("[退出]")
			return
		}
	}
}

func (client *Client) log(v ...interface{}) {
	log.Println("[", client.IpAddress, "]", fmt.Sprint(v...))
}
