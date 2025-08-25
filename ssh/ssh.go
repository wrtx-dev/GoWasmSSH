//go:build js && wasm
// +build js,wasm

package ssh

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/pkg/sftp"
	"github.com/wrtx-dev/gowasmssh/js"

	"golang.org/x/crypto/ssh"
)

type SSHClient struct {
	session         *ssh.Session
	url             string
	modes           *ssh.TerminalModes
	conn            net.Conn
	host            string
	port            int
	onData          js.JsValue
	term            js.JsValue
	user            string
	password        string
	phrase          string
	key             string
	showFingerPrint bool
	auth            []ssh.AuthMethod
	msgBox          js.JsValue
	confrimBox      js.JsValue
	statusCallBack  js.JsValue
	mut             sync.Mutex
	donotWarn       bool
	client          *ssh.Client
	sessionInput    js.JsValue
	sftp            *sftp.Client
}

func (c *SSHClient) close() {
	c.mut.Lock()
	if c.conn != nil {
		c.conn.Close()
		c.client.Close()
		c.session.Close()
		c.client = nil
		c.conn = nil
		c.session = nil
		if !c.onData.IsUndefined() {
			c.onData.Call("dispose")
		}
		c.onData = js.Global().Get("undefined")
	}
	if !c.statusCallBack.IsUndefined() {
		c.statusCallBack.Invoke(false)
	}
	c.mut.Unlock()
}

func (c *SSHClient) genAuthModes() error {
	if c.auth == nil {
		c.auth = make([]ssh.AuthMethod, 0)
	}
	if c.key != "" {
		signer, err := ssh.ParsePrivateKey([]byte(c.key))
		if err != nil {
			if _, ok := err.(*ssh.PassphraseMissingError); ok && c.phrase != "" {
				signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(c.key), []byte(c.phrase))
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
		c.auth = append(c.auth, ssh.PublicKeys(signer))
	}
	if c.password != "" {
		c.auth = append(c.auth, ssh.Password(c.password))
	}
	if len(c.auth) == 0 {
		return fmt.Errorf("parse ssh auth modes error")
	}
	return nil
}

func (c *SSHClient) connectTo() error {
	var url string
	ph := js.Global().Get("window").Get("privateProxyLocation")
	if !ph.IsUndefined() {
		url = "ws://" + ph.String()
	} else {
		url = c.url
	}
	url = fmt.Sprintf("%s/%s/%d", url, c.host, c.port)
	conn, err := js.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to connect to: %v err: %v", c.url, err)
	}
	c.url = url
	c.conn = conn
	return nil
}

func (c *SSHClient) disconnect(_ js.JsValue, args []js.JsValue) interface{} {
	go func() {
		c.close()
		c.term.Call("writeln", "User terminates the session")
		c.msgBox = js.Global().Get("undefined")
		c.confrimBox = js.Global().Get("undefined")
		c.statusCallBack = js.Global().Get("undefined")
		c.sessionInput = js.Global().Get("undefined")
	}()
	return nil
}

func (c *SSHClient) breakWithMsg(title, msg string) {
	c.close()
	if !c.donotWarn {
		if !c.msgBox.IsUndefined() {
			c.msgBox.Invoke(title, msg)
		}

		c.donotWarn = !c.donotWarn
	}
	c.session = nil

}

func (c *SSHClient) warnning(msg string) {
	if !c.msgBox.IsUndefined() {
		c.msgBox.Invoke("Warnning", msg)
	}
}

func (c *SSHClient) errorMsg(msg string) {
	if !c.msgBox.IsUndefined() {
		c.msgBox.Invoke("Error", msg)
	}
}

var idx = 0

func (c *SSHClient) confirmFingerPrint(msg string, accepted chan bool) {
	confirm := js.JsFuncOf(func(_ js.JsValue, args []js.JsValue) interface{} {
		idx++
		accepted <- args[0].Bool()
		return nil
	})
	go func() {
		c.confrimBox.Invoke("ssh Security Alert", msg, confirm)
	}()
}

func (c *SSHClient) jsSetHostInfo(_ js.JsValue, args []js.JsValue) interface{} {
	if len(args) < 2 {
		return js.Global().Get("error").New("need host and port")
	}
	c.host, c.port = args[0].String(), args[1].Int()
	return nil
}

func (c *SSHClient) jsSetUserPassword(_ js.JsValue, args []js.JsValue) interface{} {
	if len(args) < 2 {
		return js.Global().Get("error").New("need user, password ")
	}
	c.user, c.password = args[0].String(), args[1].String()

	return nil
}

