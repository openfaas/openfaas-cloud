import React from 'react';
import { Link } from 'react-router-dom';
import PropTypes from 'prop-types';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faExternalLinkAlt } from '@fortawesome/free-solid-svg-icons'

import CopyText from './CopyText';

import './FunctionDetailSummary.css';

const renderContainerImage = image => {
  const imageWithoutTag = image.split(':')[0];
  const parts = imageWithoutTag.split('/');

  if (parts.length === 2) {
    return (
      <a
        href={`https://hub.docker.com/r/${parts[0]}/${parts[1]}/tags`}
        target="_blank"
      >
        {image}
      </a>
    );
  } else {
    return image;
  }
};

export const FunctionDetailSummary = ({ fn }) => {
  const to = `${fn.shortName}/log?repoPath=${fn.gitOwner}/${
    fn.gitRepo
  }&commitSHA=${fn.gitSha}`;
  const repo = `${fn.gitOwner}/${fn.gitRepo}`;

  return (
    <div className="fn-detail-summary row">
      <div className="col-md-5">
        <div className="panel panel-default fn-detail-deployment">
          <div className="panel-body">
            <div>
              <div className="pull-right">
                <Link className="btn btn-default" to={to}>
                  <FontAwesomeIcon icon="folder-open" /> Build Logs
                </Link>
              </div>
              <h4>
                Deployment <FontAwesomeIcon icon="info-circle" />
              </h4>
            </div>
            <dl className="FunctionDetailsSummary__dl dl-horizontal">
              <dt className="FunctionDetailsSummary__dl__dt">Name:</dt>
              <dd>{fn.shortName}</dd>
              <dt className="FunctionDetailsSummary__dl__dt">Image:</dt>
              <dd>{renderContainerImage(fn.image)}</dd>
              <dt className="FunctionDetailsSummary__dl__dt">
                  Endpoint:
              </dt>
              <dd>
                <CopyText copyValue={fn.endpoint}>
                    <a className="pointer" style={{ marginRight: 5 }}>
                        {fn.endpoint}
                    </a>
                </CopyText>
                <a href={fn.endpoint} target="__blank">
                  <FontAwesomeIcon icon={faExternalLinkAlt} size="xs" />
                </a>
              </dd>
              <dt className="FunctionDetailsSummary__dl__dt">Replicas:</dt>
              <dd>{fn.replicas}</dd>
            </dl>
          </div>
        </div>
      </div>
      <div className="col-md-5">
        <div className="panel panel-default fn-detail-git">
          <div className="panel-body">
            <div>
              <h4>
                Git <FontAwesomeIcon icon="code-branch" />
              </h4>
            </div>
            <dl className="dl-horizontal">
              <dt>Repository:</dt>
              <dd>
                <a href={`https://github.com/${repo}`} target="_blank">
                  {repo}
                </a>
              </dd>
              <dt>SHA:</dt>
              <dd>
                <a
                  href={`https://github.com/${repo}/commit/${fn.gitSha}`}
                  target="_blank"
                >{`${fn.gitSha}`}</a>
              </dd>
              <dt>Deploy Time:</dt>
              <dd>{`${fn.sinceDuration}`}</dd>
            </dl>
          </div>
        </div>
      </div>
      <div className="col-sm-4 col-md-2">
        <div className="panel panel-default invocation-count">
          <div className="panel-body">
            <div>
              <h4>
                Invocations <span className="visible-lg-inline">&nbsp;<FontAwesomeIcon icon="bolt" /></span>
              </h4>
            </div>
            <div>
              <p>{fn.invocationCount}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

FunctionDetailSummary.propTypes = {
  fn: PropTypes.object.isRequired,
};
