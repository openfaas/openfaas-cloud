import axios from 'axios';
import moment from 'moment';

class FunctionsApi {
  constructor() {
    this.selectedRepo = '';
    this.prettyDomain = window.PRETTY_URL;
    this.queryPrettyUrl = window.QUERY_PRETTY_URL === 'true';

    if (process.env.NODE_ENV === 'production') {
      this.baseURL = window.PUBLIC_URL;
      this.apiBaseUrl = `${window.BASE_HREF}api`;
    } else {
      this.baseURL = 'http://localhost:8080';
      this.apiBaseUrl = '/api';
    }

    this.cachedFunctions = {};
  }

  parseFunctionResponse({ data }, user) {
    data.sort((a, b) => {
      if (
        !a ||
        !b ||
        (!a.labels['Git-DeployTime'] || !b.labels['Git-DeployTime'])
      ) {
        return 0;
      }
      return (
        parseInt(b.labels['Git-DeployTime'], 10) -
        parseInt(a.labels['Git-DeployTime'], 10)
      );
    });

    const userPrefixRegex = new RegExp(`^${user}-`);

    return data.map(item => {
      const since = new Date(
        parseInt(item.labels['Git-DeployTime'], 10) * 1000
      );
      const sinceDuration = moment(since).fromNow();

      const shortName = item.name.replace(userPrefixRegex, '');

      let endpoint;

      if (this.queryPrettyUrl) {
        endpoint = this.prettyDomain
          .replace('user', user)
          .replace('function', shortName);
      } else {
        endpoint = this.baseURL + '/function/' + item.name;
      }

      let shortSha = item.labels['Git-SHA'];
      if (shortSha) {
        shortSha = shortSha.substr(0, 7);
      } else {
        shortSha = 'unknown';
      }

      return {
        name: item.name,
        image: item.image,
        shortName,
        endpoint,
        shortSha,
        sinceDuration,
        invocationCount: item.invocationCount,
        replicas: item.replicas,
        gitRepo: item.labels['Git-Repo'],
        gitOwner: item.labels['Git-Owner'],
        gitDeployTime: item.labels['Git-DeployTime'],
        gitSha: item.labels['Git-SHA'],
      };
    });
  }
  fetchFunctions(user) {
    const url = `${this.apiBaseUrl}/list-functions?user=${user}`;
    return axios
      .get(url)
      .then(res => this.parseFunctionResponse(res, user))
      .then(data => {
        this.cachedFunctions = data.reduce((cache, fn) => {
          cache[`${user}/${fn.gitOwner}/${fn.gitRepo}/${fn.shortName}`] = fn;
          return cache;
        }, {});
        return data;
      });
  }

  fetchFunction(user, gitRepo, fnShortname) {
    return new Promise((resolve, reject) => {
      const key = `${user}/${gitRepo}/${fnShortname}`;

      const cachedFn = this.cachedFunctions[key];
      if (cachedFn) {
        resolve(cachedFn);
        return;
      }

      // fetch functions if cache not found
      this.fetchFunctions(user).then(() => {
        const fn = this.cachedFunctions[key];
        fn !== undefined
          ? resolve(fn)
          : reject(new Error(`Function ${key} not found`));
      });
    });
  }

  fetchFunctionLog({ commitSHA, repoPath, functionName }) {
    const url = `${
      this.apiBaseUrl
    }/pipeline-log?commitSHA=${commitSHA}&repoPath=${repoPath}&function=${functionName}`;
    return axios.get(url).then(res => {
      return res.data;
    });
  }
}

export const functionsApi = new FunctionsApi();
