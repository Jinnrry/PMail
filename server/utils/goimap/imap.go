package goimap

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/Jinnrry/pmail/utils/context"
	"github.com/Jinnrry/pmail/utils/id"
	log "github.com/sirupsen/logrus"
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

func (s *Server) authenticate(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	if args == "LOGIN" {
		write(session, "+ VXNlciBOYW1lAA=="+eol, "")
		line, err2 := reader.ReadString('\n')
		if err2 != nil {
			if conn != nil {
				_ = conn.Close()
			}
			session.Conn = nil
			session.IN_IDLE = false
			return
		}
		account, err := base64.StdEncoding.DecodeString(line)
		if err != nil {
			showBad(session, "Data Error.", nub)
			return
		}
		write(session, "+ UGFzc3dvcmQA"+eol, "")
		line, err = reader.ReadString('\n')
		if err2 != nil {
			if conn != nil {
				_ = conn.Close()
			}
			session.Conn = nil
			session.IN_IDLE = false
			return
		}
		password, err := base64.StdEncoding.DecodeString(line)
		res := s.Action.Login(session, string(account), string(password))
		if res.Type == SUCCESS {
			showSucc(session, res.Message, nub)
		} else if res.Type == BAD {
			showBad(session, res.Message, nub)
		} else {
			showNo(session, res.Message, nub)
		}
	} else {
		showBad(session, "Unsupported AUTHENTICATE mechanism.", nub)
	}
}

func (s *Server) capability(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	res := s.Action.CapaBility(session)
	if res.Type == BAD {
		write(session, fmt.Sprintf("* BAD %s%s", res.Message, eol), nub)
	} else {
		ret := "*"
		for _, command := range res.Data {
			ret += " " + command
		}
		ret += eol
		write(session, ret, nub)
		showSucc(session, res.Message, nub)
	}
}

func (s *Server) create(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	if session.Status != AUTHORIZED {
		showBad(session, "Need Login", nub)
		return
	}
	if args == "" {
		paramsErr(session, "CREATE", nub)
		return
	}
	res := s.Action.Create(session, args)
	showSucc(session, res.Message, nub)
}

func (s *Server) delete(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	if session.Status != AUTHORIZED {
		showBad(session, "Need Login", nub)
		return
	}
	if args == "" {
		paramsErr(session, "DELETE", nub)
		return
	}
	res := s.Action.Delete(session, args)
	if res.Type == SUCCESS {
		showSucc(session, res.Message, nub)
	} else if res.Type == BAD {
		showBad(session, res.Message, nub)
	} else {
		showNo(session, res.Message, nub)
	}
}

func (s *Server) rename(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	if session.Status != AUTHORIZED {
		showBad(session, "Need Login", nub)
		return
	}
	if args == "" {
		paramsErr(session, "RENAME", nub)
	} else {
		dt := strings.Split(args, " ")
		res := s.Action.Rename(session, dt[0], dt[1])
		if res.Type == SUCCESS {
			showSucc(session, res.Message, nub)
		} else if res.Type == BAD {
			showBad(session, res.Message, nub)
		} else {
			showNo(session, res.Message, nub)
		}
	}
}

func (s *Server) list(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	if session.Status != AUTHORIZED {
		showBad(session, "Need Login", nub)
		return
	}
	if args == "" {
		paramsErr(session, "LIST", nub)
	} else {
		dt := strings.Split(args, " ")
		dt[0] = strings.Trim(dt[0], `"`)
		dt[1] = strings.Trim(dt[1], `"`)
		res := s.Action.List(session, dt[0], dt[1])
		if res.Type == SUCCESS {
			showSuccWithData(session, res.Data, res.Message, nub)
		} else if res.Type == BAD {
			showBad(session, res.Message, nub)
		} else {
			showNo(session, res.Message, nub)
		}
	}
}

func (s *Server) append(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	if session.Status != AUTHORIZED {
		showBad(session, "Need Login", nub)
		return
	}
	log.WithContext(session.Ctx).Debugf("Append: %+v", args)
}

