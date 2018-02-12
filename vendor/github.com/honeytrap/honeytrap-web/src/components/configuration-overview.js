import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { addSession, fetchSessions } from '../actions/index';
import { Link } from 'react-router';
import View from './view';

class ConfigurationOverview extends Component {
	constructor(props) {
		super(props);
	}

	componentWillMount() {
		const { dispatch } = this.props;
	}

	renderTable() {
	}

	render() {
		return (
			<View title="Overview" subtitle="Configuration">
            Configuration
			</View>
		);
	}
}

function mapStateToProps(state) {
	return {
	};
}

export default connect(mapStateToProps)(ConfigurationOverview);
