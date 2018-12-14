import React, { Component } from 'react';
import PropTypes from 'prop-types';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Table, UncontrolledDropdown, DropdownToggle, DropdownMenu, DropdownItem } from 'reactstrap';

import { withRouter } from 'react-router-dom';

import { FunctionTableItem, genOwnerInitials } from '../FunctionTableItem';

import './FunctionTable.css'

function renderBody(fns, user, clickHandler, filter) {
  if (fns.length === 0) {
    return (
      <tr>
        <td>No functions available.</td>
      </tr>
    );
  } else {
    return fns.filter(({ gitOwner }) => filter(gitOwner)).map((fn) => {

      return (
        <FunctionTableItem key={fn.shortSha + fn.name} fn={fn} user={user} onClick={clickHandler} />
      );
    });
  }
}

class FunctionTableWithoutRouter extends Component {
  constructor(props) {
    super(props);

    this.state = {
      filter: localStorage.getItem('filter') || '',
    };

    this.saveFilter = this.saveFilter.bind(this);
    this.handleOwnerClick = this.handleOwnerClick.bind(this);
    this.clearFilter = this.clearFilter.bind(this);
    this.renderOwnersElems = this.renderOwnersElems.bind(this);
  }

  saveFilter(filter) {
    this.setState({ filter });

    localStorage.setItem('filter', filter);
  }

  handleOwnerClick(owner) {
    this.saveFilter(owner);
  }

  clearFilter() {
    this.saveFilter('');
  }

  renderOwnersElems(fns) {
    const { filter } = this.state;

    const owners = fns.reduce((acc, { gitOwner }) => {
      if (acc.includes(gitOwner)) {
        return acc;
      }

      return [...acc, gitOwner];
    }, []);

    const elems = owners.map(owner => (
      <DropdownItem
        tag="a"
        onClick={() => this.handleOwnerClick(owner)}
        key={owner}
        active={filter === owner}
      >
        { `(${genOwnerInitials(owner).toUpperCase()}) ${owner}` }
      </DropdownItem>
    ));

    if (filter !== '') {
      elems.push(<DropdownItem key="divider" divider />);
      elems.push((
        <DropdownItem key="clear" onClick={this.clearFilter}>
          Clear
        </DropdownItem>
      ));
    }

    return elems;
  }

  render() {
    const { isLoading, fns, user, history } = this.props;
    const onRowClick = to => history.push(to);

    let tableProps = {
      responsive: true,
      className: 'function-table',
    };

    if (fns && fns.length > 0) {
      tableProps.hover = true;
    }

    const tbody = isLoading ? (
      <tr>
        <td colSpan="8" class="text-center">
          <FontAwesomeIcon icon="spinner" spin />
        </td>
      </tr>
    ) : (
      renderBody(fns, user, onRowClick, item => this.state.filter === '' || this.state.filter === item)
    );

    return (
      <Table {...tableProps}>
        <thead>
        <tr>
          <th>
            <UncontrolledDropdown setActiveFromChild>
              <DropdownToggle tag="a" caret>
                Owner
              </DropdownToggle>
              <DropdownMenu>
                { this.renderOwnersElems(fns) }
              </DropdownMenu>
            </UncontrolledDropdown>
          </th>
          <th>Name</th>
          <th style={{ width: '42px' }} />
          <th>Repository</th>
          <th>SHA</th>
          <th>Deployed</th>
          <th>Invocations</th>
          <th>Replicas</th>
          <th />
        </tr>
        </thead>
        <tbody id="items">
          { tbody }
        </tbody>
      </Table>
    );
  }
}

const FunctionTable = withRouter(FunctionTableWithoutRouter);

FunctionTable.propTypes = {
  isLoading: PropTypes.bool.isRequired,
  fns: PropTypes.array.isRequired,
  user: PropTypes.string.isRequired,
};

export { FunctionTable };
