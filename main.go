package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	sysUser "os/user"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

//const (
//	APP_NAME        = "qssh"
//	APP_DESCRIPTION = "Quickly ssh"
//)
//
//var (
//	App     = kingpin.New(APP_NAME, APP_DESCRIPTION)
//	add     = App.Command("add", "Add a ssh account.")
//	connect = App.Command("connect", "Connect to the ssh.")
//)

type Host struct {
	IP       string
	Username string
	Port     int
}

func connect2(user, host string, port int) (*ssh.Session, error) {
	var (
		//auth         []ssh.AuthMethod
		addr    string
		client  *ssh.Client
		session *ssh.Session
		err     error
	)

	// A public key may be used to authenticate against the remote
	// server by using an unencrypted PEM-encoded private key file.
	//
	// If you have an encrypted private key, the crypto/x509 package
	// can be used to decrypt it.
	sUsr, err := sysUser.Current()

	if err != nil {
		log.Fatalf("unable to get user: %v", err)
	}

	configFilePath := fmt.Sprintf("%s/%s", sUsr.HomeDir, ".ssh/id_rsa")

	key, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// get auth method
	//auth = make([]ssh.AuthMethod, 0)
	//auth = append(auth, ssh.Password(password))

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if client, err = ssh.Dial("tcp", addr, config); err != nil {
		return nil, err
	}

	// create session
	if session, err = client.NewSession(); err != nil {
		return nil, err
	}

	return session, nil
}

func sshConnect(h *Host) {

	session, err := connect2(h.Username, h.IP, h.Port)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		panic(err)
	}
	defer term.Restore(fd, oldState)

	// excute command
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	termWidth, termHeight, err := term.GetSize(fd)
	if err != nil {
		panic(err)
	}

	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	// Request pseudo terminal
	if err := session.RequestPty("xterm-256color", termHeight, termWidth, modes); err != nil {
		log.Fatal(err)
	}

	session.Run("bash")
}

func findHost(configFilePath, server string) *Host {

	for _, e := range getConfigLine(configFilePath) {
		result := strings.Split(strings.TrimSpace(e), " ")
		if len(result) == 2 {
			if server == result[0] {
				part1 := strings.Split(result[1], "@")
				host := new(Host)
				host.Username = part1[0]

				part2 := strings.Split(part1[1], ":")

				if len(part2) == 2 {
					host.IP = part2[0]
					host.Port, _ = strconv.Atoi(part2[1])
				} else {
					host.IP = part1[1]
					host.Port = 22
				}

				return host
			}
		}
	}
	return nil
}

func connectCommand(configFilePath, server string) {
	//server := os.Args[2]
	host := findHost(configFilePath, server)

	if host != nil {
		sshConnect(host)
	} else {
		fmt.Printf("Unknown host %s \r\n", server)
	}

	//f.WriteString(line+"\r\n")

}

func getConfigLine(configFilePath string) []string {
	f, _ := os.OpenFile(configFilePath, os.O_RDONLY, 0644)
	defer f.Close()
	buf, _ := io.ReadAll(f)
	configsStr := string(buf)
	//fmt.Printf(configsStr)
	lines := strings.Split(configsStr, "\n")
	lines = lines[0 : len(lines)-1]
	return lines
}

func copyToRemote(host *Host, localFile, remoteFile string) {
	remote := fmt.Sprintf("%s@%s:%s", host.Username, host.IP, remoteFile)
	log.Default().Println(remote)
	cmd := exec.Command("scp", "-P", strconv.Itoa(host.Port), localFile, remote)
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
}

func main() {
	usr, err := sysUser.Current()
	if err != nil {
		log.Fatal(err)
	}

	configFilePath := fmt.Sprintf("%s/%s", usr.HomeDir, ".qssh")
	fmt.Printf("config path: %s\n", configFilePath)
	if len(os.Args) < 2 {
		fmt.Print("add, connect, list, copy")
		return
	}

	switch os.Args[1] {
	// Register user
	case "add":
		if len(os.Args) >= 4 {
			line := strings.Join(os.Args[2:4], " ")
			f, _ := os.OpenFile(configFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			defer f.Close()
			f.WriteString(line + "\r\n")
		}
	case "connect":
		if len(os.Args) >= 3 {
			server := os.Args[2]
			connectCommand(configFilePath, server)

		}
	case "list":
		lines := getConfigLine(configFilePath)
		for _, e := range lines {
			fmt.Printf("%s\r\n", e)
		}
	case "push":
		//if len(os.Args) >= 2 {
		//	server := os.Args[1]
		//	cmd := exec.Command("ssh-copy-id", "-i","~/.ssh/id_rsa.pub", server)
		//	out, err := cmd.CombinedOutput()
		//	if err != nil {
		//		log.Fatalf("cmd.Run() failed with %s\n", err)
		//	}
		//	fmt.Printf("combined out:\n%s\n", string(out))
		//}
	case "copy":
		if len(os.Args) >= 3 {
			server := os.Args[2]
			localFile := os.Args[3]
			host := findHost(configFilePath, server)

			if host != nil {

				copyToRemote(host, localFile, "~")
				//fmt.Printf("combined out:\n%s\n", string(out))
			}
		}
	default:
		if len(os.Args) >= 2 {
			server := os.Args[1]
			connectCommand(configFilePath, server)

		}

	}

	//f, err := os.OpenFile(configFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	//defer f.Close()
	//f.WriteString("test\r\n")

	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	////ioutil.ReadAll()
	//
	//f.WriteString("test\r\n")
	//
	//fmt.Printf("File contents: %s \r\n", content)

}
