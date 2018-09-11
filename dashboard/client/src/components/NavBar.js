import React, { Component } from 'react';
import './NavBar.css';

import { NavLink, withRouter, matchPath } from 'react-router-dom';

class NavBarWithRouter extends Component {
  createNavLink(pathname, user) {
    if (!user) {
      return null;
    }

    const to = `/${user}`;
    return (
      <li className={pathname === to ? 'active' : null}>
        <NavLink to={`/${user}`} exact>
          Home
        </NavLink>
      </li>
    );
  }

  render() {
    const { pathname } = this.props.history.location;
    const match = matchPath(pathname, {
      path: '/:user',
      strict: false,
    });
    let user;
    if (match && match.params) {
      user = match.params.user;
    }

    const navLink = this.createNavLink(pathname, user);
    return (
      <nav className="navbar navbar-inverse">
        <div className="container-fluid">
          <div className="navbar-header">
            <a className="navbar-brand" href="#">
              <img
                alt="OpenFaaS"
                src="https://docs.openfaas.com/images/logo.svg"
              />
            </a>
            <p className="navbar-text">
              <a href="#">OpenFaaS Cloud</a>
            </p>
          </div>
          <ul className="nav navbar-nav">{navLink}</ul>
        </div>
      </nav>
    );
  }
}

export const NavBar = withRouter(NavBarWithRouter);