func (c *SSHClient) jsSetShowFingerPrint(_ js.JsValue, args []js.JsValue) interface{} {
	if len(args) < 1 {
		return js.Global().Get("error").New("need showFingerPrint bool")
	}
	c.showFingerPrint = args[0].Bool()
	return nil
}

func (c *SSHClient) jsSetTerminal(_ js.JsValue, args []js.JsValue) interface{} {
	if len(args) < 1 {
		return js.Global().Get("error").New("need terminal object")
	}
	c.term = args[0]
	return nil
}

func (c *SSHClient) jsSetPrivateKey(_ js.JsValue, args []js.JsValue) interface{} {
	if len(args) < 1 {
		return js.Global().Get("error").New("need private key")
	}
	if len(args) == 2 {
		c.key, c.phrase = args[0].String(), args[1].String()
	} else {
		c.key = args[0].String()
	}
	return nil
}

func (c *SSHClient) jsSessionInput(_ js.JsValue, args []js.JsValue) interface{} {
	if !c.sessionInput.IsUndefined() {
		anyArgs := make([]any, len(args))
		for i, arg := range args {
			anyArgs[i] = arg
		}
		c.sessionInput.Invoke(anyArgs...)
	}
	return nil
}

func (c *SSHClient) jsSetCallback(_ js.JsValue, args []js.JsValue) interface{} {
	if len(args) < 3 {
		return js.Global().Get("error").New("need 3 callback function")
	}
	c.msgBox = args[0]
	c.confrimBox = args[1]
	c.statusCallBack = args[2]
	return nil
}

func (c *SSHClient) jsSSHFunc(_ js.JsValue, args []js.JsValue) interface{} {
	go func() {

		if c.session != nil {
			c.close()
		}
		if !c.onData.IsUndefined() {
			c.onData.Call("dispose")
		}

		if err := c.connectTo(); err != nil {
			c.errorMsg(fmt.Sprintf("connect to host %s err: %v", c.host, err))
			return
		}

		if err := c.genAuthModes(); err != nil {
			c.errorMsg(fmt.Sprintf("%v", err))
			return
		}
		hostKeyCallback := func(hostname string, remote net.Addr, key ssh.PublicKey) error {

			// c.accepted <- v.Bool()
			fmt.Printf("%x", key.Marshal())
			msg := fmt.Sprintf("%s:%x", key.Type(), key.Marshal())

			var accepted = make(chan bool)

			c.confirmFingerPrint(msg, accepted)
			ac := <-accepted
			if !ac {
				return fmt.Errorf("user canceld the server key")
			}
			return nil
		}

		if !c.showFingerPrint {
			hostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			}
		}
		sshConf := &ssh.ClientConfig{
			User:            c.user,
			Auth:            c.auth,
			HostKeyCallback: hostKeyCallback,
		}
		sc, nc, r, err := ssh.NewClientConn(c.conn, c.host, sshConf)
		if err != nil {
			c.errorMsg(fmt.Sprintf("failed to open ssh connection: %v", err))
			if sc != nil {
				sc.Close()
			}
			c.auth = nil
			return
		}

		sshClient := ssh.NewClient(sc, nc, r)
		c.client = sshClient
		session, err := sshClient.NewSession()
		if err != nil {
			c.errorMsg(fmt.Sprintf("create new ssh session failed: %v", err))
			sc.Close()
			return
		}
		c.session = session
		stdout, err := session.StdoutPipe()
		if err != nil {
			c.errorMsg(fmt.Sprintf("failed to open stdout: %v", err))
			return
		}
		go func() {
			ob := make([]byte, 2048)
			for {
				n, err := stdout.Read(ob)
				if err != nil {
					if err != io.EOF {
						c.breakWithMsg("Error!!!!", fmt.Sprintf("connection closed: %v", err))
					} else {
						c.close()
					}
					return
				}
				c.term.Call("write", string(ob[:n]))
			}
		}()
		stderr, err := session.StderrPipe()
		if err != nil {
			c.breakWithMsg("Error!!!", fmt.Sprintf("failed to open stderr: %v", err))
			return
		}
		go func() {
			ob := make([]byte, 2048)
			for {
				n, err := stderr.Read(ob)
				if err != nil {
					// js.Alert(fmt.Sprintf("losed connection: %v", err))
					if err != io.EOF {
						c.breakWithMsg("Error!!!!", fmt.Sprintf("connection closed: %v", err))
					} else {
						c.close()
					}
					return
				}
				c.term.Get("term").Call("write", string(ob[:n]))
			}
		}()
		stdin, err := session.StdinPipe()
		if err != nil {
			c.breakWithMsg("Error!!!", fmt.Sprintf("open stdin failed: %v", err))
			return
		}
		if c.modes == nil {
			modes := ssh.TerminalModes{
				ssh.ECHO:          1,
				ssh.ICRNL:         1,
				ssh.IXON:          1,
				ssh.IXANY:         1,
				ssh.IMAXBEL:       1,
				ssh.OPOST:         1,
				ssh.ONLCR:         1,
				ssh.ISIG:          1,
				ssh.ICANON:        1,
				ssh.IEXTEN:        1,
				ssh.ECHOE:         1,
				ssh.ECHOK:         1,
				ssh.ECHOCTL:       1,
				ssh.ECHOKE:        1,
				ssh.TTY_OP_ISPEED: 14400,
				ssh.TTY_OP_OSPEED: 14400,
			}
			c.modes = &modes
		}
		rows, cols := c.term.Get("rows").Int(), c.term.Get("cols").Int()
		if err := session.RequestPty("xterm", rows, cols, *c.modes); err != nil {
			c.errorMsg(fmt.Sprintf("failed to request a pseudo terminal: %v", err))
			c.close()
			return
		}
		if err := session.Shell(); err != nil {
			c.breakWithMsg("Error!!!!", fmt.Sprintf("failed to start ssh shell: %v", err))
			c.close()
		}
		sessionInput := js.JsFuncOf(func(this js.JsValue, args []js.JsValue) interface{} {
			if len(args) < 1 {
				return nil
			}
			data := []byte(args[0].String())
			if _, err := stdin.Write(data); err != nil {
				c.breakWithMsg("Error!!!", "losed connection")
				return fmt.Errorf("error writing to stdin: %v", err)
			}
			return nil
		})
		c.sessionInput = sessionInput.Value
		if !c.statusCallBack.IsUndefined() {
			c.statusCallBack.Invoke(true)
		}
		sfc, err := sftp.NewClient(sshClient)
		if err == nil {
			c.sftp = sfc
		}

	}()
	return nil
}

