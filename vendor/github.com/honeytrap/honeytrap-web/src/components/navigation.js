import React, { Component } from 'react';
import { Link, NavLink } from 'react-router-dom';

class Navigation extends Component {
    render() {
        return (
            <ul className="nav nav-pills nav-stacked">
                <li className="nav-header">Dashboard</li>
                <li>
                <NavLink activeClassName="active" to="/">Overview</NavLink>
                </li>
                <li className="nav-header">Events</li>
                <li>
                <NavLink activeClassName="active" to="/events">Overview</NavLink>
                </li>
                <li className="nav-header">Agents</li>
                <li>
                <NavLink activeClassName="active" to="/agents">Overview</NavLink>
                </li>
                <li className="nav-header">Configuration</li>
                <li>
                <NavLink activeClassName="active" to="/configuration/">Overview</NavLink>
                </li>
                <li className="nav-header">Other</li>
                <li >
                    <a href="https://honeytrap.io/">
                        honeytrap.io
                    </a>
                </li>
            </ul>
        );
    }
}

export default Navigation;
