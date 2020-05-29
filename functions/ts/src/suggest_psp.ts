import { Configs, kubernetesObjectResult } from 'kpt-functions';
import { isPodSecurityPolicy } from './gen/io.k8s.api.policy.v1beta1';

export async function suggestPsp(configs: Configs) {
  // Iterate over all PodSecurityPolicy objects in the input and flag any
  // that do not have 'allowPrivilegeEscalation' field set to true.
  const results = configs
    .get(isPodSecurityPolicy)
    .filter(psp => psp.spec && psp.spec.allowPrivilegeEscalation !== false)
    .map(psp =>
      kubernetesObjectResult(
        'Suggest explicitly disabling privilege escalation',
        psp,
        { path: 'spec.allowPrivilegeEscalation', suggestedValue: false },
        'warn',
        { category: 'security' }
      )
    );

  configs.addResults(...results);
}

suggestPsp.usage = `
Lints PodSecurityPolicy by suggesting 'spec.allowPrivilegeEscalation' field be set to 'false'.
`;
