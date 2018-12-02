import React from 'react';
import {
  Badge,
  ListGroup,
  ListGroupItem,
  Nav,
  NavLink,
  Progress
} from 'reactstrap';

import './FunctionInvocation.css';

const OPTIONS = {
  '1hr': '60m',
  '24hr': '1440m'
};

export class FunctionInvocation extends React.Component {
  state = {
    selected: '1hr'
  };

  render() {
    const { functionInvocationData } = this.props;
    let { success, failure } = functionInvocationData;
    const navLinks = Object.keys(OPTIONS).map(option => {
      return (
        <NavLink
          key={option}
          href="#"
          active={option === this.state.selected}
          onClick={() => this.navLinkClickHandle(option)}
        >
          {option}
        </NavLink>
      );
    });

    const total = success + failure;
    const successPercent = (success / total) * 100;
    const failurePercent = (failure / total) * 100;

    return (
      <div className="">
        <Nav className="d-flex justify-content-center">
          <span className="d-flex align-items-center mr-4 font-weight-bold">
            Period:
          </span>
          {navLinks}
        </Nav>
        <div>
          <Progress multi={true} className="mt-3 d-flex justify-content-center">
            <Progress bar={true} color="success" value={successPercent} />
            <Progress bar={true} color="danger" value={failurePercent} />
          </Progress>
          <span className="font-weight-bold">{total}</span> invocations
        </div>
        <ListGroup className="mt-3 m0 flex-row">
          <ListGroupItem className="d-flex flex-fill flex-column align-items-center">
            <h5 className="m-0">
              <Badge color="success">{success}</Badge>
            </h5>
            <span>Success</span>
          </ListGroupItem>
          <ListGroupItem className="d-flex flex-fill flex-column align-items-center">
            <h5 className="m0">
              <Badge color="danger">{failure}</Badge>
            </h5>
            <span>Error</span>
          </ListGroupItem>
        </ListGroup>
      </div>
    );
  }

  navLinkClickHandle = option => {
    this.setState({
      selected: option
    });
    this.props.changeFunctionInvocationTimePeriod(OPTIONS[option]);
  };
}
