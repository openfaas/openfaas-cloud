'use strict';

const fs = require('fs');
const request = require('request');

const handleLogout = (context) => {
  const now = new Date();
  const year = now.getFullYear();
  const month = now.getMonth();
  const day = now.getDate();
  const expires = new Date(year, month, day);
  const headers = {
    'Set-Cookie': [
      'openfaas_cloud_token=',
      `Expires=${expires.toUTCString()}`,
      `Domain=${process.env.cookie_root_domain}`,
      'Path=/',
    ].join('; ')
  };

  fs.readFile(`${__dirname}/dist/logout.html`, (err, data) => {
    if (err) {
      return context.status(500).fail(err);
    }

    context
      .headers(headers)
      .status(200)
      .succeed(data.toString());
  });
};

module.exports = (event, context) => {
  const { method, path , query} = event;

  parseOrganiations: parseOrganizations;
  decodeCookie: decodeCookie;
  getCookie: getCookie;
  getRequestedEntityFromPath: getRequestedEntityFromPath;
  isResourceInTokenClaims: isResourceInTokenClaims;

  if (method !== 'GET') {
    context.status(400).fail('Bad Request');
    return;
  }

  if (/^\/logout\/?$/.test(path)) {
    return handleLogout(context);
  }
  let cookie = getCookie(event);
  let decodedCookie = decodeCookie(cookie);
  let organizations = parseOrganizations(decodedCookie);

  if (/^\/api\/(list-functions|system-metrics|pipeline-log|function-logs).*/.test(path)) {

    // See if a user is trying to query functions they do not have permissions to view
    if (!isResourceInTokenClaims(path, query, decodedCookie["sub"], organizations)) {
      console.log("the user '" + decodedCookie["sub"] + "' tried to access a resource they are not entitled to")
      context.status(403).succeed('Forbidden');
      return;
    }

    // proxy api requests to the gateway
    const gatewayUrl = process.env.gateway_url.replace(/\/$/, '');
    const proxyPath = path.replace(/^\/api\//, '');
    const url = `${gatewayUrl}/function/${proxyPath}`;
    var reqHeaders = event.headers;
    reqHeaders['host'] = gatewayUrl.replace('http://', '');
    console.log(`proxying request to: ${url}`);
    request(
      {
        url,
        method,
        headers: reqHeaders,
        qs: event.query,
      },
      (err, response, body) => {
        console.log('proxy response code:', response.statusCode);
        if (err) {
          console.log('Proxy request failed', err);
          context.status(500).fail('Proxy Request Failed');
          return;
        }
        context
          .headers(response.headers)
          .status(response.statusCode)
          .succeed(body);
      }
    );
    return;
  }

  let headers = {
    'Content-Type': '',
  };
  if (/.*\.js/.test(path)) {
    headers['Content-Type'] = 'application/javascript';
  } else if (/.*\.css/.test(path)) {
    headers['Content-Type'] = 'text/css';
  } else if (/.*\.ico/.test(path)) {
    headers['Content-Type'] = 'image/x-icon';
  } else if (/.*\.json/.test(path)) {
    headers['Content-Type'] = 'application/json';
  } else if (/.*\.map/.test(path)) {
    headers['Content-Type'] = 'application/octet-stream';
  }

  let contentPath = `${__dirname}${path}`;

  if (!headers['Content-Type']) {
    contentPath = `${__dirname}/dist/index.html`;
  }

  fs.readFile(contentPath, (err, data) => {
    if (err) {
      context
        .headers(headers)
        .status(500)
        .fail(err);

      return;
    }

    let content = data.toString();

    function get_all_claims(organizations, decodedCookie) {
      if (decodedCookie && organizations.length > 0) {
        return  organizations + "," + decodedCookie["sub"];
      }
      return ""
    }

    if (!headers['Content-Type']) {
      headers['Content-Type'] = 'text/html';

      const isSignedIn = /openfaas_cloud_token=.*\s*/.test(event.headers.cookie);

      console.log(path);
      if (path === "/" && isSignedIn) {
        headers["Location"] =   "/dashboard/"+ decodedCookie["sub"];
        return context
            .headers(headers)
            .status(307)
            .succeed();
      }

      let claims = get_all_claims(organizations, decodedCookie);

      const { base_href, public_url, pretty_url, query_pretty_url } = process.env;
      content = content.replace(/__BASE_HREF__/g, base_href);
      content = content.replace(/__PUBLIC_URL__/g, public_url);
      content = content.replace(/__PRETTY_URL__/g, pretty_url);
      content = content.replace(/__QUERY_PRETTY_URL__/g, query_pretty_url);
      content = content.replace(/__IS_SIGNED_IN__/g, isSignedIn);
      content = content.replace(/__ALL_CLAIMS__/g, claims);

    }

    context
      .headers(headers)
      .status(200)
      .succeed(content);
  });
};

var parseOrganizations = function (decodedCookie) {
  if (decodedCookie && 'organizations' in decodedCookie) {
    return decodedCookie.organizations;
  }
  return '';
}

var base64Decode = function base64Decode(str) {
  return Buffer.from(str, 'base64').toString('binary');
}

var decodeCookie = function (token) {
  try {
      return JSON.parse(base64Decode(token.split('.')[1]));
  } catch (e) {
    return null;
  }
}

var getCookie = function (event = {}) {
  if (!event.headers && !event.headers.cookie) {
    console.log("event does not contain a cookie");
    return null;
  }
  return event.headers.cookie;
}

var getRequestedEntityFromPath = function (path) {
  var params = new URLSearchParams(path)

  return params.get('user')
}

var isResourceInTokenClaims = function (path, queryString, user, organisations) {
  // check if the user is trying to access the fn logs or fn stats for one of their functions
  if(/^\/api\/(system-metrics|function-logs).*/.test(path)) {
    if (!isRepoOwnedByUser(queryString, user, organisations)) {
      return false
    }
  }

  if (user === queryString["user"]) {
    return true
  }
  return organisations.indexOf(queryString["user"]) >= 0
}

function isRepoOwnedByUser(query, user, organisations) {
  let functionName = query["function"];

  if (functionName.startsWith(user.toLowerCase())) {
    return true
  }

  for (let orgName in organisations) {
    if (functionName.startsWith(organisations[orgName].toLowerCase())) {
      return true
    }
  }
  return false;
}