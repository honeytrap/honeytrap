import React, { Component } from 'react';

export default class SessionListHeader extends Component {
	render() {
		return (
			<div>				
				<div className="intro text-center">
					<h3 className="page-title">Get in on the action</h3>
					<div className="row">
						<div className="col-xs-12 col-sm-8 col-sm-offset-2">
							<p className="page-lead">Check out some of the live honeypots happening and being recorded right now around the world</p>
						</div>
					</div>
				</div>
			</div>
		);
	}
}