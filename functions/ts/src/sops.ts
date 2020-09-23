import {
  Configs,
  KubernetesObject,
  isKubernetesObject,
  generalResult,
  getAnnotation,
  addAnnotation,
  removeAnnotation,
} from 'kpt-functions';
import { DumpOptions, safeDump, safeLoad } from 'js-yaml';
import rw from 'rw';
import { spawnSync, execSync } from 'child_process';

const TEMP_PATH = '/tmp/tmp.yaml';

const IGNORE_MAC = 'ignore-mac';
const VERBOSE = 'verbose';
const OVERRIDE_PREEXEC_CMD = 'override-preexec-cmd';
const OVERRIDE_DETACHED_ANNOTATIONS = 'override-detached-annotations';

// pre-exec command may be overriden from config
let preExecCmd =
  '[ "$SOPS_IMPORT_PGP" == "" ] || (echo "$SOPS_IMPORT_PGP" | gpg --import)';

// list of annotations that will be detached before decryption
// this is needed, because tools like kpt add some annotations
// during processing, but sops fails trying to decrypt them.
// After decryption they will be attached back unchanged.
// it can be overriden from config
// TODO: there is a discusion about the complete list here:
// https://github.com/kubernetes-sigs/kustomize/issues/2996
// Please update the list when it is finished.
let detachedAnnotations: string[] = [
  'config.kubernetes.io/index',
  'config.kubernetes.io/path',
  'config.k8s.io/id',
  'kustomize.config.k8s.io/id',
];

interface SopsKubernetesObject extends KubernetesObject {}

function isSopsKubernetesObject(
  o: KubernetesObject
): o is SopsKubernetesObject {
  return o && o.hasOwnProperty('sops');
}

const YAML_STYLE: DumpOptions = {
  // indentation width to use (in spaces).
  indent: 2,
  // when true, will not add an indentation level to array elements.
  noArrayIndent: true,
  // 'undefined' is an invalid value for safeDump.
  // TODO(frankf): Explore ways to make this safer.
  // Either reason about not having 'undefined' in all cases OR
  // only skip 'undefined'.
  skipInvalid: true,
  // unset lineWidth from default of 80 to avoid reformatting
  lineWidth: -1,
  // avoid refs because many YAML parsers in the k8s ecosystem don't support them
  noRefs: true,
};

async function writeFile(path: string, data: string): Promise<void> {
  return new Promise((resolve, reject) => {
    rw.writeFile(path, data, 'utf8', (err: object) => {
      if (err) return reject(err);
      resolve();
    });
  });
}

export async function sops(configs: Configs) {
  // Validate config data and read arguments.
  const args = readSopsArguments(configs);

  // run preexec if it's not empty
  if (preExecCmd !== '') {
    execSync(preExecCmd);
  }

  // Put all documents into the local place and
  // cleanup the configs storage. This is to keep
  // the order of documents
  const allDocs = configs.getAll();
  configs.deleteAll();

  for (const object of allDocs) {
    if (isSopsKubernetesObject(object)) {
      // this function
      await decryptAndInsertSopsKubernetesObject(args, configs, object);
    } else {
      // add it back
      configs.insert(object);
    }
  }
}

function detachAnnotations(object: SopsKubernetesObject): Map<string, string> {
  const detached = new Map<string, string>();

  for (const annotation of detachedAnnotations) {
    const value = getAnnotation(object, annotation);
    if (value !== undefined) {
      detached.set(annotation, value);
      removeAnnotation(object, annotation);
    }
  }
  return detached;
}

function attachDetachedAnnotations(
  object: SopsKubernetesObject,
  detached: Map<string, string>
) {
  detached.forEach((value: string, key: string) => {
    addAnnotation(object, key, value);
  });
}

async function decryptAndInsertSopsKubernetesObject(
  args: string[],
  configs: Configs,
  object: SopsKubernetesObject
) {
  let error;
  args.push(...['-d', TEMP_PATH]);

  // write encrypted file to the temp file
  const detached = detachAnnotations(object);
  const stringifiedObject = safeDump(object, YAML_STYLE);
  await writeFile(TEMP_PATH, stringifiedObject);
  attachDetachedAnnotations(object, detached);

  try {
    // run sops
    const child = spawnSync('sops', args);
    error = child.stderr;
    // read the decrypted yaml from stdout and parse
    const decryptedObject = safeLoad(child.stdout);
    if (object && isKubernetesObject(decryptedObject)) {
      attachDetachedAnnotations(decryptedObject, detached);
      configs.insert(decryptedObject);
      return;
    }
  } catch (err) {
    configs.addResults(
      generalResult(
        `Exception for apiVersion: ${object.apiVersion}, kind: ${object.kind}, name: ${object.metadata.name}: ${err}`,
        'error'
      )
    );
  }
  if (error && error.length > 0) {
    configs.addResults(
      generalResult(
        `Sops command results in error for\n${stringifiedObject}\n ${error.toString()}`,
        'error'
      )
    );
  }
  // if we're here - there was an error
  // putting the original object back
  configs.insert(object);
}

function readSopsArguments(configs: Configs) {
  const args: string[] = [];
  const configMap = configs.getFunctionConfigMap();
  if (!configMap) {
    return args;
  }
  configMap.forEach((value: string, key: string) => {
    if (key === OVERRIDE_PREEXEC_CMD) {
      preExecCmd = value;
    } else if (key === OVERRIDE_DETACHED_ANNOTATIONS) {
      detachedAnnotations = value.replace(/\s/g, '').split(',');
    } else if (key === VERBOSE || key === IGNORE_MAC) {
      args.push('--' + key);
    } else {
      args.push('--' + key);
      args.push(value);
    }
  });
  return args;
}

sops.usage = `
Sops function (see https://github.com/mozilla/sops).
So far supports only decrypt operation:
runs sops -d for all documents that have field 'sops:' and put the decrypted result back.

Can be configured using a ConfigMap with the following flags:
ignore-mac: true [Optional: default empty] Ignore Message Authentication Code during decryption.
verbose: true [Optional: default empty]    Enable sops verbose logging output.
keyservice value [Optional: default empty] Specify the key services to use in addition to the local one.
                                           Can be specified more than once.
                                           Syntax: protocol://address. Example: tcp://myserver.com:5000
override-detached-annotations: [Optional:
default see detachedAnnotations var]       The list of annotations that didn't present when the document
                                           was encrypted, but added by different tools later. The function
                                           will detach them before decryption and added unchanged
                                           after successfull decryption. This allows sops to check the
                                           consistency of the decrypted document.

For more details see 'sops --help'.

Example:

To decrypt the documents use:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image: gcr.io/kpt-functions/sops
    config.kubernetes.io/local-config: "true"
data:
  verbose: true
`;
