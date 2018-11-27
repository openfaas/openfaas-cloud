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
  const { method, path } = event;

  parseOrganizations: parseOrganizations;
  decodeCookie: decodeCookie;
  getCookie: getCookie;

  if (method !== 'GET') {
    context.status(400).fail('Bad Request');
    return;
  }

  if (/^\/logout\/?$/.test(path)) {
    return handleLogout(context);
  }

  if (/^\/api\/(list-functions|system-metrics|pipeline-log).*/.test(path)) {
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

    if (!headers['Content-Type']) {
      headers['Content-Type'] = 'text/html';

      const isSignedIn = /openfaas_cloud_token=.*\s*/.test(event.headers.cookie);

      let cookie = getCookie(event);
      let organizations = parseOrganizations(cookie);

      const { base_href, public_url, pretty_url, query_pretty_url } = process.env;
      content = content.replace(/__BASE_HREF__/g, base_href);
      content = content.replace(/__PUBLIC_URL__/g, public_url);
      content = content.replace(/__PRETTY_URL__/g, pretty_url);
      content = content.replace(/__QUERY_PRETTY_URL__/g, query_pretty_url);
      content = content.replace(/__IS_SIGNED_IN__/g, isSignedIn);
      content = content.replace(/__ORGANIZATIONS__/g, organizations);

    }

    context
      .headers(headers)
      .status(200)
      .succeed(content);
  });
};

var parseOrganizations = function (cookie) {
  var decodedCookie = decodeCookie(cookie);
  console.log("decodedCookie -> ", decodedCookie);
  if (decodedCookie && 'organizations' in decodedCookie) {
    return decodedCookie.organizations;
  }
  return '';
}

var atob = function atob(str) {
  return Buffer.from(str, 'base64').toString('binary');
}

var decodeCookie = function (token) {
  try {
    return JSON.parse(atob(token.split('.')[1]));
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
