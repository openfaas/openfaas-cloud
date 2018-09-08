import React, { Component } from 'react';
import queryString from 'query-string';
import AceEditor from 'react-ace';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

import 'brace/mode/sh';
import 'brace/theme/monokai';

import { functionsApi } from '../api/functionsApi';
export class FunctionLogPage extends Component {
  constructor(props) {
    super(props);
    const { commitSHA, repoPath } = queryString.parse(props.location.search);
    const { functionName } = props.match.params;

    this.state = {
      isLoading: true,
      log: '',
      commitSHA,
      repoPath,
      functionName,
    };
  }
  componentDidMount() {
    const { commitSHA, repoPath, functionName } = this.state;
    this.setState({ isLoading: true });
    functionsApi.fetchFunctionLog({ commitSHA, repoPath, functionName }).then(res => {
      this.setState({ isLoading: false, log: res });
    });
  }

  render() {
    const { commitSHA, repoPath, functionName } = this.state;

    const editorOptions = {
      width: '100%',
      height: '600px',
      mode: 'sh',
      theme: 'monokai',
      readOnly: true,
      wrapEnabled: true,
      showPrintMargin: false,
      editorProps: {
        $blockScrolling: true,
      },
    };

    const panelBody = this.state.isLoading ? (
      <div style={{ textAlign: 'center' }}>
        <FontAwesomeIcon icon="spinner" spin />{' '}
      </div>
    ) : (
      <AceEditor {...editorOptions} value={this.state.log} />
    );
    return (
      <div className="panel panel-success">
        <div className="panel-heading">
          Logs for {functionName} in {repoPath} @ {commitSHA}
        </div>
        <div className="panel-body">{panelBody}</div>
      </div>
    );
  }
}
