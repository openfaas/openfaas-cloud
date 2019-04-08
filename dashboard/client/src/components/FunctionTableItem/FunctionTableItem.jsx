import React from 'react';
import { Link } from "react-router-dom";
import { Button } from 'reactstrap';
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faUserSecret } from "@fortawesome/free-solid-svg-icons";
import { ReplicasProgress } from "../ReplicasProgress";

const genLogPath = ({ shortName, gitOwner, gitRepo, gitSha }, user) => (
  `${user}/${shortName}/log?repoPath=${gitOwner}/${gitRepo}&commitSHA=${gitSha}`
);

const genFnDetailPath = ({ shortName, gitOwner, gitRepo }, user) => (
  `/${user}/${shortName}?repoPath=${gitOwner}/${gitRepo}`
);

const genRepoUrl = ({ gitRepoURL, gitBranch }) => {
  if(gitBranch === "") {
    return `${gitRepoURL}/commits/master`
  }
  return `${gitRepoURL}/commits/${gitBranch}`
};

const FunctionTableItem = ({ onClick, fn, user }) => {
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
  const logPath = genLogPath(fn, user);
  const fnDetailPath = genFnDetailPath(fn, user);

  const handleClick = () => onClick(fnDetailPath);

  return (
    <tr onClick={handleClick} className="cursor-pointer">
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
  FunctionTableItem
};
