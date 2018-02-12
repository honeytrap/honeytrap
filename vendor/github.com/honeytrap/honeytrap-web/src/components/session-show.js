import React, { Component } from 'react';
import { connect } from 'react-redux';
import { fetchSession, fetchSessionContent } from '../actions/index';

class SessionShow extends Component {
	constructor() {
		super();

		this.addContentLine = this.addContentLine.bind(this);
		this.line = 0;
		this.totalLines = this.getContent().length;
	}

	componentWillMount() {
		const { dispatch } = this.props;

		dispatch(fetchSession(this.props.params.id));		
		this.startTimeout(1000);
	}

	startTimeout(ms) {
		this.timeOut = setInterval(this.addContentLine, ms);
  	}

  	addContentLine() {
  		const { dispatch } = this.props;

  		if(this.line < this.totalLines) {
  			dispatch(fetchSessionContent(this.getContent()[this.line]))
    		this.line += 1;
    	}
    	if (this.line == this.totalLines) { 
      	clearInterval(this.timeOut);
    	}    	
  	}

	renderContent() {
		const { content } = this.props;

		return (
			content.map((line, i) => {
				return (
					<div key={i}>{line}</div>
				);
			})
		);
	}

	getContent() {
		return [
			"root@honeytrap:~# logout",
			"root@honeytrap:~# exit",
			"root@honeytrap:~# logout",
			"root@honeytrap:~# exit",
			"root@honeytrap:~# logout",
			"root@honeytrap:~# exit",
			"root@honeytrap:~# logout",
			"root@honeytrap:~# exit",
			"root@honeytrap:~# logout",
		]
	}

	render() {
		if(!this.props.session) {
			return (
				<div>Loading...</div>
			);
		}

		return (
			<div>
				<table className="table table-condensed ng-scope" style={{marginTop: "20px", marginBottom: "20px"}}>
					<tbody>
						<tr>
							<th className="col-sm-2">Date</th>
							<td className="dark">{this.props.session.date}</td>
						</tr>
						<tr>
							<th>Origin</th>
							<td className="dark">								
								<span>{this.props.session.location}</span>
							</td>
						</tr>
						<tr>
							<th>Start date</th>
							<td className="dark">{this.props.session.started}</td>
						</tr>
						<tr>
							<th>End date</th>
							<td className="dark">{this.props.session.ended}</td>
						</tr>
						<tr>
							<th>Username</th>
							<td className="dark">{this.props.session.username}</td>
						</tr>
						<tr>
							<th>Password</th>
							<td className="dark">{this.props.session.password}</td>
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
						{this.renderContent()}		
					</pre>
				</div>				
			</div>
		);
	}
}

function mapStateToProps(state) {
	return {
		session: state.sessions.session,
		content: state.sessions.content
	}
}

export default connect(mapStateToProps)(SessionShow);