func (c *SSHClient) resize(_ js.JsValue, args []js.JsValue) interface{} {
	if c.session == nil {
		return nil
	}
	rows, cols := c.term.Get("rows").Int(), c.term.Get("cols").Int()
	if err := c.session.WindowChange(rows, cols); err != nil {
		c.warnning(fmt.Sprintf("change window's size failed: %v", err))
	}
	return nil
}

func RegisterSSHNewConnection() {
	js.Global().Set("sshNewConnection", js.JsFuncOf(SSHNewConnection))
}

func SSHNewConnection(this js.JsValue, args []js.JsValue) interface{} {
	c := SSHClient{}
	if len(args) == 1 && args[0].Type().String() == "string" {
		proxyUrl := args[0].String()
		c.url = proxyUrl
	}
	sshClient := js.Global().Get("Object").New()
	sshClient.Set("connect", js.JsFuncOf(c.jsSSHFunc))
	sshClient.Set("disconnect", js.JsFuncOf(c.disconnect))
	sshClient.Set("resize", js.JsFuncOf(c.resize))
	sshClient.Set("setHostInfo", js.JsFuncOf(c.jsSetHostInfo))
	sshClient.Set("setUserPassword", js.JsFuncOf(c.jsSetUserPassword))
	sshClient.Set("setShowFingerPrint", js.JsFuncOf(c.jsSetShowFingerPrint))
	sshClient.Set("setTerminal", js.JsFuncOf(c.jsSetTerminal))
	sshClient.Set("setPrivateKey", js.JsFuncOf(c.jsSetPrivateKey))
	sshClient.Set("setCallback", js.JsFuncOf(c.jsSetCallback))
	sshClient.Set("sessionInput", js.JsFuncOf(c.jsSessionInput))
	sshClient.Set("sftClient", js.JsFuncOf(c.jsGetSFTClient))
	return sshClient
}

func (c *SSHClient) jsGetSFTClient(_ js.JsValue, args []js.JsValue) interface{} {
	if c.sftp != nil {
		return NewSFTPClient(c.sftp)
	}
	return js.Global().Get("undefined")
}
