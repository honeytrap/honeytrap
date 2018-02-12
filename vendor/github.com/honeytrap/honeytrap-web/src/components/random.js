import React, { Component } from 'react';

class Socket extends Component {
  constructor() {
    super();

    this.state = ({ messages: [], socket: null });
  }

  componentDidMount() {
    this.connection = new WebSocket('ws://localhost:1234', 'echo-protocol');
    console.log(this.connection);

    this.connection.onopen = (data) => {
      console.log('connection opened!');
      this.connection.send(JSON.stringify({ type: "id" }));
    }

    this.connection.onclose = () => {
      console.log('connection closed');
    }

    this.setState ({socket: this.connection})

    this.connection.onmessage = (message) => { 
      console.log(message.type);
      this.setState({
        messages : this.state.messages.concat([ message.data ])
      })
    }
  }

  componentWillUnmount() {
    this.state.socket.close();
  }

  renderMessages() {
    return this.state.messages.map((message, i) => {
      return (
        <li key={i}>{message}</li>
      );
    })
  }

  exampleMessage() {
    return {
      type: "message",
      data: "sample text"
    }
  }

  exampleBroadcast() {
    return {
      type: "broadcast",
      data: "sample text"
    }
  }

  render() {
    return (
      <div>
        <table className="table table-condensed ng-scope" style={{marginTop: "20px", marginBottom: "20px"}}>
          <tbody>
            <tr>
              <th className="col-sm-2">Date</th>
              <td className="dark">10/02/2017</td>
            </tr>
            <tr>
              <th>Origin</th>
              <td className="dark">               
                <span>unknown</span>
              </td>
            </tr>
            <tr>
              <th>Start date</th>
              <td className="dark">10/02/2017 10:10</td>
            </tr>
            <tr>
              <th>End date</th>
              <td className="dark"> 10/02/2017 10:11</td>
            </tr>
            <tr>
              <th>Username</th>
              <td className="dark">root</td>
            </tr>
            <tr>
              <th>Password</th>
              <td className="dark">root</td>
            </tr>
            <tr>
              <td colSpan="2"><br /></td>
            </tr>
            <tr>
              <th colSpan="2">Request types:</th>
            </tr>
            <tr>
              <td>pty-req</td>
              <td className="dark">xterm-256colorx;û%%ÿÿ !"#$'()2356789:;=>FHIJKZ[\]</td>
            </tr>
            <tr>
              <td>shell</td>
              <td className="dark"></td>
            </tr>
            <tr>
              <td>exit-status</td>
              <td className="dark"></td>
            </tr>
          </tbody>
        </table>
        <div className="terminal">
          <pre>
            Welcome to Ubuntu 14.04.5 LTS (GNU/Linux 4.4.0-31-generic x86_64)<br />
            <br />
            &nbsp;* Documentation:  https://help.ubuntu.com/<br />
            Last login: Fri Feb 10 08:25:09 2017 from 10.0.3.1<br />      
            {this.renderMessages()}
          </pre>
        </div>
        <button onClick={() => this.state.socket.send(JSON.stringify(this.exampleMessage()))}>Send</button>
        <button onClick={() => this.state.socket.send(JSON.stringify(this.exampleBroadcast()))}>Broadcast</button>
      </div>
    );
  }
}

export default Socket;