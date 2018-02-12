import React, { Component } from 'react';

import { connect } from 'react-redux';

import Header from './header';
import SessionList from './session-list';

import View from './view';
import moment from 'moment';

import * as d3 from 'd3';
import * as topojson from 'topojson';

import Color from 'color';
import classNames from 'classnames';

import { fetchCountries, clearHotCountries } from '../actions/index';

function darken(col, amt) {
    var usePound = false;

    if (col[0] == "#") {
        col = col.slice(1);
        usePound = true;
    }

    var num = parseInt(col,16);

    var r = (num >> 16) + amt;

    if (r > 255) r = 255;
    else if  (r < 0) r = 0;

    var b = ((num >> 8) & 0x00FF) + amt;

    if (b > 255) b = 255;
    else if  (b < 0) b = 0;

    var g = (num & 0x0000FF) + amt;

    if (g > 255) g = 255;
    else if (g < 0) g = 0;

    return (usePound?"#":"") + (g | (b << 8) | (r << 16)).toString(16);
}

class Earth extends React.Component {
    constructor() {
        super();

        this.hotCountries = [];

        this.state = {
            angle: 90,
            countries: [],
            loading: true,
        };
    }

    componentDidMount() {
        let { dispatch } = this.props;

        dispatch(fetchCountries()).then(() => {
            this.setState({loading: false});
        });

        let canvas = this.refs.canvas;
        const context = canvas.getContext('2d');

        const angle = 90;
        this.projection = d3.geoOrthographic()
            .clipAngle(angle);

        var drag = d3.drag()
                     .on('drag', () => {
                         var dx = d3.event.dx;
                         var dy = d3.event.dy;

                         var rotation = this.projection.rotate();
                         var radius = this.projection.scale();

                         var scale = d3.scaleLinear()
                                       .domain([-1 * radius, radius])
                                       .range([-90, 90]);

                         var degX = scale(dx);
                         var degY = scale(dy);

                         rotation[0] += degX;
                         rotation[1] -= degY;

                         if (rotation[1] > 90)   rotation[1] = 90;
                         if (rotation[1] < -90)  rotation[1] = -90;
                         if (rotation[0] >= 180) rotation[0] -= 360;

                         this.projection.rotate(rotation);

                         this.updateCanvas();
                     });

        const { width, height } = canvas;

        var zoom = d3.zoom()
                .scaleExtent([200, 2000]);

        zoom
            .on('zoom', (event) => {
                console.log(d3.event.transform);
                this.projection.scale(d3.event.transform.k, d3.event.transform.k);
                this.updateCanvas();
            });

        d3.select(this.refs.canvas).call(drag);
        d3.select(this.refs.canvas).call(zoom);

        window.addEventListener("resize", () => this.updateDimensions);
    }

    componentWillUnmount() {
        window.removeEventListener("resize", () => this.updateDimensions);
    }
 
    componentWillReceiveProps(nextProps, nextState) {
        if (!nextProps.countries.length)
            return;

        const { countries } = nextProps.topology;
        if (!countries)
            return;

        this.updateCanvas();

        this.hotCountries = nextProps.countries.reduce((red, value) => {
            let country = countries.find((v) => {
                return v.iso_a2 == value.isocode;
            });

            if (!country)
                return red;

            red.push({
                ...value,
                ...country,
            });

            return red;
        }, []);

        if (this.hotCountries.length == 0)
            return;

        // sort on time
        this.hotCountries.sort((left, right) => {
            return moment(left.last).utc().diff(moment(right.last).utc());
        });


        let last = this.hotCountries[this.hotCountries.length - 1]

        const p = d3.geoCentroid(last);

        let projection = this.projection;

        d3.transition()
          .duration(2500)
          .tween("rotate", () => {
              var r = d3.interpolate(projection.rotate(), [-p[0], -p[1]]);
              return (t) => {
                  projection.rotate(r(t));

                  this.updateCanvas();
              };
          });

        return
    }

    updateDimensions() {
        var w = window,
            d = document,
            documentElement = d.documentElement,
            body = d.getElementsByTagName('body')[0],
            width = w.innerWidth || documentElement.clientWidth || body.clientWidth,
            height = w.innerHeight|| documentElement.clientHeight|| body.clientHeight;

        this.setState({width: width, height: height});
    }

    updateCanvas() {
        requestAnimationFrame(() => {
            let canvas = this.refs.canvas;
            if (!canvas)
                return;

            const context = canvas.getContext('2d');

            let path = d3.geoPath().
                context(context).
                projection(
                    this.projection
                        .translate([canvas.width/2, (canvas.height * (5/12))])
                );

            context.clearRect(0, 0, canvas.width, canvas.height);

            context.beginPath();
            path({type: 'Sphere'});
            context.fillStyle = '#1b202d';
            context.fill();

            const { world } = this.props.topology;
            if (!world)
                return;

            var land = topojson.feature(world, world.objects.land);
            context.beginPath();
            path(land);
            context.fillStyle = 'white';
            context.fill();

            context.strokeStyle = 'gray';
            context.stroke();

            const total = this.hotCountries.reduce((total, country) => {
                total += country.count;
                return (total);
            }, 0);

            this.hotCountries.forEach((country) => {
                const min = 1 + moment().diff(country.last, 'minutes');
                let color = Color('#440000');
                context.beginPath();
                color = color.lighten((Math.log(min) * 50) * (country.count / total));
                context.fillStyle = color.hexString();
                path(country);
                context.fill();
            });

            context.beginPath();
            context.fillStyle = 'white';
            path(topojson.mesh(world));
            context.stroke();

            /*
              var circle = d3.geo.circle().angle(5).origin([-0.8432, 51.4102]);
              circles = [];
              circles.push( circle() );
              circle.origin([-122.2744, 37.7561]);
              circles.push( circle() );
              context.fillStyle = "rgba(0,100,0,.5)";
              context.lineWidth = .2;
              context.strokeStyle = "#000";
              context.beginPath();
              path({type: "GeometryCollection", geometries: circles});
              context.fill();
              context.stroke();
            */

            /*
              context.lineWidth = 2;
              context.strokeStyle = "rgba(0,100,0,.7)";
              context.beginPath();
              path({type: "LineString", coordinates: [[-0.8432, 51.4102],[-122.2744, 37.7561]] });
              context.stroke();
            */
        });
    }

    render() {
        const { loading } = this.state;

        return (
            <div>
                <canvas className={ classNames({ 'hidden': loading }) } style={{ 'cursor': 'move' }} ref="canvas" width={900} height={800}/>
            </div>
        );
    }
}

function mapStateToProps(state) {
    return {
        topology: state.sessions.topology,
    };
}

export default connect(mapStateToProps)(Earth);
