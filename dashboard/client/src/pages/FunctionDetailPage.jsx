import React, { Component } from 'react';
import queryString from 'query-string';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Card, CardHeader, CardBody } from 'reactstrap';

import { functionsApi } from '../api/functionsApi';
import { FunctionDetailSummary } from '../components/FunctionDetailSummary';
import { GetBadgeModal } from '../components/GetBadgeModal';
import { ModalRunOnMyOF } from '../components/ModalRunOnMyOF';

export class FunctionDetailPage extends Component {
  constructor(props) {
    super(props);
    const { repoPath } = queryString.parse(props.location.search);
    const { user, functionName } = props.match.params;

    this.handleShowBadgeModal = this.handleShowBadgeModal.bind(this);
    this.handleCloseBadgeModal = this.handleCloseBadgeModal.bind(this);

    this.handleShowRunOnMyOFModal = this.handleShowRunOnMyOFModal.bind(this);
    this.handleCloseRunOnMyOFModal = this.handleCloseRunOnMyOFModal.bind(this);

    this.state = {
      isLoading: true,
      fn: null,
      functionInvocationData: null,
      user,
      repoPath,
      functionName,
      showBadgeModal: false,
      showRunOnMyOFModal: false
    };
  }

  changeFunctionInvocationTimePeriod = timePeriod => {
    const { user, functionName } = this.state;

    functionsApi
      .fetchFunctionInvocation({
        user,
        functionName,
        timePeriod
      })
      .then(res => {
        this.setState({ functionInvocationData: res });
      });
  };

  componentDidMount() {
    const { user, repoPath, functionName } = this.state;

    this.setState({ isLoading: true });

    functionsApi.fetchFunction(user, repoPath, functionName).then(res => {
      this.setState({ isLoading: false, fn: res });
    });

    this.changeFunctionInvocationTimePeriod('60m');
  }

  handleShowBadgeModal() {
    this.setState({ showBadgeModal: true });
  }

  handleCloseBadgeModal() {
    this.setState({ showBadgeModal: false });
  }

  handleShowRunOnMyOFModal() {
    this.setState({ showRunOnMyOFModal: true });
  }

  handleCloseRunOnMyOFModal() {
    this.setState({ showRunOnMyOFModal: false });
  }

  render() {
    const { isLoading, fn, functionInvocationData } = this.state;
    let panelBody = (
      <FunctionDetailSummary
        fn={fn}
        changeFunctionInvocationTimePeriod={
          this.changeFunctionInvocationTimePeriod
        }
        functionInvocationData={functionInvocationData}
        handleShowBadgeModal={this.handleShowBadgeModal}
        handleShowRunOnMyOFModal={this.handleShowRunOnMyOFModal}
      />
    );

    if (isLoading) {
      panelBody = (
        <div style={{ textAlign: 'center' }}>
          <FontAwesomeIcon icon="spinner" spin />{' '}
        </div>
      );
    }

    return (
      <Card outline color="success">
        <CardHeader className="bg-success color-success d-flex align-items-center justify-content-between">
          <div>Function Overview</div>
        </CardHeader>
        <CardBody>{panelBody}</CardBody>
        <GetBadgeModal
          state={this.state.showBadgeModal}
          closeModal={this.handleCloseBadgeModal}
        />
        <ModalRunOnMyOF
          fn={fn || {}}
          state={this.state.showRunOnMyOFModal}
          closeModal={this.handleCloseRunOnMyOFModal}
        />
      </Card>
    );
  }
}
