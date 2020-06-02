import { Configs, kubernetesObjectResult, TestRunner } from 'kpt-functions';
import { kubeval } from './kubeval';
import { Namespace, PodTemplateSpec } from './gen/io.k8s.api.core.v1';
import { Deployment, DeploymentSpec } from './gen/io.k8s.api.apps.v1';
import {
  LabelSelector,
  ObjectMeta,
} from './gen/io.k8s.apimachinery.pkg.apis.meta.v1';

const RUNNER = new TestRunner(kubeval);

describe('kubeval', () => {
  it('handles objects without errors', async () => {
    await RUNNER.assert(
      new Configs([Namespace.named('something')]),
      new Configs([Namespace.named('something')])
    );
  });
  it('reacts on errors', async () => {
    const deployment = new Deployment({
      metadata: new ObjectMeta({
        name: 'something',
      }),
      spec: new DeploymentSpec({
        selector: new LabelSelector(),
        template: new PodTemplateSpec(),
        // schema violation:
        paused: ('horse' as unknown) as boolean,
      }),
    });
    await RUNNER.assert(
      new Configs([deployment]),
      new Configs([deployment], undefined, [
        kubernetesObjectResult(
          'Invalid type. Expected: [boolean,null], given: string',
          deployment,
          {
            path: 'spec.paused',
          }
        ),
      ])
    );
  });
});