func (s *Server) cselect(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	if session.Status != AUTHORIZED {
		showBad(session, "Need Login", nub)
		return
	}
	res := s.Action.Select(session, args)
	if res.Type == SUCCESS {
		showSuccWithData(session, res.Data, res.Message, nub)
	} else if res.Type == BAD {
		showBad(session, res.Message, nub)
	} else {
		showNo(session, res.Message, nub)
	}
}

func (s *Server) fetch(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader, uid bool) {
	if session.Status != AUTHORIZED {
		showBad(session, "Need Login", nub)
		return
	}
	if args == "" {
		paramsErr(session, "FETCH", nub)
	} else {
		dt := strings.SplitN(args, " ", 2)
		if len(dt) != 2 {
			showBad(session, "Error Params", nub)
			return
		}

		res := s.Action.Fetch(session, dt[0], dt[1], uid)
		if res.Type == SUCCESS {
			showSuccWithData(session, res.Data, res.Message, nub)
		} else if res.Type == BAD {
			showBad(session, res.Message, nub)
		} else {
			showNo(session, res.Message, nub)
		}
	}
}

func (s *Server) store(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader, uid bool) {
	if session.Status != AUTHORIZED {
		showBad(session, "Need Login", nub)
		return
	}
	if args == "" {
		paramsErr(session, "RENAME", nub)
	} else {
		dt := strings.SplitN(args, " ", 2)
		res := s.Action.Store(session, dt[0], dt[1], uid)
		if res.Type == SUCCESS {
			showSucc(session, res.Message, nub)
		} else if res.Type == BAD {
			showBad(session, res.Message, nub)
		} else {
			showNo(session, res.Message, nub)
		}
	}
}

func (s *Server) cclose(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	res := s.Action.Close(session)
	if res.Type == SUCCESS {
		showSucc(session, res.Message, nub)
	} else if res.Type == BAD {
		showBad(session, res.Message, nub)
	} else {
		showNo(session, res.Message, nub)
	}
}

func (s *Server) expunge(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	if session.Status != AUTHORIZED {
		showBad(session, "Need Login", nub)
		return
	}
	res := s.Action.Expunge(session)
	if res.Type == SUCCESS {
		showSucc(session, res.Message, nub)
	} else if res.Type == BAD {
		showBad(session, res.Message, nub)
	} else {
		showNo(session, res.Message, nub)
	}
}

func (s *Server) examine(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	if session.Status != AUTHORIZED {
		showBad(session, "Need Login", nub)
		return
	}
	if args == "" {
		paramsErr(session, "EXAMINE", nub)
	}
	res := s.Action.Examine(session, args)
	if res.Type == SUCCESS {
		showSucc(session, res.Message, nub)
	} else if res.Type == BAD {
		showBad(session, res.Message, nub)
	} else {
		showNo(session, res.Message, nub)
	}
}

func (s *Server) unsubscribe(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	if session.Status != AUTHORIZED {
		showBad(session, "Need Login", nub)
		return
	}
	if args == "" {
		paramsErr(session, "UNSUBSCRIBE", nub)
	} else {
		res := s.Action.UnSubscribe(session, args)
		if res.Type == SUCCESS {
			showSucc(session, res.Message, nub)
		} else if res.Type == BAD {
			showBad(session, res.Message, nub)
		} else {
			showNo(session, res.Message, nub)
		}
	}
}

func (s *Server) lsub(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	if session.Status != AUTHORIZED {
		showBad(session, "Need Login", nub)
		return
	}
	if args == "" {
		paramsErr(session, "LSUB", nub)
	} else {
		dt := strings.Split(args, " ")
		dt[0] = strings.Trim(dt[0], `"`)
		res := s.Action.LSub(session, dt[0], dt[1])
		if res.Type == SUCCESS {
			showSuccWithData(session, res.Data, res.Message, nub)
		} else if res.Type == BAD {
			showBad(session, res.Message, nub)
		} else {
			showNo(session, res.Message, nub)
		}
	}
}

