import React from 'react';
import PropTypes from 'prop-types';

import { NotificationManager } from 'react-notifications';

const copyRepoLink = (value, ignoreNotification) => {
    const el = document.createElement('textarea');
    el.value = value;

    el.setAttribute('readonly', '');
    el.style.position = 'absolute';
    el.style.left = '-9999px';

    document.body.appendChild(el);

    el.select();
    document.execCommand('copy');

    document.body.removeChild(el);

    document.queryCommandSupported('copy');

    if (ignoreNotification !== true) {
        NotificationManager.success('Link copied', null, 1500);
    }
};

const CopyText = ({ children, copyValue, ignoreNotification }) => {
    return (
        <span onClick={() => copyRepoLink(copyValue, ignoreNotification)}>
            {children}
        </span>
    );
};

CopyText.propTypes = {
    copyValue: PropTypes.string,
    ignoreNotification: PropTypes.bool,
};

CopyText.defaultProps = {
    children: null,
    copyValue: '',
    ignoreNotification: false,
};

export default CopyText;
