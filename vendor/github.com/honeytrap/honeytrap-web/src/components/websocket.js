import React, { Component } from 'react';
import { connect } from 'react-redux';

// import Socket from '../utils/socket'
import { default as Socket } from '../utils/socket';

class Websocket extends Component {
    componentWillMount() {
	      const { dispatch } = this.props;

        let url;

        if (process.env.WEBSOCKET_URI) {
            url = process.env.WEBSOCKET_URI;
        } else {
            const {location} = window;
            url = ((location.protocol === "https:") ? "wss://" : "ws://") + location.host + "/ws";
        }

        let socket = new Socket(url);
        socket.startWS(dispatch);
    }

    render() {
        return null;
    }
}

function select(state, ownProps) {
    return {
        ...ownProps
    };
}

export default connect(select)(Websocket);
