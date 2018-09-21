import React, { Component } from 'react';
import queryString from 'query-string';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Modal, Form, Well } from 'react-bootstrap';
import { CopyToClipboard } from 'react-copy-to-clipboard';
import { functionsApi } from '../api/functionsApi';
import { FunctionDetailSummary } from '../components/FunctionDetailSummary';


export class FunctionDetailPage extends Component {
  constructor(props) {
    super(props);
    const { repoPath } = queryString.parse(props.location.search);
    const { user, functionName } = props.match.params;
    this.handleShowBadgeModal = this.handleShowBadgeModal.bind(this);
    this.handleCloseBadgeModal = this.handleCloseBadgeModal.bind(this);
    this.hoverOn = this.hoverOn.bind(this);
    this.hoverOff = this.hoverOff.bind(this);
    this.state = {
      isLoading: true,
      fn: null,
      user,
      repoPath,
      functionName,
      showBadgeModal: false,
      copiedIcon: false,
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
    this.setState({ showBadgeModal: false, copiedIcon: false });
  }
  handleShowBadgeModal() {
    this.setState({ showBadgeModal: true });
  }
  hoverOn() {
    document.body.style.cursor = "pointer";
  }
  hoverOff() { 
    document.body.style.cursor = "default"; 
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
                Add the following markdown snippet to your README.md in your Git repo to show you are using OpenFaaS &reg; Cloud.
              </p>
              <p>
                <a href={badgeLink}>
                  <img src={badgeImage} alt="Powered by OpenFaaS &reg; Cloud" />
                </a>
              </p>
              <Form>                
                <CopyToClipboard text={badgeMd}
                  onCopy={() => this.setState({copiedIcon: true})}>                  
                  <Well 
                    bsSize="small" 
                    onMouseEnter={this.hoverOn} 
                    onMouseLeave={this.hoverOff}
                    style={{ marginBottom: "10px" }}
                  >
                    {badgeMd}
                  </Well>                  
                </CopyToClipboard>
                {this.state.copiedIcon ? <span style={{color: 'green'}}>Copied!</span> : null}
              </Form>              
            </Modal.Body>            
          </Modal>
      </div>
    );
  }
}
