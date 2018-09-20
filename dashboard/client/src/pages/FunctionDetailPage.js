import React, { Component } from 'react';
import queryString from 'query-string';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Link } from 'react-router-dom';
import { functionsApi } from '../api/functionsApi';
import { FunctionDetailSummary } from '../components/FunctionDetailSummary';
import { Modal, Button, FormGroup, ControlLabel, FormControl } from 'react-bootstrap';

export class FunctionDetailPage extends Component {
  constructor(props) {
    super(props);
    const { repoPath } = queryString.parse(props.location.search);
    const { user, functionName } = props.match.params;
    this.handleShowBadgeModal = this.handleShowBadgeModal.bind(this);
    this.handleCloseBadgeModal = this.handleCloseBadgeModal.bind(this);
    this.state = {
      isLoading: true,
      fn: null,
      user,
      repoPath,
      functionName,
      showBadgeModal: false,
    };    
  }
  componentDidMount() {
    const { user, repoPath, functionName } = this.state;
    this.setState({ isLoading: true });
    functionsApi.fetchFunction(user, repoPath, functionName).then(res => {
      this.setState({ isLoading: false, fn: res });
    });
  }
  handleCloseBadgeModal() {
    this.setState({ showBadgeModal: false });
  }

  handleShowBadgeModal() {
    this.setState({ showBadgeModal: true });
  }
  render() {
    const { isLoading, fn } = this.state;
    const badgeImage = "https://img.shields.io/badge/openfaas-cloud-blue.svg"
    const badgeLink = "https://www.openfaas.com"
    const badgeMd = `[![OpenFaaS](${badgeImage})](${badgeLink})`
    let panelBody;
    
    if (isLoading) {
      panelBody = (
        <div className="panel-body">
          <div style={{ textAlign: 'center' }}>
            <FontAwesomeIcon icon="spinner" spin />{' '}
          </div>
        </div>
      );
    } else {
      panelBody = (
        <div className="panel-body">
          <FunctionDetailSummary fn={fn} handleShowBadgeModal={this.handleShowBadgeModal}/>
        </div>
      );
    }
    return (
      <div className="panel panel-success">
        <div className="panel-heading">Function Overview</div>
        {panelBody}
          <Modal show={this.state.showBadgeModal} onHide={this.handleCloseBadgeModal}>
            <Modal.Header closeButton>
              <Modal.Title>Get Badge</Modal.Title>
            </Modal.Header>
            <Modal.Body>
              <p>
                <Link to={badgeLink}>
                  <img src={badgeImage} alt="OpenFaaS" />
                </Link>
              </p>
              <form>
                <FormGroup
                  controlId="formGetBadge"
                >
                  <FormControl
                    type="text"
                    value={badgeMd}
                  />
                  <FormControl.Feedback />
                </FormGroup>
              </form>
            </Modal.Body>
          </Modal>
      </div>
    );
  }
}
