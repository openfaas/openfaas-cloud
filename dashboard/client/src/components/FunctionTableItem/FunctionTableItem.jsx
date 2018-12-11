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

const genRepoUrl = ({ gitOwner, gitRepoURL }) => (
  `${gitRepoURL}/commits/master`
);

const genOwner = ({ gitOwner }) => (
  (gitOwner.split(/[-_]+/).length > 1) ?
    gitOwner.split(/[-_]+/)[0].substring(0, 1).concat(gitOwner.split(/[-_]+/)[1].substring(0, 1)) :
      gitOwner.split(/[-_]+/)[0].substring(0, 2)

);

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
  const owner = genOwner(fn);

  const handleClick = () => onClick(fnDetailPath);

  return (
    <tr onClick={handleClick} className="cursor-pointer">
      <td>
        <div class="rounded border border-secondary w-50 text-center">
          { owner }
        </div>
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
  FunctionTableItem
};
