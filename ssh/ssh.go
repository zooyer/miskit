package ssh

import (
	"errors"
	"io"
	"net"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// 创建ssh客户端
func Client(user, password, addr string) (client *ssh.Client, err error) {
	var config = ssh.ClientConfig{
		User:    user,
		Auth:    []ssh.AuthMethod{ssh.Password(password)},
		Timeout: time.Second * 30,
		//这个是问你要不要验证远程主机，以保证安全性。这里不验证
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	if client, err = ssh.Dial("tcp", addr, &config); err != nil {
		return
	}

	return
}

// 创建ssh客户端与会话
func Session(user, password, addr string) (client *ssh.Client, session *ssh.Session, err error) {
	if client, err = Client(user, password, addr); err != nil {
		return
	}

	if session, err = client.NewSession(); err != nil {
		return
	}

	return
}

// 创建sftp客户端
func SftpClient(user, password, addr string) (client *sftp.Client, err error) {
	sshClient, err := Client(user, password, addr)
	if err != nil {
		return
	}

	if client, err = sftp.NewClient(sshClient); err != nil {
		return
	}

	return
}

// 数据拷贝
func ScpReader(client *sftp.Client, filename string, reader io.Reader) (err error) {
	if client == nil {
		return errors.New("sftp client is nil")
	}
	if reader == nil {
		return errors.New("sftp reader is nil")
	}

	file, err := client.Create(filename)
	if err != nil {
		return
	}
	defer file.Close()

	if _, err = io.Copy(file, reader); err != nil {
		return
	}

	return
}

// 文件拷贝
func Scp(local, remote, password string, fn func(current, total int64)) (err error) {
	user, addr, filename, err := parse(remote)
	if err != nil {
		return
	}

	sftpClient, err := SftpClient(user, password, addr)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	src, err := os.Open(local)
	if err != nil {
		return err
	}
	defer src.Close()

	var total int64
	if fn != nil {
		stat, err := src.Stat()
		if err != nil {
			return err
		}
		total = stat.Size()
	}

	if err = ScpReader(sftpClient, filename, newReader(src, func(size int) {
		if fn != nil {
			fn(int64(size), total)
		}
	})); err != nil {

	}

	return
}

// 会话执行命令
func CommandSession(session *ssh.Session, cmd string) (output string, err error) {
	data, err := session.CombinedOutput(cmd)
	if err != nil {
		return
	}
	return strings.TrimSpace(string(data)), nil
}

// 执行命令
func Command(remote, password, cmd string) (output string, err error) {
	user, addr, _, err := parse(remote)
	if err != nil {
		return
	}

	client, session, err := Session(user, password, addr)
	if err != nil {
		return
	}
	defer client.Close()
	defer session.Close()

	return CommandSession(session, cmd)
}

func username() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	if runtime.GOOS == "windows" {
		if fields := strings.Split(u.Username, "\\"); len(fields) > 1 {
			return fields[1]
		}
	}
	return u.Username
}

func parse(target string) (user, addr, filename string, err error) {
	var port = "22"

	defer func() {
		addr += ":" + port
		if user == "" {
			user = username()
		}
		if filename == "" {
			filename = "/"
		}
	}()

	if index := strings.Index(target, "@"); index != -1 {
		user = target[:index]
		target = target[index+1:]
	}

	switch index := strings.Index(target, ":"); strings.Count(target, ":") {
	case 1:
		addr = target[:index]
		if target = target[index+1:]; len(target) == 0 {
			return
		}
		if c := target[0]; c >= '0' && c <= '9' {
			if _, err = strconv.Atoi(target[1:]); err != nil {
				return
			}
			port = target
		} else {
			filename = target
		}
	case 2:
		addr = target[:index]
		target = target[index+1:]
		if index = strings.Index(target, ":"); index != -1 {
			port = target[:index]
			target = target[index+1:]
		}
		filename = target
	default:
		addr = target
	}

	return
}
