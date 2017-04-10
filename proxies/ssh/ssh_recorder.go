package proxies

import (
	"fmt"
	"time"

	"github.com/honeytrap/honeytrap/proxies"
	"github.com/honeytrap/honeytrap/pushers"

	"github.com/satori/go.uuid"
)

// SSHAction defines a action for the SSH connection stream.
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

// NewSSHRecorder returns a new instance of the SSHRecorder.
func NewSSHRecorder(p *pushers.Pusher, e pushers.Events) *SSHRecorder {
	// contains info about the container
	return &SSHRecorder{p, e}
}

// SSHRecorder defines a recorder for handling ssh connections.
type SSHRecorder struct {
	*pushers.Pusher
	events pushers.Events
}

// SSHRecorderSession defines a struct to use the underline SSHRecorder for a giving
// ssh session.
type SSHRecorderSession struct {
	r         *SSHRecorder
	seq       int
	sessionID uuid.UUID
	startDate time.Time
	endDate   time.Time
	conn      *proxies.ProxyConn
	username  string
	password  string
	events    pushers.Events
}

// NewSession creates a new session session recorder.
func (r *SSHRecorder) NewSession(c *proxies.ProxyConn) *SSHRecorderSession {
	sessionID := uuid.NewV4()
	startDate := time.Now()
	return &SSHRecorderSession{conn: c, sessionID: sessionID, seq: 0, r: r, startDate: startDate, events: r.events}
}

// Connect records the connect operation for the underline ssh connection.
func (rs *SSHRecorderSession) Connect() {
	rs.r.Push("ssh", "connect", rs.conn.Container.Name(), rs.sessionID.String(), &SSHAction{ContainerID: rs.conn.Container.Name(), RemoteAddr: rs.conn.RemoteHost(), SessionID: rs.sessionID.String(), Sequence: rs.seq, Sensor: "connect", Time: time.Now(), StartDate: rs.startDate, Payload: nil})
	rs.seq++
}

// Start records the start operation for the underline ssh connection.
func (rs *SSHRecorderSession) Start() {
	rs.r.Push("ssh", "Session-Open-Packet", rs.conn.Container.Name(), rs.sessionID.String(), &SSHAction{ContainerID: rs.conn.Container.Name(), RemoteAddr: rs.conn.RemoteHost(), SessionID: rs.sessionID.String(), Sequence: rs.seq, Sensor: "Session-Open-packet", Time: time.Now(), StartDate: rs.startDate, Payload: nil})
	rs.seq++
}

// AuthorizationPublicKey records the publickey authroization operation for the
// underline ssh connection.
func (rs *SSHRecorderSession) AuthorizationPublicKey(username, keyType string, key []byte) {
	rs.r.Push("ssh", "Session-Authentication-PublicKey", rs.conn.Container.Name(), rs.sessionID.String(), &SSHAction{ContainerID: rs.conn.Container.Name(), ChannelID: "", Username: username, KeyType: keyType, Key: fmt.Sprintf("%x", key), RemoteAddr: rs.conn.RemoteHost(), SessionID: rs.sessionID.String(), Sequence: rs.seq, Sensor: "Session-Authentication-PublicKey", Time: time.Now(), StartDate: rs.startDate, Payload: nil})
	rs.seq++
}

// AuthorizationSuccess records the publickey authroization success operation for
// the underline ssh connection.
func (rs *SSHRecorderSession) AuthorizationSuccess(username, password, client string) {
	rs.r.Push("ssh", "Session-Authentication-Success", rs.conn.Container.Name(), rs.sessionID.String(), &SSHAction{ContainerID: rs.conn.Container.Name(), ChannelID: "", Username: username, Password: password, RemoteAddr: rs.conn.RemoteHost(), SessionID: rs.sessionID.String(), Sequence: rs.seq, Sensor: "Session-Authentication-Success", Time: time.Now(), StartDate: rs.startDate, Client: client, Payload: nil})
	rs.seq++
	rs.username = username
	rs.password = password
}

// AuthorizationFailed records the publickey authroization failure operation for
// the underline ssh connection.
func (rs *SSHRecorderSession) AuthorizationFailed(username, password, client string) {
	rs.r.Push("ssh", "Session-Authentication-Failed", rs.conn.Container.Name(), rs.sessionID.String(), &SSHAction{ContainerID: rs.conn.Container.Name(), ChannelID: "", Username: username, Password: password, RemoteAddr: rs.conn.RemoteHost(), SessionID: rs.sessionID.String(), Sequence: rs.seq, Sensor: "Session-Authentication-Failed", Time: time.Now(), StartDate: rs.startDate, Client: client, Payload: nil})
	rs.seq++
}

// Data records the ssh data payload operation for the underline ssh connection.
func (rs *SSHRecorderSession) Data(sensor string, channelID uuid.UUID, payload []byte) {
	data := make([]byte, len(payload))
	copy(data, payload)
	rs.r.Push("ssh", sensor, rs.conn.Container.Name(), rs.sessionID.String(), &SSHAction{ContainerID: rs.conn.Container.Name(), ChannelID: channelID.String(), Username: rs.username, Password: rs.password, RemoteAddr: rs.conn.RemoteHost(), SessionID: rs.sessionID.String(), Sequence: rs.seq, Sensor: sensor, Time: time.Now(), StartDate: rs.startDate, Payload: data})
	rs.seq++
}

// CustomData records the ssh custom data payload operation for the underline ssh connection.
func (rs *SSHRecorderSession) CustomData(tag string, payload []byte) {
	rs.r.Push("ssh", tag, rs.conn.Container.Name(), rs.sessionID.String(), &SSHAction{ContainerID: rs.conn.Container.Name(), RemoteAddr: rs.conn.RemoteHost(), Username: rs.username, Password: rs.password, SessionID: rs.sessionID.String(), Sequence: 0, Sensor: tag, Time: time.Now(), StartDate: rs.startDate, Payload: payload})
}

// Stop records the stop  call for the underline ssh connection.
func (rs *SSHRecorderSession) Stop() {
	rs.r.Push("ssh", "Session-Closed-packet", rs.conn.Container.Name(), rs.sessionID.String(), &SSHAction{ContainerID: rs.conn.Container.Name(), RemoteAddr: rs.conn.RemoteHost(), Username: rs.username, Password: rs.password, SessionID: rs.sessionID.String(), Sequence: rs.seq, Sensor: "Session-Closed-packet", Time: time.Now(), StartDate: rs.startDate, EndDate: time.Now(), Payload: nil})
	rs.seq++
}

// SSHRecordSessionWriter defines a writer for the ssh session.
type SSHRecordSessionWriter struct {
	*SSHRecorderSession

	tag string
}

// Write writes the provided bytes into the underline session recorder.
func (r *SSHRecordSessionWriter) Write(b []byte) (int, error) {
	r.CustomData(r.tag, b)
	return len(b), nil
}

// NewSSHRecordSessionWriter returns a new instance of a session write recorder.
func NewSSHRecordSessionWriter(tag string, r *SSHRecorderSession) *SSHRecordSessionWriter {
	return &SSHRecordSessionWriter{r, tag}
}
