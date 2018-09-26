import React from 'react';
import { Progress } from 'reactstrap';

const ReplicasProgress = ({ fn }) => {
  const { replicas, minReplicas, maxReplicas } = fn;

  const lowerReplicas = minReplicas && minReplicas.length && parseInt(minReplicas, 10) ? minReplicas : 1;
  const upperReplicas = maxReplicas && maxReplicas.length && parseInt(maxReplicas, 10) ? maxReplicas : 20;
  const percentage = Math.floor((lowerReplicas / upperReplicas) * 100);

  let status = null;

  if (percentage < 66) {
    status = 'success';
  } else if (66 <= percentage && percentage < 90) {
    status = 'warning';
  } else {
    status = 'danger';
  }

  return (
    <div className="d-flex align-items-center">
      <Progress color={status} value={percentage} className="flex-grow-1"/>
      <div className="flex-grow-0 flex-shrink-1 pl-2">
        { replicas }/{ upperReplicas }
      </div>
    </div>
  );
};

export {
  ReplicasProgress
};
