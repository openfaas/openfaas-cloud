import React, { Component } from 'react';
import AceEditor from 'react-ace';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Card, CardHeader, CardBody } from 'reactstrap';

import 'brace/mode/sh';
import 'brace/theme/monokai';

import { functionsApi } from '../api/functionsApi';
import {faExclamationTriangle} from "@fortawesome/free-solid-svg-icons";

const onEditorLoad = (editor) => {
  editor.scrollToLine(editor.getSession().getLength());
  editor.navigateLineEnd();
};

export class FunctionLogPage extends Component {
  constructor(props) {
    super(props);
    const { functionName, user} = props.match.params;

    this.state = {
      isLoading: true,
      log: '',
      functionName,
      user
    };
  }

  componentDidMount() {
    const { functionName, user } = this.state;

    this.setState({ isLoading: true });

    const longFnName = `${user.toString().toLowerCase()}-${functionName}`

    functionsApi.fetchFunctionLog({ longFnName, user }).then(res => {
        this.setState({ isLoading: false, log: res });
    });
  }

  getPanelBody(logs, isLoading) {
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

        if (isLoading === true) {
            return  (
                <div style={{ textAlign: 'center' }}>
                    <FontAwesomeIcon icon="spinner" spin />{' '}
                </div>
            );
        }

       if (logs.replace(/\s/g,'').length > 0) {
           return  <AceEditor {...editorOptions} value={logs} />;
       } else return (
           <Card>
               <CardHeader>
                   <FontAwesomeIcon icon={faExclamationTriangle} /> No lt rebase qqogs found for this function, try invoking the function and re-loading this page
               </CardHeader>
               <AceEditor {...editorOptions} value={logs} />;
           </Card>
       )
    }

    render() {
        const {
            functionName,
            log,
            isLoading,
        } = this.state;

        return (
            <Card outline color="success">
                <CardHeader className="bg-success color-success">
                    Function logs for {functionName}
                </CardHeader>
                <CardBody>
                    {this.getPanelBody(log, isLoading)}
                </CardBody>
            </Card>
        );
    }
}

