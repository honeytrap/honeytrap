import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { addSession, fetchSessions } from '../actions/index';
import SessionListHeader from '../components/session-list-header';
import { Link } from 'react-router';

class SessionList extends Component {
	constructor(props) {
		super(props);
		
		this.onClickAddSession = this.onClickAddSession.bind(this);
	}

	componentWillMount() {
		const { dispatch } = this.props;

		dispatch(fetchSessions());
	}

	renderTable() {
		if(!this.props.sessions) {
			return (
				<div>Loading...</div>
			);
		}

		const { sessions } = this.props;

		return sessions.map((session, i) => {
			return (				
					<tr key={i}>
						<td><Link to={"/session/" + session.id}>{session.id}</Link></td>
						<td>{session.date}</td>
						<td>{session.location}</td>
						<td>{session.started}</td>
						<td>{session.ended}</td>
						<td>{session.username}</td>
						<td>{session.password}</td>
					</tr>
				
			);
		})
	}

	makeId()
	{
    	var text = "";
    	var possible = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-";

    	for( var i=0; i < 36; i++ )
        	text += possible.charAt(Math.floor(Math.random() * possible.length));

    	return text;
	}

	onClickAddSession() {
		const { dispatch } = this.props;
    console.debug("test");

		dispatch(addSession(this.makeId()));
	}

	render() {
		return (
			<div>
				<div className="block-faded">
					<table className="table table-generic">
						<thead>
							<tr>
								<th className="header">Session</th>
								<th className="header">Date</th>
						<th className="header">Location</th>
								<th className="header">Started</th>
								<th className="header">Ended</th>
								<th className="header">Username</th>
								<th className="header">Password</th>
							</tr>
						</thead>
						<tbody>
							{this.renderTable()}							
						</tbody>
					</table>
				</div>
				<button className="b-button" onClick={this.onClickAddSession}>Add random session</button>
			</div>
		);
	}
}

function mapStateToProps(state) {
	return {
		sessions: state.sessions.all
	};
}

export default connect(mapStateToProps)(SessionList);
