import { Configs, TestRunner } from 'kpt-functions';
import { kubeval } from './kubeval';
import { Namespace, ConfigMap } from './gen/io.k8s.api.core.v1';

const RUNNER = new TestRunner(kubeval);

describe('kubeval', () => {
  it('handles undefined function config', async () => {
    const input = new Configs(undefined, undefined);

    await RUNNER.assert(input, new Configs(undefined));
  });

  const namespace = Namespace.named('namespace');
  it('outputs error given namespace function config', async () => {
    const input = new Configs(undefined, namespace);

    await RUNNER.assert(input, new Configs(undefined, namespace), Error);
  });

  const emptyConfigMap = new ConfigMap({ metadata: { name: 'config' } });
  it('handles empty function config', async () => {
    const input = new Configs(undefined, emptyConfigMap);

    await RUNNER.assert(input, new Configs(undefined, emptyConfigMap));
  });

  const configMap = new ConfigMap({
    metadata: { name: 'config' },
    data: { ignore_missing_schemas: 'true', strict: 'true' },
  });
  it('handles empty configs', async () => {
    const input = new Configs([], configMap);
    const output = new Configs([], configMap);
    await RUNNER.assert(input, output);
  });
});
