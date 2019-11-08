import React, { Component } from 'react';
import queryString from 'query-string';
import AceEditor from 'react-ace';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Card, CardHeader, CardBody } from 'reactstrap';

import 'brace/mode/sh';
import 'brace/theme/monokai';

import { functionsApi } from '../api/functionsApi';

const onEditorLoad = (editor) => {
  editor.scrollToLine(editor.getSession().getLength());
  editor.navigateLineEnd();
};

export class BuildLogPage extends Component {
  constructor(props) {
    super(props);
    const { commitSHA, repoPath } = queryString.parse(props.location.search);
    const { functionName, user} = props.match.params;

    this.state = {
      isLoading: true,
      log: '',
      commitSHA,
      repoPath,
      functionName,
      user
    };
  }

  componentDidMount() {
    const { commitSHA, repoPath, functionName, user } = this.state;

    this.setState({ isLoading: true });

    functionsApi.fetchBuildLog({ commitSHA, repoPath, functionName, user }).then(res => {
      this.setState({ isLoading: false, log: res });
    });
  }

  render() {
    const {
      commitSHA,
      repoPath,
      functionName,
      log,
      isLoading,
    } = this.state;

    const editorOptions = {
      width: '100%',
      height: '600px',
      mode: 'sh',
      theme: 'monokai',
      readOnly: true,
      onLoad: onEditorLoad,
      wrapEnabled: true,
      showPrintMargin: false,
      editorProps: {
        $blockScrolling: true,
      },
    };

    let panelBody = <AceEditor {...editorOptions} value={log} />;

    if (isLoading === true) {
      panelBody = (
        <div style={{ textAlign: 'center' }}>
          <FontAwesomeIcon icon="spinner" spin />{' '}
        </div>
      );
    }

    return (
      <Card outline color="success">
        <CardHeader className="bg-success color-success">
          Build logs from {functionName} @ {commitSHA} - ({repoPath})
        </CardHeader>
        <CardBody>
          { panelBody }
        </CardBody>
      </Card>
    );
  }
}
