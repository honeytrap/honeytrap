import React, { Component } from 'react';

class Search extends Component {
    render() {
        return (
            <form className="sidebar-form" >
                <input className="form-control" type="text" placeholder="Search..." />
                <button type="submit" className="btn-link">
                    <span className="icon icon-magnifying-glass"></span>
                </button>
            </form>
        );
    }
}

export default Search;
