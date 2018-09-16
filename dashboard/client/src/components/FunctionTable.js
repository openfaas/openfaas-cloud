import React from 'react';
import PropTypes from 'prop-types';
import './FunctionTable.css';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

import { Link, withRouter } from 'react-router-dom';

function renderBody(fns, user, clickHandler) {
  if (fns.length === 0) {
    return (
      <tr>
        <td>No functions available.</td>
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
      const fnDetailPath = `${user}/${shortName}?repoPath=${gitOwner}/${gitRepo}`;

      const repoUrl = `https://github.com/${gitOwner}/${gitRepo}/commits/master`;

      const handleRowClick = () => clickHandler(fnDetailPath);

      // FIXME: This needs to use the `replicas` and `maxReplicas`. The code below is just mocking data.
      const percentage = Math.floor(((i + 1) / fns.length) * 100);
      let progressClassName = 'progress-bar';
      if (percentage < 66) {
        progressClassName += ' progress-bar-success';
      } else if (66 <= percentage && percentage < 90) {
        progressClassName += ' progress-bar-warning';
      } else {
        progressClassName += ' progress-bar-danger';
      }
      return (
        <tr key={i} onClick={handleRowClick}>
          <td>
            <a
              className="btn btn-default btn-xs"
              href={endpoint}
              onClick={e => e.stopPropagation()}
            >
              <FontAwesomeIcon icon="link" />
            </a>
          </td>
          <td>{shortName}</td>
          <td>
            <a href={repoUrl} onClick={e => e.stopPropagation()}>
              {gitRepo}
            </a>
          </td>
          <td>{shortSha}</td>
          <td>{sinceDuration}</td>
          <td>{invocationCount}</td>
          <td>
            <div className="progress">
              <div
                className={progressClassName}
                role="progressbar"
                aria-valuenow={percentage}
                aria-valuemin="0"
                aria-valuemax="100"
                style={{ width: `${percentage}%` }}
              />
            </div>
            <div className="text-center">
              <small>
                {i + 1}/{fns.length}
              </small>
            </div>
          </td>
          <td>
            <Link
              className="btn btn-default btn-xs"
              to={logPath}
              onClick={e => e.stopPropagation()}
            >
              <FontAwesomeIcon icon="folder-open" />
            </Link>
          </td>
        </tr>
      );
    });
  }
}

export const FunctionTable = withRouter(({ isLoading, fns, user, history }) => {
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

  let tableClassName = 'table';
  if (fns && fns.length > 0) {
    tableClassName += ' table-hover';
  }
  return (
    <div className="function-table table-responsive">
      <table className={tableClassName}>
        <thead>
          <tr>
            <th style={{ width: '42px' }} />
            <th>Name</th>
            <th>Repository</th>
            <th>SHA</th>
            <th>Deployed</th>
            <th>Invocations</th>
            <th>Replicas</th>
            <th />
          </tr>
        </thead>
        <tbody id="items">{tbody}</tbody>
      </table>
    </div>
  );
});

FunctionTable.propTypes = {
  isLoading: PropTypes.bool.isRequired,
  fns: PropTypes.array.isRequired,
  user: PropTypes.string.isRequired,
};
