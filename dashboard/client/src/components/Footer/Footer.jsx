import React from 'react';
import { Container, Row, Col } from 'reactstrap';

const Footer = () => (
  <Container className="padding-12-0">
    <Row>
      <Col>
        Powered by <a href="https://www.openfaas.com">OpenFaaS</a>
      </Col>
    </Row>
  </Container>
);

export {
  Footer
};
