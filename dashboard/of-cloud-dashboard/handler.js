'use strict';

const axios = require('axios');
const fs = require('fs');
const fsPromises = fs.promises
var qs = require('qs');

module.exports = async (event, context) => {
  const { method, path , query} = event;

  if (method !== 'GET') {
    return context.status(405).fail('Method not allowed');
  }

  if (/^\/logout\/?$/.test(path)) {
    return handleLogout(context);
  }

  let cookie = getCookie(event);
  let decodedCookie = decodeCookie(cookie);
  let organizations = parseOrganizations(decodedCookie);

  if (/^\/api\/(list-functions|metrics|pipeline-log|function-logs).*/.test(path)) {

    // See if a user is trying to query functions they do not have permissions to view
    if (!isResourceInTokenClaims(path, query, decodedCookie, organizations)) {
      console.log("The user '" + decodedCookie["sub"] + "' tried to access a resource they are not entitled to")
      return context.status(403).succeed('Forbidden');
    }

    // proxy api requests to the gateway
    const gatewayUrl = process.env.gateway_url.replace(/\/$/, '');
    const proxyPath = path.replace(/^\/api\//, '');
    const url = `${gatewayUrl}/function/${proxyPath}`;
    var reqHeaders = event.headers;
    reqHeaders['host'] = gatewayUrl.replace('http://', '');

    let upstreamURL = url;
    if(Object.keys(query).length > 0) {
      upstreamURL = upstreamURL+"?"+qs.stringify(query);
    }

    try {
      let opts = {
        url: upstreamURL,
        method: method,
        headers: reqHeaders,
      };
  
      let res = await axios(opts)
      let ctx = context;

      console.log(`${method} ${upstreamURL} - ${res.status}`);

      return ctx.status(res.status)
                .headers(res.headers)
                .succeed(res.data);

    } catch(err) {
      console.log(`${method} ${upstreamURL} - 500, error: ${err}`);
      return context.status(500).fail('Proxy request failed');
    }
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

  let fileData = ""
  try {
    fileData = await fsPromises.readFile(contentPath);
  } catch (err) {
    return context
    .headers(headers)
    .status(500)
    .fail(err);
  }

  let content = fileData.toString();

  if (!headers['Content-Type']) {
    headers['Content-Type'] = 'text/html';

    const isSignedIn = /openfaas_cloud_token=.*\s*/.test(event.headers.cookie);

    if (path === "/" && isSignedIn) {
      let statusCode = 404

      // If we have a cookie, and it has a subject, then redirect to the subject's dashboard
      if (decodedCookie && decodedCookie["sub"]) {
        headers["Location"] =   "/dashboard/"+ decodedCookie["sub"];
        statusCode = 307
      }

      return context
          .headers(headers)
          .status(statusCode)
          .succeed()
      
    }

    let claims = get_all_claims(organizations, decodedCookie);

    content = replaceTokens(content, isSignedIn, claims)
  }

  context
    .headers(headers)
    .status(200)
    .succeed(content);
}

function replaceTokens(content, isSignedIn, claims) {
    const { base_href, public_url, pretty_url, query_pretty_url } = process.env;
    let replaced = content

    replaced = replaced.replace(/__BASE_HREF__/g, base_href);
    replaced = replaced.replace(/__PUBLIC_URL__/g, public_url);
    replaced = replaced.replace(/__PRETTY_URL__/g, pretty_url);
    replaced = replaced.replace(/__QUERY_PRETTY_URL__/g, query_pretty_url);
    replaced = replaced.replace(/__IS_SIGNED_IN__/g, isSignedIn);
    replaced = replaced.replace(/__ALL_CLAIMS__/g, claims);

    return replaced
}

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
    console.log("Event does not contain a cookie");
    return null;
  }
  return event.headers.cookie;
}

var getRequestedEntityFromPath = function (path) {
  var params = new URLSearchParams(path)

  return params.get('user')
}

var isResourceInTokenClaims = function (path, queryString, decodedCookie, organisations) {
  if (!decodedCookie) {
    // We are running without auth if we get to this point without a cookie - the edge-router validates auth
    return true
  }
  let user = decodedCookie["sub"]
  // check if the user is trying to access the fn logs or fn stats for one of their functions
  if(/^\/api\/(metrics|function-logs).*/.test(path)) {
    if (!isRepoOwnedByUser(queryString, user, organisations)) {
      return false
    }
  }

  if (user === queryString["user"]) {
    return true
  }

  let orgs = organisations.split(",");
  return orgs.indexOf(queryString["user"]) >= 0
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

const handleLogout = async (context) => {
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

  let data = "";
  try {
    data = await fsPromises.readFile(`${__dirname}/dist/logout.html`)
  } catch (err) {
    return context.status(500).fail(err);
  }

  return context
    .headers(headers)
    .status(200)
    .succeed(data.toString());
}


function get_all_claims(organizations, decodedCookie) {
  if (decodedCookie && organizations.length > 0) {
    return  organizations.length > 0 ? organizations + "," + decodedCookie["sub"] : decodedCookie["sub"];
  }
  return ""
}