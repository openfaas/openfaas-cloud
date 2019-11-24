import React from 'react';
import { Link } from "react-router-dom";
import { Button } from 'reactstrap';
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faUserSecret } from "@fortawesome/free-solid-svg-icons";
import { ReplicasProgress } from "../ReplicasProgress";

const genLogPath = ({ shortName, gitOwner, gitRepo, gitSha}) => (
  `${gitOwner}/${shortName}/function-log?repoPath=${gitOwner}/${gitRepo}&commitSHA=${gitSha}`
);

const genFnDetailPath = ({ shortName, gitOwner, gitRepo }) => (
  `/${gitOwner}/${shortName}?repoPath=${gitOwner}/${gitRepo}`
);

const genRepoUrl = ({ gitRepoURL, gitBranch }) => {
  if(gitBranch === "") {
    return `${gitRepoURL}/commits/master`
  }
  return `${gitRepoURL}/commits/${gitBranch}`
};

const genOwnerInitials = (owner) => (
  (owner.split(/[-_]+/).length > 1) ?
    owner.split(/[-_]+/)[0].substring(0, 1).concat(owner.split(/[-_]+/)[1].substring(0, 1)) :
      owner.split(/[-_]+/)[0].substring(0, 2)

);

const FunctionTableItem = ({ onClick, fn }) => {
  const {
    shortName,
    gitRepo,
    shortSha,
    gitPrivate,
    endpoint,
    sinceDuration,
    invocationCount,
  } = fn;

  const repoUrl = genRepoUrl(fn);
  const logPath = genLogPath(fn);
  const fnDetailPath = genFnDetailPath(fn);

  const handleClick = () => onClick(fnDetailPath);

  return (
    <tr onClick={handleClick} className="cursor-pointer">
      <td>
          { fn.gitOwner}
      </td>
      <td>{shortName}</td>
      <td>
        <Button
          outline
          size="xs"
          href={endpoint}
          onClick={e => e.stopPropagation()}
        >
          <FontAwesomeIcon icon="link" />
        </Button>
      </td>
      <td>
        <div className="d-flex justify-content-between align-items-center">
          <a href={repoUrl} onClick={e => e.stopPropagation()}>
            { gitRepo }
          </a>
          { gitPrivate && <FontAwesomeIcon icon={faUserSecret} /> }
        </div>
      </td>
      <td>
        { shortSha }
      </td>
      <td>
        { sinceDuration }
      </td>
      <td>
        { invocationCount }
      </td>
      <td>
        <ReplicasProgress fn={fn} />
      </td>
      <td>
        <Button
          outline
          size="xs"
          to={logPath}
          onClick={e => e.stopPropagation()}
          tag={Link}
        >
          <FontAwesomeIcon icon="folder-open" />
        </Button>
      </td>
    </tr>
  )
};

export {
  FunctionTableItem,
  genOwnerInitials,
};
