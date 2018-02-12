import React, { Component } from 'react';

class Server extends Component {
  componentDidMount() {
    this.connection = new WebSocket('ws://localhost:1234', 'echo-protocol');

    this.connection.onopen = () => {
      console.log('connection opened!');
    }

    this.connection.onclose = () => {
      console.log('connection closed');
    }

    this.setState ({socket: this.connection})

  }

  componentWillUnmount() {
    this.state.socket.close();
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
        <h1>Serverish</h1>
        <button onClick={() => this.state.socket.send(JSON.stringify(this.exampleBroadcast()))}>Broadcast</button>
      </div>
    );
  }
}

export default Server;