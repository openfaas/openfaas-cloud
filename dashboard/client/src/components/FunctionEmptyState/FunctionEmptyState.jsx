import React from 'react';
import {Button, CardBody, CardTitle, Col, Row} from 'reactstrap';

import './FunctionEmptyState.css';
import {OFIcon} from "../OFIcon";
import {OFBadge} from "../OFBadge";

const genSize = (size, order) => ({ size, order });

const FunctionEmptyState = () => {
  return (
    <CardBody className="FunctionEmptyState">
      <CardTitle tag="h2">
        Welcome to the OpenFaaS Cloud Dashboard!
      </CardTitle>
      <Row>
        <Col
          className="FunctionEmptyState__row__col FunctionEmptyState__row__col--img"
           xs={genSize(12, 1)}
           md={genSize(6, 1)}
        >
          <OFIcon />
        </Col>
        <Col
          className="FunctionEmptyState__row__col FunctionEmptyState__row__col--img"
          xs={genSize(12, 5)}
          md={genSize(6, 2)}
        >
          <OFBadge />
        </Col>
        <Col
          className="FunctionEmptyState__row__col"
          xs={genSize(12, 2)}
          md={genSize(6, 3)}
        >
          <h5>
            Create a new function to start!
          </h5>
        </Col>
        <Col
          className="FunctionEmptyState__row__col"
          xs={genSize(12, 6)}
          md={genSize(6, 4)}
        >
          <h5>
            Learn how to create, build, deploy a function!
          </h5>
        </Col>
        <Col
          className="FunctionEmptyState__row__col"
          xs={genSize(12, 3)}
          md={genSize(6, 5)}
        >
          <p className="FunctionEmptyState__table__row__col__p">
            Create a brand new function with one of the supported language templates.
          </p>
        </Col>
        <Col
          className="FunctionEmptyState__row__col"
          xs={genSize(12, 7)}
          md={genSize(6, 6)}
        >
          <p className="FunctionEmptyState__table__row__col__p">
            This is a good place to start if you're new to the project and want to build a working knowledge of how to
            start shipping serverless functions.
          </p>
        </Col>
        <Col
          className="FunctionEmptyState__row__col mb-4 mb-md-0"
          xs={genSize(12, 4)}
          md={genSize(6, 7)}
        >
          <Button tag="a" href="https://gist.github.com/alexellis/d648d927c34c082bb5d965f06b818026" target="_blank">
            CREATE A FUNCTION
          </Button>
        </Col>
        <Col
          className="FunctionEmptyState__row__col"
          xs={genSize(12, 8)}
          md={genSize(6, 8)}
        >
          <Button tag="a" href="https://docs.openfaas.com/tutorials/workshop/" target="_blank">
            WORKSHOP
          </Button>
        </Col>
      </Row>
    </CardBody>
  )
};

export {
  FunctionEmptyState,
}
