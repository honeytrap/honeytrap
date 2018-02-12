import React, { Component } from 'react';

import { connect } from 'react-redux';

import Header from './header';
import SessionList from './session-list';

import View from './view';

class Attacks extends Component {
    constructor(props) {
        super(props);
    }

    componentWillMount() {
        const { dispatch } = this.props;
    }

    renderTable() {
        if(!this.props.events) {
            return (
                <div>Loading...</div>
            );
        }

        const { events } = this.props;


        return events.map((event, i) => {
            const message = (event.message || event.payload );
            
            return (				
                    <tr key={i}>
                        <td>{event.date.format('lll')}</td>
                        <td>{event.sensor}</td>
                        <td>{event.category}</td>
                        <td>{event["source-ip"] } ({event["source-port"] })</td>
                        <td>{event["destination-ip"] } ({event["destination-port"] })</td>
                        <td>{message}</td>
                    </tr>
                
            );
        });
    }

    render() {
        const events = this.renderTable();

        return (
            <View title="Overview" subtitle="Attacks">
                <table className="table">
                    <thead>
                        <tr>
                            <th className="header">Date</th>
                            <th className="header">Sensor</th>
                            <th className="header">Category</th>
                            <th className="header">Source</th>
                            <th className="header">Destination</th>
                            <th className="header">Message</th>
                        </tr>
                    </thead>
                    <tbody>
                        { events }
                    </tbody>
                </table>
            </View>
        );
    }
}

function mapStateToProps(state) {
    return {
        events: state.sessions.events
    };
}

export default connect(mapStateToProps)(Attacks);