func (s *Server) status(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	if session.Status != AUTHORIZED {
		showBad(session, "Need Login", nub)
		return
	}
	if args == "" {
		paramsErr(session, "STATUS", nub)
	} else {
		var mailBox string
		var params []string
		if strings.HasPrefix(args, `"`) {
			dt := strings.Split(args, `"`)
			if len(dt) >= 3 {
				mailBox = dt[1]
			}
			dt[2] = strings.Trim(dt[2], "() ")
			params = strings.Split(dt[2], " ")
		} else {
			dt := strings.SplitN(args, " ", 2)
			dt[0] = strings.ReplaceAll(dt[0], `"`, "")
			dt[1] = strings.Trim(dt[1], "()")
			mailBox = dt[0]
			params = strings.Split(dt[1], " ")
		}

		res := s.Action.Status(session, mailBox, params)
		if res.Type == SUCCESS {
			showSuccWithData(session, res.Data, res.Message, nub)
		} else if res.Type == BAD {
			showBad(session, res.Message, nub)
		} else {
			showNo(session, res.Message, nub)
		}
	}
}

func (s *Server) check(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	if session.Status != AUTHORIZED {
		showBad(session, "Need Login", nub)
		return
	}
	res := s.Action.Check(session)
	if res.Type == SUCCESS {
		showSucc(session, res.Message, nub)
	} else if res.Type == BAD {
		showBad(session, res.Message, nub)
	} else {
		showNo(session, res.Message, nub)
	}
}

func (s *Server) search(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader, uid bool) {
	if session.Status != AUTHORIZED {
		showBad(session, "Need Login", nub)
		return
	}
	if args == "" {
		paramsErr(session, "SEARCH", nub)
	} else {
		var res CommandResponse
		if args == "ALL" {
			res = s.Action.Search(session, "", "UID 1:*", uid)
		} else {
			res = s.Action.Search(session, "", args, uid)
		}

		if res.Type == SUCCESS {
			content := "* SEARCH"
			for _, datum := range res.Data {
				content += " " + datum
			}
			content += eol
			content += fmt.Sprintf("%s OK SEARCH completed (Success)%s", nub, eol)
			write(session, content, nub)
		} else if res.Type == BAD {
			showBad(session, res.Message, nub)
		} else {
			showNo(session, res.Message, nub)
		}
	}
}

func (s *Server) copy(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	if session.Status != AUTHORIZED {
		showBad(session, "Need Login", nub)
		return
	}
	if args == "" {
		paramsErr(session, "COPY", nub)
	} else {
		dt := strings.SplitN(args, " ", 2)
		res := s.Action.Copy(session, dt[0], dt[1])
		if res.Type == SUCCESS {
			showSucc(session, res.Message, nub)
		} else if res.Type == BAD {
			showBad(session, res.Message, nub)
		} else {
			showNo(session, res.Message, nub)
		}
	}
}

func (s *Server) noop(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	res := s.Action.Noop(session)
	if res.Type == SUCCESS {
		showSucc(session, res.Message, nub)
	} else if res.Type == BAD {
		showBad(session, res.Message, nub)
	} else {
		showNo(session, res.Message, nub)
	}
}

func (s *Server) login(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	if args == "" {
		paramsErr(session, "LOGIN", nub)
	} else {
		dt := strings.SplitN(args, " ", 2)
		res := s.Action.Login(session, strings.Trim(dt[0], `"`), strings.Trim(dt[1], `"`))
		if res.Type == SUCCESS {
			showSucc(session, res.Message, nub)
		} else if res.Type == BAD {
			showBad(session, res.Message, nub)
		} else {
			showNo(session, res.Message, nub)
		}
	}
}

func (s *Server) logout(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	res := s.Action.Logout(session)
	write(session, "* BYE PMail Server logging out"+eol, nub)
	if res.Type == SUCCESS {
		showSucc(session, res.Message, nub)
	} else if res.Type == BAD {
		showBad(session, res.Message, nub)
	} else {
		showNo(session, res.Message, nub)
	}
	if conn != nil {
		_ = conn.Close()
	}
}

