package goimap

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
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

	defer func() {
		if conn != nil {
			_ = conn.Close()
		}
	}()

	session := &Session{
		Conn:      conn,
		Status:    UNAUTHORIZED,
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
					if session.Conn != nil {
						write(session.Conn, "* BYE AutoLogout; idle for too long", "")
						_ = session.Conn.Close()
					}
					session.Conn = nil
					session.IN_IDLE = false
					return
				}
				time.Sleep(3 * time.Second)
			}
		}()
	}

	reader := bufio.NewReader(conn)
	write(conn, `* OK [CAPABILITY IMAP4 IMAP4rev1 AUTH=PLAIN AUTH=LOGIN] PMail Server ready`, "")

	for {
		rawLine, err := reader.ReadString('\n')
		if err != nil {
			if conn != nil {
				_ = conn.Close()
			}
			session.Conn = nil
			session.IN_IDLE = false
			return
		}
		session.AliveTime = time.Now()

		nub, cmd, args := getCommand(rawLine)
		log.Debugf("Imap Input:\t %s", rawLine)
		if cmd != "IDLE" {
			session.IN_IDLE = false
		}

		switch cmd {
		case "":
			if conn != nil {
				conn.Close()
				conn = nil
			}
			break

		case "CAPABILITY":
			commands, err := s.Action.CapaBility(session)
			if err != nil {
				write(conn, fmt.Sprintf("* BAD %s%s", err.Error(), eol), nub)
			} else {
				ret := "*"
				for _, command := range commands {
					ret += " " + command
				}
				write(conn, ret, nub)
				showSucc(conn, nub)
			}

		case "CREATE":
			if session.Status != AUTHORIZED {
				showBad(conn, errors.New("Need Login"), nub)
				break
			}
			if args == "" {
				paramsErr(conn, "CREATE", nub)
				break
			}
			err := s.Action.Create(session, args)
			output(conn, nub, err)
		case "DELETE":
			if session.Status != AUTHORIZED {
				showBad(conn, errors.New("Need Login"), nub)
				break
			}
			if args == "" {
				paramsErr(conn, "DELETE", nub)
				break
			}
			err := s.Action.Delete(session, args)
			output(conn, nub, err)
		case "RENAME":
			if session.Status != AUTHORIZED {
				showBad(conn, errors.New("Need Login"), nub)
				break
			}
			if args == "" {
				paramsErr(conn, "RENAME", nub)
			} else {
				dt := strings.Split(args, " ")
				err := s.Action.Rename(session, dt[0], dt[1])
				output(conn, nub, err)
			}
		case "LIST":
			if session.Status != AUTHORIZED {
				showBad(conn, errors.New("Need Login"), nub)
				break
			}
			if args == "" {
				paramsErr(conn, "LIST", nub)
			} else {
				dt := strings.Split(args, " ")
				dt[0] = strings.ReplaceAll(dt[0], `"`, "")
				rets, err := s.Action.List(session, dt[0], dt[1])
				if err != nil {
					showBad(conn, err, nub)
				} else {
					ret := ""
					for _, str := range rets {
						ret += str + eol
					}
					write(conn, ret, nub)
					showSucc(conn, nub)
				}
			}
		case "APPEND":
			if session.Status != AUTHORIZED {
				showBad(conn, errors.New("Need Login"), nub)
				break
			}
			log.Debugf("Append: %+v", args)
		case "SELECT":
			if session.Status != AUTHORIZED {
				showBad(conn, errors.New("Need Login"), nub)
				break
			}
			ret, err := s.Action.Select(session, args)
			args = strings.ReplaceAll(args, `"`, "")
			if err != nil {
				showBad(conn, err, nub)
			} else {
				for _, s2 := range ret {
					write(conn, s2, nub)
				}
			}

		case "FETCH":
			if session.Status != AUTHORIZED {
				showBad(conn, errors.New("Need Login"), nub)
				break
			}
			if args == "" {
				paramsErr(conn, "RENAME", nub)
			} else {
				dt := strings.Split(args, " ")
				ret, err := s.Action.Fetch(session, dt[0], dt[1])
				if err != nil {
					showBad(conn, err, nub)
				} else {
					write(conn, ret, nub)
					showSucc(conn, ret)
				}
			}
		case "STORE":
			if session.Status != AUTHORIZED {
				showBad(conn, errors.New("Need Login"), nub)
				break
			}
			if args == "" {
				paramsErr(conn, "RENAME", nub)
			} else {
				dt := strings.Split(args, " ")
				err := s.Action.Store(session, dt[0], dt[1])
				output(conn, nub, err)
			}
		case "CLOSE":
			err := s.Action.Close(session)
			output(conn, nub, err)
		case "EXPUNGE":
			if session.Status != AUTHORIZED {
				showBad(conn, errors.New("Need Login"), nub)
				break
			}
			err := s.Action.Expunge(session)
			output(conn, nub, err)
		case "EXAMINE":
			if session.Status != AUTHORIZED {
				showBad(conn, errors.New("Need Login"), nub)
				break
			}
			if args == "" {
				paramsErr(conn, "EXAMINE", nub)
			}
			err := s.Action.Examine(session, args)
			output(conn, nub, err)
		case "SUBSCRIBE":
			if session.Status != AUTHORIZED {
				showBad(conn, errors.New("Need Login"), nub)
				break
			}
			if args == "" {
				paramsErr(conn, "SUBSCRIBE", nub)
			} else {
				err := s.Action.Subscribe(session, args)
				output(conn, nub, err)
			}
		case "UNSUBSCRIBE":
			if session.Status != AUTHORIZED {
				showBad(conn, errors.New("Need Login"), nub)
				break
			}
			if args == "" {
				paramsErr(conn, "UNSUBSCRIBE", nub)
			} else {
				err := s.Action.UnSubscribe(session, args)
				output(conn, nub, err)
			}
		case "LSUB":
			if session.Status != AUTHORIZED {
				showBad(conn, errors.New("Need Login"), nub)
				break
			}
			if args == "" {
				paramsErr(conn, "LSUB", nub)
			} else {
				dt := strings.Split(args, " ")
				rets, err := s.Action.LSub(session, dt[0], dt[1])
				if err != nil {
					showBad(conn, err, nub)
				} else {
					ret := ""
					for _, str := range rets {
						ret += str + eol
					}
					write(conn, ret, nub)
					showSucc(conn, nub)
				}
			}
		case "STATUS":
			if session.Status != AUTHORIZED {
				showBad(conn, errors.New("Need Login"), nub)
				break
			}
			if args == "" {
				paramsErr(conn, "STATUS", nub)
			} else {
				dt := strings.SplitN(args, " ", 2)
				dt[0] = strings.ReplaceAll(dt[0], `"`, "")
				dt[1] = strings.Trim(dt[1], "()")
				params := strings.Split(dt[1], " ")

				ret, err := s.Action.Status(session, dt[0], params)
				if err != nil {
					showBad(conn, err, nub)
				} else {
					write(conn, ret, nub)
					showSucc(conn, nub)
				}
			}
		case "CHECK":
			if session.Status != AUTHORIZED {
				showBad(conn, errors.New("Need Login"), nub)
				break
			}
			err := s.Action.Check(session)
			output(conn, nub, err)
		case "SEARCH":
			if session.Status != AUTHORIZED {
				showBad(conn, errors.New("Need Login"), nub)
				break
			}
			if args == "" {
				paramsErr(conn, "SEARCH", nub)
			} else {
				dt := strings.SplitN(args, " ", 2)
				ret, err := s.Action.Search(session, dt[0], dt[1])
				if err != nil {
					showBad(conn, err, nub)
				} else {
					write(conn, ret, nub)
					showSucc(conn, nub)
				}
			}
		case "COPY":
			if session.Status != AUTHORIZED {
				showBad(conn, errors.New("Need Login"), nub)
				break
			}
			if args == "" {
				paramsErr(conn, "COPY", nub)
			} else {
				dt := strings.SplitN(args, " ", 2)
				err := s.Action.Copy(session, dt[0], dt[1])
				output(conn, nub, err)
			}

		case "NOOP":
			err := s.Action.Noop(session)
			output(conn, nub, err)
		case "LOGIN":
			if args == "" {
				paramsErr(conn, "LOGIN", nub)
			} else {
				dt := strings.SplitN(args, " ", 2)
				err := s.Action.Login(session, dt[0], dt[1])
				output(conn, nub, err)
			}
		case "LOGOUT":
			err := s.Action.Logout(session)
			write(conn, "* BYE PMail Server logging out", nub)
			output(conn, nub, err)
			if conn != nil {
				_ = conn.Close()
			}
		case "UNSELECT":
			if session.Status != AUTHORIZED {
				showBad(conn, errors.New("Need Login"), nub)
				break
			}
			err := s.Action.Unselect(session)
			output(conn, nub, err)
		case "IDLE":
			if session.Status != AUTHORIZED {
				showBad(conn, errors.New("Need Login"), nub)
				break
			}
			session.IN_IDLE = true
			err := s.Action.IDLE(session)
			if err != nil {
				write(conn, fmt.Sprintf("+ idling%s", eol), nub)
			} else {
				showBad(conn, err, nub)
			}
		default:
			rets, err := s.Action.Custom(session, cmd, args)
			if err != nil {
				write(conn, fmt.Sprintf("* BAD %s %s", err.Error(), eol), nub)
			} else {
				if len(rets) == 0 {
					write(conn, fmt.Sprintf("%s OK %s", nub, eol), nub)
				} else if len(rets) == 1 {
					write(conn, fmt.Sprintf("%s OK %s%s", nub, rets[0], eol), nub)
				} else {
					ret := fmt.Sprintf("%s OK %s", nub, eol)
					for _, re := range rets {
						ret += fmt.Sprintf("%s%s", re, eol)
					}
					ret += "." + eol
					write(conn, fmt.Sprintf(ret), nub)
				}
			}
		}

	}
}

