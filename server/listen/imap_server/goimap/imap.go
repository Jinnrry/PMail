package goimap

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"
)

var (
	eol = "\r\n"
)

// Server Imap服务实例
type Server struct {
	Domain           string        // 域名
	Port             int           // 端口
	TlsEnabled       bool          //是否启用Tls
	TlsConfig        *tls.Config   // tls配置
	ConnectAliveTime time.Duration // 连接存活时间，默认不超时
	Action           Action
	stop             chan bool
	close            bool
	lck              sync.Mutex
}

// NewImapServer 新建一个服务实例
func NewImapServer(port int, domain string, tlsEnabled bool, tlsConfig *tls.Config, action Action) *Server {
	return &Server{
		Domain:     domain,
		Port:       port,
		TlsEnabled: tlsEnabled,
		TlsConfig:  tlsConfig,
		Action:     action,
		stop:       make(chan bool, 1),
	}
}

// Start 启动服务
func (s *Server) Start() error {
	if !s.TlsEnabled {
		return s.startWithoutTLS()
	} else {
		return s.startWithTLS()
	}
}

func (s *Server) startWithTLS() error {
	if s.lck.TryLock() {
		listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", s.Port), s.TlsConfig)
		if err != nil {
			return err
		}
		s.close = false
		defer func() {
			listener.Close()
		}()

		go func() {
			for {
				conn, err := listener.Accept()
				if err != nil {
					if s.close {
						break
					} else {
						continue
					}
				}
				go s.handleClient(conn)
			}
		}()
		<-s.stop
	} else {
		return errors.New("Server Is Running")
	}

	return nil
}

func (s *Server) startWithoutTLS() error {
	if s.lck.TryLock() {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
		if err != nil {
			return err
		}
		s.close = false
		defer func() {
			listener.Close()
		}()

		go func() {
			for {
				conn, err := listener.Accept()
				if err != nil {
					if s.close {
						break
					} else {
						continue
					}
				}
				go s.handleClient(conn)
			}
		}()
		<-s.stop
	} else {
		return errors.New("Server Is Running")
	}

	return nil
}

// Stop 停止服务
func (s *Server) Stop() {
	s.close = true
	s.stop <- true
}

