import React from 'react';
import { Link } from 'react-router-dom';
import PropTypes from 'prop-types';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faAward, faUserSecret } from '@fortawesome/free-solid-svg-icons';

import { FunctionOverviewPanel } from '../FunctionOverviewPanel'
import { ReplicasProgress } from "../ReplicasProgress";

import { Button } from 'reactstrap';

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

const FunctionDetailSummary = ({ fn, handleShowBadgeModal }) => {
  const to = `${fn.shortName}/log?repoPath=${fn.gitOwner}/${
    fn.gitRepo
  }&commitSHA=${fn.gitSha}`;
  const repo = `${fn.gitOwner}/${fn.gitRepo}`;

  const deployMeta = [
    {
      label: 'Name:',
      value: fn.shortName,
    },
    {
      label: 'Image:',
      value: renderContainerImage(fn.image),
    },
    {
      label: 'Endpoint:',
      renderValue() {
        return (
          <a href={fn.endpoint} target="_blank">
            {fn.endpoint}
          </a>
        )
      },
    },
    {
      label: 'Replicas:',
      renderValue() {
        return (
          <ReplicasProgress fn={fn} className="" />
        )
      },
    }
  ];

  const gitMeta = [
    {
      label: 'Repository:',
      renderValue() {
        return (
          <a href={`https://github.com/${repo}`} target="_blank">
            { repo }
          </a>
        );
      }
    },
    {
      label: 'SHA:',
      renderValue() {
        return (
          <a
            href={`https://github.com/${repo}/commit/${fn.gitSha}`}
            target="_blank"
          >
            { fn.gitSha }
          </a>
        )
      },
    },
    {
      label: 'Deploy Time:',
      value: fn.sinceDuration,
    },
  ];

  const deployIcon = <FontAwesomeIcon icon="info-circle" className="mr-3" />;
  const gitIcon = (
    <span>
      <FontAwesomeIcon icon="code-branch" className="mr-3" />
      { fn.gitPrivate && <FontAwesomeIcon icon={faUserSecret} className="mr-3" /> }
    </span>
  );
  const invocationsIcon = <FontAwesomeIcon icon="bolt" className="mr-3 mr-lg-2 d-inline-block d-lg-none d-xl-inline-block" />;
  const deployButton = (
    <Button outline color="secondary" size="xs" tag={Link} to={to}>
      <FontAwesomeIcon icon="folder-open" className="mr-2" />
      <span>Build Logs</span>
    </Button>
  );
  const gitButton = (
    <Button outline color="secondary" size="xs" onClick={handleShowBadgeModal}>
      <FontAwesomeIcon icon={faAward} className="mr-2" />
      <span>Get Badge</span>
    </Button>
  );

  return (
    <div className="FunctionDetailSummary fn-detail-summary row">
      <div className="col-lg-5 pb-3 pb-lg-0">
        <FunctionOverviewPanel
          headerText="Deployment"
          headerIcon={deployIcon}
          button={deployButton}
        >
          <FunctionOverviewPanel.MetaList list={deployMeta} />
        </FunctionOverviewPanel>
      </div>
      <div className="col-lg-5 pb-3 pb-lg-0">
        <FunctionOverviewPanel
          headerText="Git"
          headerIcon={gitIcon}
          button={gitButton}
        >
          <FunctionOverviewPanel.MetaList list={gitMeta} sizes={{
            xs: 12,
            sm: 3,
            md: 2,
            lg: 4,
            xl: 3,
          }} />
        </FunctionOverviewPanel>
      </div>
      <div className="col-lg-2">
        <FunctionOverviewPanel
          headerText="Invocations"
          headerIcon={invocationsIcon}
          bodyClassName="d-flex justify-content-center align-items-center"
        >
          <h1 className="font-weight-bold">
            { fn.invocationCount }
          </h1>
        </FunctionOverviewPanel>
      </div>
    </div>
  );
};

FunctionDetailSummary.propTypes = {
  fn: PropTypes.object.isRequired,  
  handleShowBadgeModal: PropTypes.func.isRequired,
};

export {
  FunctionDetailSummary
};
