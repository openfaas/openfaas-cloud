import React, { Component } from 'react';
import { FunctionTable } from '../components/FunctionTable';
import { FunctionEmptyState } from "../components/FunctionEmptyState";
import { functionsApi } from '../api/functionsApi';
import {
  Card,
  CardHeader,
  CardBody,
  CardText,
} from 'reactstrap';
import {faExclamationTriangle} from "@fortawesome/free-solid-svg-icons";
import {FontAwesomeIcon} from "@fortawesome/react-fontawesome";

export class FunctionsOverviewPage extends Component {
  constructor(props) {
    super(props);

    const { user } = props.match.params;
    this.state = {
      isLoading: true,
      fns: [],
      authError: false,
      user,
    };
  }

  componentDidMount() {
    this.setState({ isLoading: true });


    functionsApi.fetchFunctions(this.state.user)
    .then(res => {
      let functions = [];
      res.forEach( (set) => {
        set.forEach((item) => {
          functions.push(item);
        });
      });

      this.setState({ isLoading: false, fns: functions });
    })
    .catch((e) => {
      if (e.response.status === 403) {
        this.setState({isLoading: false, fns: [], authError: true})
      } else {
        console.error(e);
      }
    });
  }

  linkBuilder(location) {
      return `/dashboard/${location}`
  }

  renderContentView() {
    const { user, isLoading, fns, authError } = this.state;

    if (!isLoading && authError) {
      return (
          <Card>
            <CardHeader className="color-failure">
              <FontAwesomeIcon icon={faExclamationTriangle} /> Error: You do not have valid permissions for <span id="username">{ user }</span>
            </CardHeader>
          </Card>
      )
    }

    if (!isLoading && fns.length === 0) {
      return (
        <FunctionEmptyState />
      )
    }

    return (
      <CardBody>
        <CardText>
          Welcome to the OpenFaaS Cloud Dashboard! Click on a function for more details.
        </CardText>
        <FunctionTable isLoading={isLoading} fns={fns} user={user} />
      </CardBody>
    )
  }

  render() {
    const { user } = this.state;

    return (
      <Card outline color="success">
        <CardHeader className="bg-success color-success">
              Functions for {user}
        </CardHeader>

        { this.renderContentView() }
      </Card>
    );
  }
}