func (s *Server) handleClient(conn net.Conn) {
	slog.Debug("Imap conn")

	defer conn.Close()

	session := &Session{
		Conn:      conn,
		AliveTime: time.Now(),
	}
	if s.TlsEnabled && s.TlsConfig != nil {
		session.InTls = true
	}

	// 检查连接是否超时
	if s.ConnectAliveTime != 0 {
		go func() {
			for {
				if time.Now().Sub(session.AliveTime) >= s.ConnectAliveTime {
					session.Conn.Close()
					return
				}
				time.Sleep(3 * time.Second)
			}
		}()
	}

	reader := bufio.NewReader(conn)

	fmt.Fprintf(conn, "* OK %s Imap Server powered by goimap"+eol, s.Domain)

	for {
		rawLine, err := reader.ReadString('\n')
		if err != nil {
			conn.Close()
			return
		}
		session.AliveTime = time.Now()

		nub, cmd, args := getCommand(rawLine)
		slog.Debug(fmt.Sprintf("nub:%s cmd:%s args:%v", nub, cmd, args))

		switch cmd {
		case "CAPABILITY":
			commands, err := s.Action.CapaBility(session)
			if err != nil {
				fmt.Fprintf(conn, "* BAD %s%s", err.Error(), eol)
			} else {
				ret := fmt.Sprintf("%s ", nub)
				for _, command := range commands {
					ret += fmt.Sprintf("%s%s", command, " ")
				}
				ret += fmt.Sprintf("%s", eol)
				fmt.Fprintf(conn, ret)
			}

		case "CREATE":
			if len(args) != 1 {
				paramsErr(conn, "RENAME")
				break
			}
			err := s.Action.Create(session, args[0])
			output(conn, nub, err)
		case "DELETE":
			if len(args) != 1 {
				paramsErr(conn, "RENAME")
				break
			}
			err := s.Action.Delete(session, args[0])
			output(conn, nub, err)
		case "RENAME":
			if len(args) != 2 {
				paramsErr(conn, "RENAME")
			} else {
				err := s.Action.Rename(session, args[0], args[1])
				output(conn, nub, err)
			}
		case "LIST":
			if len(args) != 2 {
				paramsErr(conn, "RENAME")
			} else {
				rets, err := s.Action.List(session, args[0], args[1])
				if err != nil {
					showBad(conn, err)
				} else {
					ret := ""
					for _, str := range rets {
						ret += str + eol
					}
					fmt.Fprintf(conn, ret)
					showSucc(conn, nub)
				}
			}
		case "APPEND":
			slog.Debug("Append %s", args)
		case "SELECT":
			if len(args) != 1 {
				paramsErr(conn, "RENAME")
			} else {
				err := s.Action.Select(session, args[0])
				output(conn, nub, err)
			}
		case "FETCH":
			if len(args) != 2 {
				paramsErr(conn, "RENAME")
			} else {
				ret, err := s.Action.Fetch(session, args[0], args[1])
				if err != nil {
					showBad(conn, err)
				} else {
					fmt.Fprintf(conn, ret)
					showSucc(conn, ret)
				}
			}
		case "STORE":
			if len(args) != 2 {
				paramsErr(conn, "RENAME")
			} else {
				err := s.Action.Store(session, args[0], args[1])
				output(conn, nub, err)
			}
		case "CLOSE":
			err := s.Action.Close(session)
			output(conn, nub, err)
		case "EXPUNGE":
			err := s.Action.Expunge(session)
			output(conn, nub, err)
		case "EXAMINE":
			if len(args) != 1 {
				paramsErr(conn, "EXAMINE")
			}
			err := s.Action.Examine(session, args[0])
			output(conn, nub, err)
		case "SUBSCRIBE":
			if len(args) != 1 {
				paramsErr(conn, "SUBSCRIBE")
			} else {
				err := s.Action.Subscribe(session, args[0])
				output(conn, nub, err)
			}
		case "UNSUBSCRIBE":
			if len(args) != 1 {
				paramsErr(conn, "UNSUBSCRIBE")
			} else {
				err := s.Action.UnSubscribe(session, args[0])
				output(conn, nub, err)
			}
		case "LSUB":
			if len(args) != 2 {
				paramsErr(conn, "LSUB")
			} else {
				rets, err := s.Action.LSub(session, args[0], args[1])
				if err != nil {
					showBad(conn, err)
				} else {
					ret := ""
					for _, str := range rets {
						ret += str + eol
					}
					fmt.Fprintf(conn, ret)
					showSucc(conn, nub)
				}
			}
		case "STATUS":
			if len(args) != 2 {
				paramsErr(conn, "STATUS")
			} else {
				ret, err := s.Action.Status(session, args[0], args[1])
				if err != nil {
					showBad(conn, err)
				} else {
					fmt.Fprintf(conn, ret)
					showSucc(conn, nub)
				}
			}
		case "CHECK":
			err := s.Action.Check(session)
			output(conn, nub, err)
		case "SEARCH":
			if len(args) < 2 {
				paramsErr(conn, "SEARCH")
			} else {
				ret, err := s.Action.Search(session, args[0], args[1])
				if err != nil {
					showBad(conn, err)
				} else {
					fmt.Fprintf(conn, ret)
					showSucc(conn, nub)
				}
			}
		case "COPY":
			if len(args) != 2 {
				paramsErr(conn, "COPY")
			} else {
				err := s.Action.Copy(session, args[0], args[1])
				output(conn, nub, err)
			}

		case "NOOP":
			err := s.Action.Noop(session)
			output(conn, nub, err)
		case "LOGIN":
			if len(args) != 2 {
				paramsErr(conn, "LOGIN")
			} else {
				err := s.Action.Login(session, args[0], args[1])
				output(conn, nub, err)
			}
		case "LOGOUT":
			err := s.Action.Logout(session)
			output(conn, nub, err)
			conn.Close()
		default:
			rets, err := s.Action.Custom(session, cmd, args)
			if err != nil {
				fmt.Fprintf(conn, "* BAD %s %s", err.Error(), eol)
			} else {
				if len(rets) == 0 {
					fmt.Fprintf(conn, "%s OK %s", nub, eol)
				} else if len(rets) == 1 {
					fmt.Fprintf(conn, "%s OK %s%s", nub, rets[0], eol)
				} else {
					ret := fmt.Sprintf("%s OK %s", nub, eol)
					for _, re := range rets {
						ret += fmt.Sprintf("%s%s", re, eol)
					}
					ret += "." + eol
					fmt.Fprintf(conn, ret)
				}
			}
		}

	}
}

// cuts the line into command and arguments
func getCommand(line string) (string, string, []string) {
	line = strings.Trim(line, "\r \n")
	cmd := strings.Split(line, " ")

	return strings.ToTitle(cmd[0]), strings.ToTitle(cmd[1]), cmd[2:]
}

func getSafeArg(args []string, nr int) string {
	if nr < len(args) {
		return args[nr]
	}
	return ""
}

func showSucc(w io.Writer, nub string) {
	fmt.Fprintf(w, "%s OK success %s", nub, eol)
}

func showBad(w io.Writer, err error) {
	if err == nil {
		fmt.Fprintf(w, "* BAD %s", eol)
		return
	}
	fmt.Fprintf(w, "* BAD %s%s", err.Error(), eol)
}

func output(w io.Writer, nub string, err error) {
	if err != nil {
		showSucc(w, nub)
	} else {
		showBad(w, err)
	}
}

func paramsErr(w io.Writer, commend string) {
	fmt.Fprintf(w, "* BAD %s parameters! %s", commend, eol)
}
