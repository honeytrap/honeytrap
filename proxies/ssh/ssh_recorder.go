package proxies

import (
	"fmt"
	"time"

	"github.com/honeytrap/honeytrap/proxies"
	"github.com/honeytrap/honeytrap/pushers"

	"github.com/satori/go.uuid"
)

type SSHAction struct {
	Time        time.Time         `json:"date"`
	ReceiveDate time.Time         `json:"receive_date"`
	StartDate   time.Time         `json:"start_date"`
	EndDate     time.Time         `json:"end_date,omitempty"`
	Sequence    int               `json:"sequence,omitempty"`
	SessionID   string            `json:"session_id,omitempty"`
	ChannelID   string            `json:"channel_id,omitempty"`
	ContainerID string            `json:"container_id,omitempty"`
	Sensor      string            `json:"sensor,omitempty"`
	Username    string            `json:"username,omitempty"`
	Password    string            `json:"password,omitempty"`
	Client      string            `json:"client,omitempty"`
	KeyType     string            `json:"key_type,omitempty"`
	Key         string            `json:"key,omitempty"`
	RemoteAddr  string            `json:"remote_address,omitempty"`
	Payload     []byte            `json:"cast,omitempty"`
	Protocol    string            `json:"protocol,omitempty"`
	Meta        map[string]string `json:"meta,omitempty"`
}

// this is the global recorder
func NewSSHRecorder(p *pushers.Pusher) *SSHRecorder {
	// contains info about the container
	return &SSHRecorder{p}
}

type SSHRecorder struct {
	*pushers.Pusher
}

type SSHRecorderSession struct {
	r         *SSHRecorder
	seq       int
	sessionID uuid.UUID
	startDate time.Time
	endDate   time.Time
	conn      *proxies.ProxyConn
	username  string
	password  string
}

func (r *SSHRecorder) NewSession(c *proxies.ProxyConn) *SSHRecorderSession {
	sessionID := uuid.NewV4()
	startDate := time.Now()
	return &SSHRecorderSession{conn: c, sessionID: sessionID, seq: 0, r: r, startDate: startDate}
}

func (rs *SSHRecorderSession) Connect() {
	rs.r.Push("ssh", "connect", rs.conn.Container.Name(), rs.sessionID.String(), &SSHAction{ContainerID: rs.conn.Container.Name(), RemoteAddr: rs.conn.RemoteHost(), SessionID: rs.sessionID.String(), Sequence: rs.seq, Sensor: "connect", Time: time.Now(), StartDate: rs.startDate, Payload: nil})
	rs.seq++
}

func (rs *SSHRecorderSession) Start() {
	rs.r.Push("ssh", "Session-Open-Packet", rs.conn.Container.Name(), rs.sessionID.String(), &SSHAction{ContainerID: rs.conn.Container.Name(), RemoteAddr: rs.conn.RemoteHost(), SessionID: rs.sessionID.String(), Sequence: rs.seq, Sensor: "Session-Open-packet", Time: time.Now(), StartDate: rs.startDate, Payload: nil})
	rs.seq++
}

func (rs *SSHRecorderSession) AuthorizationPublicKey(username, keyType string, key []byte) {
	rs.r.Push("ssh", "Session-Authentication-PublicKey", rs.conn.Container.Name(), rs.sessionID.String(), &SSHAction{ContainerID: rs.conn.Container.Name(), ChannelID: "", Username: username, KeyType: keyType, Key: fmt.Sprintf("%x", key), RemoteAddr: rs.conn.RemoteHost(), SessionID: rs.sessionID.String(), Sequence: rs.seq, Sensor: "Session-Authentication-PublicKey", Time: time.Now(), StartDate: rs.startDate, Payload: nil})
	rs.seq++
}

func (rs *SSHRecorderSession) AuthorizationSuccess(username, password, client string) {
	rs.r.Push("ssh", "Session-Authentication-Success", rs.conn.Container.Name(), rs.sessionID.String(), &SSHAction{ContainerID: rs.conn.Container.Name(), ChannelID: "", Username: username, Password: password, RemoteAddr: rs.conn.RemoteHost(), SessionID: rs.sessionID.String(), Sequence: rs.seq, Sensor: "Session-Authentication-Success", Time: time.Now(), StartDate: rs.startDate, Client: client, Payload: nil})
	rs.seq++
	rs.username = username
	rs.password = password
}

func (rs *SSHRecorderSession) AuthorizationFailed(username, password, client string) {
	rs.r.Push("ssh", "Session-Authentication-Failed", rs.conn.Container.Name(), rs.sessionID.String(), &SSHAction{ContainerID: rs.conn.Container.Name(), ChannelID: "", Username: username, Password: password, RemoteAddr: rs.conn.RemoteHost(), SessionID: rs.sessionID.String(), Sequence: rs.seq, Sensor: "Session-Authentication-Failed", Time: time.Now(), StartDate: rs.startDate, Client: client, Payload: nil})
	rs.seq++
}

func (rs *SSHRecorderSession) Data(sensor string, channelID uuid.UUID, payload []byte) {
	data := make([]byte, len(payload))
	copy(data, payload)
	rs.r.Push("ssh", sensor, rs.conn.Container.Name(), rs.sessionID.String(), &SSHAction{ContainerID: rs.conn.Container.Name(), ChannelID: channelID.String(), Username: rs.username, Password: rs.password, RemoteAddr: rs.conn.RemoteHost(), SessionID: rs.sessionID.String(), Sequence: rs.seq, Sensor: sensor, Time: time.Now(), StartDate: rs.startDate, Payload: data})
	rs.seq++
}

func (rs *SSHRecorderSession) CustomData(tag string, payload []byte) {
	rs.r.Push("ssh", tag, rs.conn.Container.Name(), rs.sessionID.String(), &SSHAction{ContainerID: rs.conn.Container.Name(), RemoteAddr: rs.conn.RemoteHost(), Username: rs.username, Password: rs.password, SessionID: rs.sessionID.String(), Sequence: 0, Sensor: tag, Time: time.Now(), StartDate: rs.startDate, Payload: payload})
}

func (rs *SSHRecorderSession) Stop() {
	rs.r.Push("ssh", "Session-Closed-packet", rs.conn.Container.Name(), rs.sessionID.String(), &SSHAction{ContainerID: rs.conn.Container.Name(), RemoteAddr: rs.conn.RemoteHost(), Username: rs.username, Password: rs.password, SessionID: rs.sessionID.String(), Sequence: rs.seq, Sensor: "Session-Closed-packet", Time: time.Now(), StartDate: rs.startDate, EndDate: time.Now(), Payload: nil})
	rs.seq++
}

type SSHRecordSessionWriter struct {
	*SSHRecorderSession

	tag string
}

func (r *SSHRecordSessionWriter) Write(b []byte) (int, error) {
	r.CustomData(r.tag, b)
	return len(b), nil
}

func NewSSHRecordSessionWriter(tag string, r *SSHRecorderSession) *SSHRecordSessionWriter {
	return &SSHRecordSessionWriter{r, tag}
}
