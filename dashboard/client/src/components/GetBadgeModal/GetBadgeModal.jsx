import React, { Component } from 'react';
import {
  Modal,
  ModalBody,
  ModalHeader,
  FormFeedback,
  Form,
  Input,
  InputGroup,
  Button,
  InputGroupAddon,
} from 'reactstrap';
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faCopy } from '@fortawesome/free-solid-svg-icons';

const badgeImage = "https://img.shields.io/badge/openfaas-cloud-blue.svg";
const badgeLink = "https://www.openfaas.com";
const badgeMd = `[![OpenFaaS](${badgeImage})](${badgeLink})`;

class GetBadgeModal extends Component {
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
    return (
      <Modal isOpen={this.props.state} toggle={this.props.closeModal} className={this.props.className}>
        <ModalHeader toggle={this.props.closeModal}>
          Get Badge
        </ModalHeader>
        <ModalBody>
          <p>
            Add the following markdown snippet to your README.md in your Git repo to show you are using OpenFaaS Cloud.
          </p>
          <p>
            <a href={badgeLink}>
              <img src={badgeImage} alt="Powered by OpenFaaS &reg; Cloud" />
            </a>
          </p>
          <Form>
            <InputGroup>
              <Input
                readOnly
                innerRef={(node) => {
                  this.input = node;
                }}
                type="text"
                valid={this.state.valid}
                value={badgeMd}
              />
              <InputGroupAddon addonType="append">
                <Button onClick={this.handleCopyClick} className="radius-right-5">
                  <span>Copy</span>
                  <FontAwesomeIcon className="ml-2" icon={faCopy} />
                </Button>
              </InputGroupAddon>
              <FormFeedback valid tooltip>
                Link copied
              </FormFeedback>
            </InputGroup>
          </Form>
        </ModalBody>
      </Modal>
    );
  }
}

export {
  GetBadgeModal
};
