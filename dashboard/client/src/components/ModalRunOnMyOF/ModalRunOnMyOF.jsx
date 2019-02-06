import React, { Component } from 'react';
import {
  Modal,
  ModalBody,
  ModalHeader,
  FormFeedback,
  Form,
  Input,
  Button,
  Row,
  Col,
} from 'reactstrap';
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faCopy } from '@fortawesome/free-solid-svg-icons';

class ModalRunOnMyOF extends Component {
  state = {
    valid: false,
  };

  constructor() {
    super();

    this.handleCopyClick = this.handleCopyClick.bind(this);
    this.closeModal = this.closeModal.bind(this);
  }

  handleCopyClick(e) {
    e.preventDefault();

    this.input.select();

    document.execCommand('copy');

    this.setState({
      valid: true,
    }, () => {
      setTimeout(() => {
        this.setState({
          valid: false,
        })
      }, 1500)
    })
  };

  closeModal() {
    this.setState({ active: false });
  }

  render() {
    const { shortName, image, gitOwner } = this.props.fn;
    const code = `{
mkdir -p /tmp/openfaas-cloud/${gitOwner}/${shortName}/
cd /tmp/openfaas-cloud/${gitOwner}/${shortName}/

cat > stack.yml <<EOF
provider:
  name: faas
  gateway: http://127.0.0.1:8080

functions:
  ${shortName}:
    skip_build: true
    image: ${ image }
EOF

faas-cli deploy
}`;

    return (
      <Modal isOpen={this.props.state} toggle={this.props.closeModal} className={this.props.className}>
        <ModalHeader toggle={this.props.closeModal}>
          Run on my <strong>OpenFaaS</strong>
        </ModalHeader>
        <ModalBody>
          <p>
            To run this function on your local <strong>OpenFaaS</strong> cluster copy and paste the below into a bash terminal.
          </p>
          <Form>
            <Row noGutters className="align-items-end">
              <Col>
                <Input
                  readOnly
                  className="text-monospace"
                  innerRef={(node) => {
                    this.input = node;
                  }}
                  type="textarea"
                  valid={this.state.valid}
                  value={code}
                  bsSize="sm"
                  rows="8"
                />
                <FormFeedback valid tooltip>
                  Copied to clipboard
                </FormFeedback>
              </Col>
              <Col className="col-auto position-relative">
                <Button onClick={this.handleCopyClick} className="radius-right-5 ml-2">
                  <span>Copy</span>
                  <FontAwesomeIcon className="ml-2" icon={faCopy} />
                </Button>
              </Col>
            </Row>
          </Form>
        </ModalBody>
      </Modal>
    );
  }
}

ModalRunOnMyOF.defaultProps = {
  fn: {},
};

export {
  ModalRunOnMyOF
};
