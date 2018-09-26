import React from 'react';
import PropTypes from 'prop-types';

import { Card, CardHeader, CardBody, Row, Col } from 'reactstrap';

const FunctionOverviewPanel = ({ children, headerText, headerIcon, button, className, bodyClassName }) => {
  return (
    <Card className={`height-100 ${className}`}>
      <CardHeader
        tag="h6"
        className="flex align-items-center justify-content-space-between min-height-49"
      >
        <span>
          { headerIcon }
          { headerText }
        </span>
        { button }
      </CardHeader>
      <CardBody className={bodyClassName}>
        { children }
      </CardBody>
    </Card>
  );
};

const getOppositeSize = (size) => {
  if (size >= 12 || size <= 0) {
    return size;
  }

  return 12 - size;
};

const MetaList = ({ list, sizes }) => {
  let currentSizes = sizes;

  if (sizes === undefined) {
    currentSizes = {
      xs: 12,
      sm: 3,
      md: 2,
      lg: 4,
      xl: 3,
    }
  }

  return (
    <div>
      { list.map(({ label, value, renderValue }) => {
        const { xs, sm, md, lg, xl } = currentSizes;

        return (
          <Row noGutters key={label} className="py-1">
            <Col xs={xs} sm={sm} md={md} lg={lg} xl={xl} className="is-bold">
              { label }
            </Col>
            <Col
              xs={getOppositeSize(xs)}
              sm={getOppositeSize(sm)}
              md={getOppositeSize(md)}
              lg={getOppositeSize(lg)}
              xl={getOppositeSize(xl)}
              className="pl-xl-2"
            >
              { typeof renderValue === 'function' ? renderValue() : value }
            </Col>
          </Row>
        )
      })}
    </div>
  )
};

MetaList.propTypes = {
  list: PropTypes.arrayOf(PropTypes.shape({
    label: PropTypes.string.isRequired,
    value: PropTypes.string,
    renderValue: PropTypes.func,
  })),
};

FunctionOverviewPanel.MetaList = MetaList;

FunctionOverviewPanel.propTypes = {
  header: PropTypes.oneOfType([
    PropTypes.element,
    PropTypes.string,
  ]),
  headerText: PropTypes.string,
  headerIcon: PropTypes.element,
  button: PropTypes.element,
};

FunctionOverviewPanel.defaultProps = {
  button: null,
};

export {
  FunctionOverviewPanel
};