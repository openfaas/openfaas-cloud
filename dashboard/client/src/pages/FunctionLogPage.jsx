import React, { Component } from 'react';
import AceEditor from 'react-ace';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {Card, CardHeader, CardBody, Button} from 'reactstrap';

import 'brace/mode/sh';
import 'brace/theme/monokai';

import { functionsApi } from '../api/functionsApi';
import {faExclamationTriangle, faSync} from "@fortawesome/free-solid-svg-icons";

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
      user,
      fetchError: false
    };
  }

  componentDidMount() {
    const { functionName, user } = this.state;

    this.setState({ isLoading: true });

    const longFnName = `${user.toString().toLowerCase()}-${functionName}`
      try {
          functionsApi.fetchFunctionLog({longFnName, user}).then(res => {
            this.setState({isLoading: false, log: res});

          })
      } catch (e) {
        this.setState({isLoading: false, log: [], fetchError: e})
      }
    };

  getPanelBody(logs, isLoading, fetchError) {
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

        if (isLoading) {
            return  (
                <div style={{ textAlign: 'center' }}>
                    <FontAwesomeIcon icon="spinner" spin />{' '}
                </div>
            );
        }

        let infoMsg = fetchError ? "There was an error fetching the logs" : "No logs found for this function, try invoking the function and re-loading this page";

        if (logs.replace(/\s/g,'').length > 0) {
           return  <AceEditor {...editorOptions} value={logs} />;
        } else return (
           <Card>
               <CardHeader>
                   <FontAwesomeIcon icon={faExclamationTriangle} /> {infoMsg}
               </CardHeader>
               <AceEditor {...editorOptions} value={logs} />
           </Card>
        )
    }

    reloadPage(functionName, user) {
      const longFnName = `${user.toString().toLowerCase()}-${functionName}`

      try {
          functionsApi.fetchFunctionLog({longFnName, user}).then(res => {
              this.setState({log: res});
          })
      } catch (e) {
          this.setState({fetchError: e})
      }
    }

    render() {
        const {
            functionName,
            log,
            isLoading,
            user
        } = this.state;

        return (
            <Card outline color="success">
                <CardHeader className="bg-success color-success">
                    Function logs for {functionName}
                    <Button
                        outline
                        size="xs"
                        title="Re-load logs"
                        className="float-right"
                        onClick={() => this.reloadPage(functionName, user)}
                    >
                        <FontAwesomeIcon icon={faSync} />
                    </Button>
                </CardHeader>
                <CardBody>
                    {this.getPanelBody(log, isLoading)}
                </CardBody>
            </Card>
        );
    }
}

