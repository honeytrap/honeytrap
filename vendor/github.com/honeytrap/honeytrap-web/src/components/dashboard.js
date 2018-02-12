import React, { Component } from 'react';

import { connect } from 'react-redux';

import Header from './header';
import SessionList from './session-list';
import Earth from './earth';

import View from './view';
import moment from 'moment';
import Flag from "react-flags";

import * as d3 from 'd3';
import * as topojson from 'topojson';

import * as countries from 'i18n-iso-countries';
import { clearHotCountries } from '../actions/index';

import classNames from 'classnames';

class Dashboard extends Component {
    constructor(props) {
        super(props);

        this.state = { now: moment() };
    }

    componentDidMount() {
        this.interval = setInterval(() => this.setState({ now: moment() }), 1000);
    }

    componentWillUnmount() {
        clearInterval(this.interval);
    }

    render() {
        const { start } = this.props.metadata;

        const { now } = this.state;

        var uptime = now.diff(moment(start), 'minutes');

        let hotCountries = this.props.hotCountries.sort((left, right) => {
          return right.count - left.count;
        }).slice(0, 10).map((country, i) => {
            const isocode = country['isocode'];

            return <tr key={i} style={{ fontFamily: 'courier', fontSize: '0.8em' }}>
                <td style={{ border: 'none', padding: '2px' }}>
                    {`${country["count"]}`  }
                </td>
                <td style={{ border: 'none', padding: '2px' }}>
                    <Flag
                        name={isocode}
                        format="png"
                        basePath="images/flags"
                        pngSize={16}
                        shiny={false}
                    />
                </td>
                <td style={{ border: 'none', padding: '2px' }}>
                    {`${ countries.getName(isocode.toUpperCase(), 'en') }`}
                </td>
            </tr>;
        });


        let prev = {};

        // sort on time
        let events = this.props.events.reduce((red, val) => {
            if (prev['source-ip'] == val["source-ip"]
            )
                return red;

            prev = val;

            red.push({ 'source-ip': val['source-ip'], 'destination-port': val['destination-port'], 'category': [ val['category'] ], 'source.country.isocode': val['source.country.isocode'] });
            return red;
        }, []).slice(0, 10).map((event, i) => {
            return <tr key={i} className={ classNames({'show': (20 > i) }) } style={{ fontFamily: 'courier', fontSize: '0.8em' }}>
            <td style={{ border: 'none', padding: '2px' }}>
            <Flag
            name={event['source.country.isocode']}
            format="png"
            basePath="images/flags"
            pngSize={16}
            shiny={false}
            />
            </td>
            <td style={{ border: 'none', padding: '2px' }}>
            { `${event["source-ip"]}` }
                </td>
                <td style={{ border: 'none', padding: '2px' }}>
                    { `${event["category"]}` }
                </td>
                <td style={{ border: 'none', padding: '2px' }}>
                    { `${event["destination-port"]}` }
                </td>
            </tr>
        });

        prev = null;

        let exec = this.props.events.reduce((red, event) => {
            if (!event['ssh.exec'])
                return red;

            const val = event['ssh.exec'][0];

            if (prev == val)
                return red;

            prev = val;

            red.push(val);
            return red;
        }, []).slice(0, 10).map((val, i) => {
            return <li key={i} style={{ fontFamily: 'courier', fontSize: '0.8em' }}>{ val }</li>;
        });

        return (
            <View title="Overview" subtitle="Dashboard">
                <div className="row" style={{ marginTop: '0px', position: 'relative' }}>
                    <div style={{ 'position': 'absolute', 'bottom': '0px', 'left': '0px', 'right': '0px' }}>
                        <div style={{ display: 'block', width: '100%', background: 'black', height: '100%', position: 'absolute', opacity: '0.3' }}></div>
                        <div className='pull-left col-md-4'>
                            <h4 style={{ fontFamily: 'courier' }}>Last attacks</h4>
                            <table className="table table-condensed">
                                <tbody>
                                { events }
                                </tbody>
                            </table>
                        </div>
                        <div className='pull-left col-md-4'>
                            <h4 style={{ fontFamily: 'courier' }}>Origin</h4>
                            <table className="table table-condensed">
                                <tbody>
                                    { hotCountries }
                                </tbody>
                            </table>
                        </div>
                        <div className='pull-left col-md-4'>
                            <h4 style={{ fontFamily: 'courier' }}>Status</h4>
                            <ul className="list-unstyled" >
                                <li style={{ fontFamily: 'courier', fontSize: '0.8em' }}>
                                    uptime:&nbsp;
                                    <span>
                                        { Math.floor(uptime / (60 * 24)) }d &nbsp;
                                        { Math.floor(uptime / 60) }h &nbsp;
                                        { Math.floor(uptime % 60) }m
                                    </span>
                                </li>
                                <li style={{ fontFamily: 'courier', fontSize: '0.8em' }}>
                                    version:&nbsp;
                                    <span>
                                        { this.props.metadata.version }
                                    </span>
                                </li>
                                <li style={{ fontFamily: 'courier', fontSize: '0.8em' }}>
                                    commitid:&nbsp;
                                    <span>
                                        { this.props.metadata.shortcommitid }
                                    </span>
                                </li>
                            </ul>
                        </div>
                    </div>
                    <Earth countries={this.props.hotCountries}></Earth>
                    <ul className="list-unstyled hidden" style={{ position: 'absolute', bottom: '0px' }}>
                        { exec }
                    </ul>
                </div>
            </View>
        );
    }
}

function mapStateToProps(state) {
    return {
        events: state.sessions.events.sort(function (left, right) {
            return moment(right.date).utc().diff(moment(left.date).utc());
        }),
        hotCountries: state.sessions.hotCountries,
        metadata: state.sessions.metadata,
    };
}

export default connect(mapStateToProps)(Dashboard);
