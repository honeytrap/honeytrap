import React, { Component } from 'react';
import { Link } from 'react-router';

class Header extends Component {
    render() {
        return (
                <div className="dashhead-titles">
                    <h6 className="dashhead-subtitle">{ this.props.subtitle }</h6>
                    <h2 className="dashhead-title">{ this.props.title }</h2>
                </div>
		    );
	  }
}

export default Header;
