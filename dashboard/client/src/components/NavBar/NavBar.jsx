import React, { Component } from 'react';
import { NavLink as NavLinkRouter, withRouter, matchPath } from 'react-router-dom';
import {
  Navbar,
  NavbarBrand,
  NavbarToggler,
  Nav,
  NavItem,
  NavLink,
  Collapse,
} from 'reactstrap';
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faGithub } from '@fortawesome/free-brands-svg-icons';
import { faSignOutAlt } from '@fortawesome/free-solid-svg-icons';

class NavBarWithRouter extends Component {
  state = {
    isActive: false,
  };

  toggle() {
    this.setState({
      isActive: !this.state.isActive,
    })
  }

  toggle = this.toggle.bind(this);

  isLoggedIn() {
    return window.SIGNED_IN === 'true'
  }

  createNavLink(currentPath, path, label, additionalClassName, icon = null, onClick = () => {}) {
    if (!path) {
      return null;
    }

    const to = `/${path}`;

    const className = [
      'py-3',
      'px-3',
      'px-md-2',
      additionalClassName,
    ].filter(item => item).join(' ');

    return (
      <NavItem active={currentPath === to}>
        <NavLink className={className} tag={NavLinkRouter} to={to} exact onClick={onClick}>
          { icon }
          { label }
        </NavLink>
      </NavItem>
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

    return (
      <Navbar
        dark
        color="openfaas"
        expand="md"
        className={[
          'background-openfaas-important',
          'border-radius-bottom-5',
          'navbar-inverse',
          'p-0',
          'overflow-hidden',
        ].join(' ')}
      >
        <NavbarBrand href="/" className="font-size-0-important margin-0-important p-0 pl-2">
          <img
            alt="OpenFaaS"
            src="https://docs.openfaas.com/images/logo.svg"
          />
        </NavbarBrand>
        <a
          href="https://docs.openfaas.com/openfaas-cloud/intro"
          className="color-white py-3 px-2"
        >
          OpenFaaS Cloud
        </a>
        <NavbarToggler className="mr-2" onClick={this.toggle} />
        <Collapse isOpen={this.state.isActive} navbar>
          <Nav navbar>
            { this.createNavLink(pathname, user, 'Home') }
            <NavItem>
              <NavLink
                className="py-3 px-3 px-md-2"
                href="https://github.com/openfaas/openfaas-cloud"
              >
                <FontAwesomeIcon icon={faGithub} className="mr-1" />
                GitHub
              </NavLink>
            </NavItem>
          </Nav>
          <Nav navbar className="ml-auto">
            { this.isLoggedIn() && this.createNavLink(
                pathname,
                'logout',
                'Logout',
                '',
                <FontAwesomeIcon icon={faSignOutAlt} className="mr-1" />,
                this.forceUpdate,
            ) }
          </Nav>
        </Collapse>
      </Navbar>
    );
  }
}

const NavBar = withRouter(NavBarWithRouter);

export { NavBar };
