import React, { Component } from 'react';

import { connect } from 'react-redux';

import Header from './header';
import SessionList from './session-list';

import View from './view';
import Flag from "react-flags";

class Agents extends Component {
    constructor(props) {
        super(props);
    }

    componentWillMount() {
        const { dispatch } = this.props;
    }

    render() {
        return (
            <View title="Overview" subtitle="Agents">
                Agents
            </View>
        );
    }
}

function mapStateToProps(state) {
    return {
    };
}

export default connect(mapStateToProps)(Agents);
