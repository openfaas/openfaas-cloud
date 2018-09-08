import React from 'react';
import PropTypes from 'prop-types';
import './FunctionTable.css';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

import { Link } from 'react-router-dom';

function renderBody(fns, user) {
  if (fns.length === 0) {
    return (
      <tr>
        <td>There are no functions.</td>
      </tr>
    );
  } else {
    return fns.map((fn, i) => {
      const {
        shortName,
        gitRepo,
        shortSha,
        gitSha,
        gitOwner,
        endpoint,
        sinceDuration,
        invocationCount,
        replicas,
      } = fn;

      const logPath = `${user}/${shortName}/log?repoPath=${gitOwner}/${gitRepo}&commitSHA=${gitSha}`;

      const repoUrl = `https://github.com/${gitOwner}/${gitRepo}/commits/master`;
      return (
        <tr key={i}>
          <td>{shortName}</td>
          <td>
            <a href={repoUrl}>{gitRepo}</a>
          </td>
          <td>
            <a href={endpoint}>
              <FontAwesomeIcon icon="link" />
            </a>
          </td>
          <td>
            <Link to={logPath}>
              <FontAwesomeIcon icon="folder-open" />
            </Link>
          </td>
          <td>{shortSha}</td>
          <td>{sinceDuration}</td>
          <td>{invocationCount}</td>
          <td>{replicas}</td>
        </tr>
      );
    });
  }
}

export const FunctionTable = ({ isLoading, fns, user }) => {
  const tbody = isLoading ? (
    <tr>
      <td colSpan="8" style={{ textAlign: 'center' }}>
        <FontAwesomeIcon icon="spinner" spin />
      </td>
    </tr>
  ) : (
    renderBody(fns, user)
  );
  return (
    <div className="table-responsive-sm table-responsive-md">
      <table className="table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Repo</th>
            <th>Endpoint</th>
            <th>Logs</th>
            <th>SHA</th>
            <th>Built</th>
            <th>Invocation Count</th>
            <th>Replicas</th>
          </tr>
        </thead>
        <tbody id="items">{tbody}</tbody>
      </table>
    </div>
  );
};

FunctionTable.propTypes = {
  isLoading: PropTypes.bool.isRequired,
  fns: PropTypes.array.isRequired,
};
