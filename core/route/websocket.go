package route

import (
	log "github.com/kico0909/cgo/core/kernel/logger"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"strconv"
	"time"
)

func (c *RouterManager) RegistorWSpath (path string, f websocketFunc) {
	tmp := &routerChip{
		path:           path,
		regPath:        handlerPathString2regexp(path),
		H:              &RouterHandler{path: path},
		wsFunc: f,
		isWS: true}
	c.Routers = append(c.Routers, tmp)
}

func (c *WSConn) MsgToString() string{
	return string(c.Msg)
}

func (c *WSConn) ReMessage(message string) error {
	return wsutil.WriteServerMessage(c.conn, c.op, []byte(message))
}

func (c *WSConn) IsError() bool {
	return c.Error != nil
}
func (c *WSConn) Close () {
	c.conn.Close()
}
func (c *WSConn) OpenCode () ws.OpCode {
	return c.op
}
func (c *WSConn) GetId () string {
	return c.id
}
func (c *WSConn) ResetId(id string) {
	c.id = id
}

func (s *routerChip ) WsRun() {
	var CONN WSConn
	var err error

	CONN.conn, _, _, err = ws.UpgradeHTTP(s.H.R, s.H.W)
	CONN.id = strconv.FormatInt(time.Now().UnixNano(), 10)
	if err != nil {
		log.Println(" Websocket Error: ", err.Error())
	}

	go func() {
		defer CONN.conn.Close()
		for {
			CONN.Msg, CONN.op, CONN.Error = wsutil.ReadClientData(CONN.conn)

			if CONN.IsError()  {
				s.wsFunc(&CONN)
				break
			}else{
				if len(string(CONN.Msg)) > 0 {
					s.wsFunc(&CONN)
				}
			}
		}
		log.Println("Websocket client link closed: ", CONN.id)
	}()
}

