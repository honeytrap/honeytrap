import React, { Component } from 'react';
import { connect } from 'react-redux';

import Header from './header';

import Dashboard from './dashboard';
import Attacks from './attacks';
import Events from './events';
import Agents from './agents';

import Navigation from './navigation';
import Search from './search';

import SessionList from './session-list';
import ConfigurationOverview from './configuration-overview';
import NotFoundPage from './not-found';

import { HashRouter, BrowserRouter, Redirect, Switch, Route } from 'react-router-dom';

class App extends Component {
    render() {
        if (!this.props.metadata)
            return <div className="loading">Loading&#8230;</div>;

        let disconnected = null;
        if (!this.props.connected) {
            disconnected = (
                <div className="alert alert-danger" role="alert">
                    Connection with sensor has been lost.
                </div>
            );
        }

        let versionAvailable = null;
        if (false) {
            versionAvailable =
                <div className="alert alert-warning" role="alert">
                    New version available. <a>Upgrade</a>.
                </div>;
        }

        return (
            <div className="container">
                <div className="row">
                    { disconnected }
                    <div className="col-sm-3 sidebar">
                        <nav className="sidebar-nav">
                            <div className="sidebar-header">
                                <a className="sidebar-brand img-responsive" href="/">
                                    <span className="icon">Honeytrap</span>
                                </a>
                            </div>
                            <div className="collapse nav-toggleable-sm" id="nav-toggleable-sm">
                                { versionAvailable }
                                <Navigation />
                                <hr className="visible-xs m-t" />
                            </div>
                        </nav>
                    </div>
                        <Switch>
                            <Route exact path="/" component={Dashboard} / >
                                <Route exact path="/agents" component={Agents} />
                                <Route exact path="/events" component={Events} />
                                <Route exact path="/configuration" component={ConfigurationOverview} />
                                <Route path="/404" component={NotFoundPage} />
                                <Redirect from='*' to='/404' />
                        </Switch>
                </div>
            </div>
        );
	  }
}

function mapStateToProps(state) {
    return {
        connected: state.sessions.connected,
        metadata: state.sessions.metadata
    };
}

export default connect(mapStateToProps)(App);