// cuts the line into command and arguments
func getCommand(line string) (string, string, string) {
	line = strings.Trim(line, "\r \n")
	cmd := strings.SplitN(line, " ", 3)
	if len(cmd) == 1 {
		return "", "", ""
	}

	for i, s := range cmd {
		cmd[i] = s
	}

	if len(cmd) == 3 {
		return strings.ToTitle(cmd[0]), strings.ToTitle(cmd[1]), cmd[2]
	}

	return strings.ToTitle(cmd[0]), strings.ToTitle(cmd[1]), ""
}

func getSafeArg(args []string, nr int) string {
	if nr < len(args) {
		return args[nr]
	}
	return ""
}

func showSucc(w io.Writer, nub string) {
	write(w, fmt.Sprintf("%s OK success %s", nub, eol), nub)
}

func showBad(w io.Writer, err error, nub string) {
	if err == nil {
		write(w, fmt.Sprintf("* BAD %s", eol), nub)
		return
	}
	write(w, fmt.Sprintf("* BAD %s%s", err.Error(), eol), nub)
}

func output(w io.Writer, nub string, err error) {
	if err != nil {
		showBad(w, err, nub)
	} else {
		showSucc(w, nub)
	}
}

func paramsErr(w io.Writer, commend string, nub string) {
	write(w, fmt.Sprintf("* BAD %s parameters! %s", commend, eol), nub)
}

func write(w io.Writer, content string, nub string) {
	content = strings.ReplaceAll(content, "$$NUM", nub)
	log.Debugf("Imap Out:\t |%s", content)
	fmt.Fprintf(w, content)
}
