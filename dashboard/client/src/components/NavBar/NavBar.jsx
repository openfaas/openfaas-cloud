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

  createNavLink(pathname, user) {
    if (!user) {
      return null;
    }

    const to = `/${user}`;

    return (
      <NavItem active={pathname === to}>
        <NavLink className="py-3 px-3 px-md-2" tag={NavLinkRouter} to={to} exact>
          Home
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
            { this.createNavLink(pathname, user) }
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
        </Collapse>
      </Navbar>
    );
  }
}

const NavBar = withRouter(NavBarWithRouter);

export { NavBar };