func (s *Server) unselect(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	if session.Status != AUTHORIZED {
		showBad(session, "Need Login", nub)
		return
	}
	res := s.Action.Unselect(session)
	if res.Type == SUCCESS {
		showSucc(session, res.Message, nub)
	} else if res.Type == BAD {
		showBad(session, res.Message, nub)
	} else {
		showNo(session, res.Message, nub)
	}
}

func (s *Server) subscribe(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	if session.Status != AUTHORIZED {
		showBad(session, "Need Login", nub)
		return
	}
	if args == "" {
		paramsErr(session, "SUBSCRIBE", nub)
	} else {
		res := s.Action.Subscribe(session, args)
		if res.Type == SUCCESS {
			showSucc(session, res.Message, nub)
		} else if res.Type == BAD {
			showBad(session, res.Message, nub)
		} else {
			showNo(session, res.Message, nub)
		}
	}
}

func (s *Server) idle(session *Session, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	if session.Status != AUTHORIZED {
		showBad(session, "Need Login", nub)
		return
	}
	session.IN_IDLE = true
	res := s.Action.IDLE(session)
	if res.Type == SUCCESS {
		write(session, "+ idling"+eol, nub)
	} else if res.Type == BAD {
		showBad(session, res.Message, nub)
	} else {
		showNo(session, res.Message, nub)
	}
}

func (s *Server) custom(session *Session, cmd string, args string, nub string, conn net.Conn, reader *bufio.Reader) {
	res := s.Action.Custom(session, cmd, args)
	if res.Type == BAD {
		write(session, fmt.Sprintf("* BAD %s %s", res.Message, eol), nub)
	} else if res.Type == NO {
		showNo(session, res.Message, nub)
	} else {
		if len(res.Data) == 0 {
			showSucc(session, res.Message, nub)
		} else {
			ret := ""
			for _, re := range res.Data {
				ret += fmt.Sprintf("%s%s", re, eol)
			}
			ret += "." + eol
			write(session, fmt.Sprintf(ret), nub)
		}
	}
}

func (s *Server) doCommand(session *Session, rawLine string, conn net.Conn, reader *bufio.Reader) {
	nub, cmd, args := getCommand(rawLine)
	log.WithContext(session.Ctx).Debugf("Imap Input:\t %s", rawLine)
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

	case "AUTHENTICATE":
		s.authenticate(session, args, nub, conn, reader)
	case "CAPABILITY":
		s.capability(session, rawLine, nub, conn, reader)
	case "CREATE":
		s.create(session, args, nub, conn, reader)
	case "DELETE":
		s.delete(session, args, nub, conn, reader)
	case "RENAME":
		s.rename(session, args, nub, conn, reader)
	case "LIST":
		s.list(session, args, nub, conn, reader)
	case "APPEND":
		s.append(session, args, nub, conn, reader)
	case "SELECT":
		s.cselect(session, args, nub, conn, reader)
	case "FETCH":
		s.fetch(session, args, nub, conn, reader, false)
	case "UID FETCH":
		s.fetch(session, args, nub, conn, reader, true)
	case "STORE":
		s.store(session, args, nub, conn, reader, false)
	case "UID STORE":
		s.store(session, args, nub, conn, reader, true)
	case "CLOSE":
		s.cclose(session, args, nub, conn, reader)
	case "EXPUNGE":
		s.expunge(session, args, nub, conn, reader)
	case "EXAMINE":
		s.examine(session, args, nub, conn, reader)
	case "SUBSCRIBE":
		s.subscribe(session, args, nub, conn, reader)
	case "UNSUBSCRIBE":
		s.unsubscribe(session, args, nub, conn, reader)
	case "LSUB":
		s.lsub(session, args, nub, conn, reader)
	case "STATUS":
		s.status(session, args, nub, conn, reader)
	case "CHECK":
		s.check(session, args, nub, conn, reader)
	case "SEARCH":
		s.search(session, args, nub, conn, reader, false)
	case "UID SEARCH":
		s.search(session, args, nub, conn, reader, true)
	case "COPY":
		s.copy(session, args, nub, conn, reader)
	case "NOOP":
		s.noop(session, args, nub, conn, reader)
	case "LOGIN":
		s.login(session, args, nub, conn, reader)
	case "LOGOUT":
		s.logout(session, args, nub, conn, reader)
	case "UNSELECT":
		s.unselect(session, args, nub, conn, reader)
	case "IDLE":
		s.idle(session, args, nub, conn, reader)
	default:
		s.custom(session, cmd, args, nub, conn, reader)
	}
}

