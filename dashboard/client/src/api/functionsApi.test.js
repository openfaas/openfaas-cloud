import axios from 'axios';
import { functionsApi } from './functionsApi';
jest.mock('axios');

const user = 'some-user';
const baseFunction = {
  name: `${user}-some-function`,
  image: `ofcommunity/${user}-some-function-repo-some-function:latest-xxxxxxx`,
  invocationCount: 0,
  replicas: 1,
};
const baseFunctionLabels = {
  'com.openfaas.cloud.git-cloud': '1',
  'com.openfaas.cloud.git-deploytime': '1536401680',
  'com.openfaas.cloud.git-owner': user,
  'com.openfaas.cloud.git-repo': 'some-function-repo',
  'com.openfaas.cloud.git-sha': 'abcdefghijklmnopqrstuvwxyz0123456789ABCD',
  'com.openfaas.cloud.git-private-repo': 'true',
  app: `${user}-some-function`,
  faas_function: `${user}-some-function`,
  uid: '111111111',
};

describe('functionsApi', () => {
  describe('fetchFuncions', () => {
    it('parses the shortname of the function', async () => {
      // Arrange
      const functionResponse = {
        ...baseFunction,
        labels: { ...baseFunctionLabels },
      };
      axios.get.mockImplementation(() =>
        Promise.resolve({ data: [functionResponse] })
      );

      // Act
      const response = await functionsApi.fetchFunctions(user);

      // Assert
      const [actual] = response;
      expect(actual.shortName).toEqual('some-function');
    });
    it('reorders the function with the Git-DeployTime to be newest first', async () => {
      // Arrange
      const functionResponse1 = {
        ...baseFunction,
        labels: { ...baseFunctionLabels },
      };
      functionResponse1.name = `${user}-older-function`;
      functionResponse1.labels['com.openfaas.cloud.git-deploytime'] = '1500000000';

      const functionResponse2 = {
        ...baseFunction,
        labels: { ...baseFunctionLabels },
      };
      functionResponse2.name = `${user}-newer-function`;
      functionResponse2.labels['com.openfaas.cloud.git-deploytime'] = '1500000001';

      axios.get.mockImplementation(() =>
        Promise.resolve({ data: [functionResponse1, functionResponse2] })
      );

      // Act
      const response = await functionsApi.fetchFunctions(user);

      // Assert
      const [first, second] = response;
      // Check that the newer function comes first
      expect(first.name).toEqual(functionResponse2.name);
      expect(second.name).toEqual(functionResponse1.name);
    });
  });
});
