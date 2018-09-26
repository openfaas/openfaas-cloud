import React, { Component } from 'react';
import { FunctionTable } from '../components/FunctionTable';
import { functionsApi } from '../api/functionsApi';
import { Card, CardHeader, CardBody, CardText } from 'reactstrap';

export class FunctionsOverviewPage extends Component {
  constructor(props) {
    super(props);

    const { user } = props.match.params;
    this.state = {
      isLoading: true,
      fns: [],
      user,
    };
  }

  componentDidMount() {
    this.setState({ isLoading: true });

    functionsApi.fetchFunctions(this.state.user).then(res => {
      this.setState({ isLoading: false, fns: res });
    });
  }

  render() {
    const { user, isLoading, fns } = this.state;

    return (
      <Card outline color="success">
        <CardHeader className="bg-success color-success">
          Functions for <span id="username">{user}</span>
        </CardHeader>
        <CardBody>
          <CardText>
            Welcome to the OpenFaaS Cloud Dashboard! Click on a function for more details.
          </CardText>
          <FunctionTable isLoading={isLoading} fns={fns} user={user} />
        </CardBody>
      </Card>
    );
  }
}
