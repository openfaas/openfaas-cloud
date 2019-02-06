import React, { Component } from 'react';
import queryString from 'query-string';
import { NavLink, withRouter } from 'react-router-dom';
import { Breadcrumb, BreadcrumbItem } from 'reactstrap'

class BreadcrumbsWithRouter extends Component {
  constructor(props) {
    super(props);

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

    const links = parts.map(({ path, name }, i) => {
      const itemProps = {};
      let itemContent = (
        <NavLink to={path}>
          { name }
        </NavLink>
      );

      if (i === parts.length - 1) {
        itemProps.active = true;

        itemContent = name;
      }

      return (
        <BreadcrumbItem key={`${name}${i}`} {...itemProps}>
          { itemContent }
        </BreadcrumbItem>
      );
    });

    return (
      <Breadcrumb listClassName="background-color-inherit-important m-0">
        { links }
      </Breadcrumb>
    );
  }
}

const Breadcrumbs = withRouter(BreadcrumbsWithRouter);

export { Breadcrumbs };
