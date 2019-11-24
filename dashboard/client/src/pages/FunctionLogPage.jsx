import React, { Component } from 'react';
import AceEditor from 'react-ace';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {Card, CardHeader, CardBody, Button, Input, InputGroup, InputGroupText, Container, Col, Row} from 'reactstrap';

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
      fetchError: false,
      numLines: 100,
      since: 0
    };

    this.changeNumLines = this.changeNumLines.bind(this);
    this.changeSince = this.changeSince.bind(this);
  }

  componentDidMount() {
    const { functionName, user, numLines, since } = this.state;

    this.setState({ isLoading: true });

    const longFnName = `${user.toString().toLowerCase()}-${functionName}`
      try {
          functionsApi.fetchFunctionLog({longFnName, user, numLines, since }).then(res => {
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

    changeNumLines(event) {
        this.setState({numLines: event.target.value}, () => {
            const { functionName, user, numLines, since } = this.state;
            this.reloadPage(functionName, user, numLines, since)
        });
    }

    changeSince(event) {
        this.setState({since: event.target.value}, () => {
            const { functionName, user, numLines, since } = this.state;
            this.reloadPage(functionName, user, numLines, since)
        });
    }

    reloadPage(functionName, user, numLines, since) {
      const longFnName = `${user.toString().toLowerCase()}-${functionName}`

      try {
          functionsApi.fetchFunctionLog({longFnName, user, numLines, since}).then(res => {
              this.setState({log: res, isLoading: false});
          })
      } catch (e) {
          this.setState({fetchError: e, isLoading: false})
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
                    <Container>
                        <Row>
                            <Col xs="12" md="4" lg="6">
                                Function logs for {functionName}
                            </Col>
                            <Col xs="12" md="8" lg="6">
                                <Row>
                                    <Col xs="5" md="4" lg="4">
                                        <InputGroup size="sm">
                                            <Input placeholder="Num of lines" onChange={this.changeNumLines} value={this.state.numLines} min={0} max={500} type="number" step="5" />
                                            <InputGroupText className="form-control">Lines</InputGroupText>
                                        </InputGroup>
                                    </Col>
                                    <Col xs="5" md="4" lg="4">
                                        <InputGroup size="sm">
                                            <Input placeholder="since minutes" onChange={this.changeSince} value={this.state.since} min={0} max={1440} type="number" step="15" />
                                            <InputGroupText className="form-control">Mins</InputGroupText>
                                        </InputGroup>
                                    </Col>
                                    <Col xs="1" md="1" lg="1">
                                        <InputGroup size="sm">
                                            <InputGroupText type="button"
                                                title="Re-load logs"
                                                className="float-right form-control reload-width"
                                                onClick={() => this.reloadPage(functionName, user, this.state.numLines, this.state.since)}
                                            >
                                                <FontAwesomeIcon icon={faSync} />
                                            </InputGroupText>
                                        </InputGroup>
                                    </Col>
                                </Row>
                            </Col>
                        </Row>
                    </Container>
                </CardHeader>
                <CardBody>
                    {this.getPanelBody(log, isLoading)}
                </CardBody>
            </Card>
        );
    }
}

