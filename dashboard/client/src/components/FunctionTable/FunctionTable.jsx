import React from 'react';
import PropTypes from 'prop-types';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Table } from 'reactstrap';

import { withRouter } from 'react-router-dom';

import { FunctionTableItem } from '../FunctionTableItem';

import './FunctionTable.css'

function renderBody(fns, user, clickHandler) {
  if (fns.length === 0) {
    return (
      <tr>
        <td>No functions available.</td>
      </tr>
    );
  } else {
    return fns.map((fn) => {
      return (
        <FunctionTableItem key={fn.shortSha} fn={fn} user={user} onClick={clickHandler} />
      );
    });
  }
}

const FunctionTable = withRouter(({ isLoading, fns, user, history }) => {
  const onRowClick = to => history.push(to);
  const tbody = isLoading ? (
    <tr>
      <td colSpan="8" style={{ textAlign: 'center' }}>
        <FontAwesomeIcon icon="spinner" spin />
      </td>
    </tr>
  ) : (
    renderBody(fns, user, onRowClick)
  );

  let tableProps = {
    responsive: true,
    className: 'function-table',
  };

  if (fns && fns.length > 0) {
    tableProps.hover = true;
  }

  return (
    <Table {...tableProps}>
      <thead>
      <tr>
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
});

FunctionTable.propTypes = {
  isLoading: PropTypes.bool.isRequired,
  fns: PropTypes.array.isRequired,
  user: PropTypes.string.isRequired,
};

export { FunctionTable };