func (s *Server) handleClient(conn net.Conn) {

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

	tc := &context.Context{}
	tc.SetValue(context.LogID, id.GenLogID())
	session.Ctx = tc

	if s.TlsEnabled && s.TlsConfig != nil {
		session.InTls = true
	}

	// 检查连接是否超时
	if s.ConnectAliveTime != 0 {
		go func() {
			for {
				if time.Now().Sub(session.AliveTime) >= s.ConnectAliveTime {
					if session.Conn != nil {
						write(session, "* BYE AutoLogout; idle for too long", "")
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
	write(session, fmt.Sprintf(`* OK [CAPABILITY IMAP4 IMAP4rev1 AUTH=LOGIN] PMail Server ready%s`, eol), "")

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

		s.doCommand(session, rawLine, conn, reader)

	}
}

// cuts the line into command and arguments
func getCommand(line string) (string, string, string) {
	line = strings.Trim(line, "\r \n")
	cmd := strings.SplitN(line, " ", 3)
	if len(cmd) == 1 {
		return "", "", ""
	}

	if len(cmd) == 3 {
		if strings.ToTitle(cmd[1]) == "UID" {
			args := strings.SplitN(cmd[2], " ", 2)
			if len(args) >= 2 {
				return cmd[0], strings.ToTitle(cmd[1]) + " " + strings.ToTitle(args[0]), args[1]
			}
		}

		return cmd[0], strings.ToTitle(cmd[1]), cmd[2]
	}

	return cmd[0], strings.ToTitle(cmd[1]), ""
}

func getSafeArg(args []string, nr int) string {
	if nr < len(args) {
		return args[nr]
	}
	return ""
}

func showSucc(s *Session, msg, nub string) {
	if msg == "" {
		write(s, fmt.Sprintf("%s OK success %s", nub, eol), nub)
	} else {
		write(s, fmt.Sprintf("%s %s %s", nub, msg, eol), nub)
	}
}

func showSuccWithData(s *Session, data []string, msg string, nub string) {
	content := ""
	for _, datum := range data {
		content += fmt.Sprintf("%s%s", datum, eol)
	}
	content += fmt.Sprintf("%s OK %s%s", nub, msg, eol)
	write(s, content, nub)
}

func showBad(s *Session, err string, nub string) {
	if nub == "" {
		nub = "*"
	}

	if err == "" {
		write(s, fmt.Sprintf("%s BAD %s", nub, eol), nub)
		return
	}
	write(s, fmt.Sprintf("%s BAD %s%s", nub, err, eol), nub)
}

func showNo(s *Session, msg string, nub string) {
	write(s, fmt.Sprintf("%s NO %s%s", nub, msg, eol), nub)
}

func paramsErr(session *Session, commend string, nub string) {
	write(session, fmt.Sprintf("* BAD %s parameters! %s", commend, eol), nub)
}

func write(session *Session, content string, nub string) {
	if !strings.HasSuffix(content, eol) {
		log.WithContext(session.Ctx).Errorf("Error:返回结尾错误  %s", content)
	}
	log.WithContext(session.Ctx).Debugf("Imap Out:\t |%s", content)
	fmt.Fprintf(session.Conn, content)
}
