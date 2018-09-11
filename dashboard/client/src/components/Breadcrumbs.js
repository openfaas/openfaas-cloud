import React, { Component } from 'react';
import queryString from 'query-string';

import { NavLink, withRouter } from 'react-router-dom';

import './Breadcrumbs.css';

class BreadcrumbsWithRouter extends Component {
  constructor(props) {
    super(props);
    console.log(props);
    let parts = this.pathToParts(
      props.location.pathname,
      props.location.search
    );
    this.state = {
      parts,
    };
  }
  pathToParts(pathname, search = '') {
    let parts = pathname
      .replace(/\/$/, '')
      .split('/')
      .slice(1);
    return parts.reduce((acc, p, i) => {
      let previousPath;
      if (i === 0) {
        previousPath = '';
      } else {
        previousPath = acc[acc.length - 1].path;
      }

      const q = queryString.parse(search);
      let qs = '';
      if (i === 1) {
        // The function detail part needs a query param for the repoPath
        // Currently the second part of the url is always the function name
        // so we can always add the repoPath query
        qs = `?repoPath=${q.repoPath}`;
      }
      acc.push({ name: p, path: `${previousPath}/${p}${qs}` });
      return acc;
    }, []);
  }
  componentDidMount() {
    this.unlisten = this.props.history.listen(location => {
      let parts = this.pathToParts(location.pathname, location.search);
      this.setState({
        parts,
      });
    });
  }
  componentWillUnmount() {
    this.unlisten();
  }
  render() {
    const { parts } = this.state;

    const links = parts.map((p, i) => {
      const isActive = i === parts.length - 1;
      return (
        <li key={i} className={isActive ? 'active' : undefined}>
          {isActive ? p.name : <NavLink to={p.path}>{p.name}</NavLink>}
        </li>
      );
    });

    let breadcrumbClassName = 'breadcrumb';
    if (parts.length <= 1) {
      // hide the breadcrumb at root
      breadcrumbClassName += ' invisible';
    }

    return <ol className={breadcrumbClassName}>{links}</ol>;
  }
}

export const Breadcrumbs = withRouter(BreadcrumbsWithRouter);
