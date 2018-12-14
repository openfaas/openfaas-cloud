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

const genOwnerInitials = (owner) => (
  (owner.split(/[-_]+/).length > 1) ?
    owner.split(/[-_]+/)[0].substring(0, 1).concat(owner.split(/[-_]+/)[1].substring(0, 1)) :
      owner.split(/[-_]+/)[0].substring(0, 2)

);

// For detailed explanation on what is the method bellow doing, please check
// https://stackoverflow.com/questions/3426404/create-a-hexadecimal-colour-based-on-a-string-with-javascript
const stringToColour = (str) => {
  const hash = [...str].reduce((acc, curr) => curr.charCodeAt(0) + ((acc << 5) - acc), 0);

  let colour = '#';

  for (let i = 0; i < 3; i++) {
    const value = (hash >> (i * 8)) & 0xFF;

    colour += ('00' + value.toString(16)).substr(-2);
  }

  return colour;
};

const padZero = (str, len) => {
  len = len || 2;

  const zeros = new Array(len).join('0');

  return (zeros + str).slice(-len);
};

function invertColor(hex, blackAndWhite) {
  if (hex.indexOf('#') === 0) {
    hex = hex.slice(1);
  }

  // convert 3-digit hex to 6-digits.
  if (hex.length === 3) {
    hex = hex[0] + hex[0] + hex[1] + hex[1] + hex[2] + hex[2];
  }

  if (hex.length !== 6) {
    throw new Error('Invalid HEX color.');
  }

  const r = parseInt(hex.slice(0, 2), 16);
  const g = parseInt(hex.slice(2, 4), 16);
  const b = parseInt(hex.slice(4, 6), 16);

  if (blackAndWhite) {
    // http://stackoverflow.com/a/3943023/112731
    return (r * 0.299 + g * 0.587 + b * 0.114) > 186
      ? '#000000'
      : '#ffffff';
  }

  // invert color components
  const rs = (255 - r).toString(16);
  const gs = (255 - g).toString(16);
  const bs = (255 - b).toString(16);

  // pad each with zeros and return
  return "#" + padZero(rs) + padZero(gs) + padZero(bs);
}

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
  const owner = genOwnerInitials(fn.gitOwner);

  const handleClick = () => onClick(fnDetailPath);

  const ownerColor = stringToColour(fn.gitOwner);

  return (
    <tr onClick={handleClick} className="cursor-pointer">
      <td>
        <div
          className="rounded w-50 text-center text-uppercase"
          title={fn.gitOwner}
          style={{
            backgroundColor: ownerColor,
            color: invertColor(ownerColor, true),
            fontWeight: 700,
          }}
        >
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
  FunctionTableItem,
  genOwnerInitials,
};
