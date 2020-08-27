import {
  Configs,
  KubernetesObject,
  isKubernetesObject,
  generalResult,
  getAnnotation,
  addAnnotation,
  removeAnnotation,
} from 'kpt-functions';
import {
  DumpOptions,
  safeDump,
  safeLoad,
} from 'js-yaml';
import rw from 'rw';
import {
  spawnSync,
} from 'child_process';

const IGNORE_MAC = 'ignore-mac';
const VERBOSE = 'verbose';
const temp_path = '/tmp/tmp.yaml';

//var debuglog = require('debuglog')('sops');

interface SopsKubernetesObject extends KubernetesObject {
}

function isSopsKubernetesObject(o: any): o is SopsKubernetesObject {
  return o && o.sops;
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
    rw.writeFile(path, data, 'utf8', (err: any) => {
      if (err) return reject(err);
      resolve();
    });
  });
}

export async function sops(configs: Configs) {
  // Validate config data and read arguments.
  const args = readSopsArguments(configs);

  // Put all documents into the local place and
  // cleanup the configs storage. This is to keep
  // the order of documents
  let allDocs = configs.getAll();
  configs.deleteAll();

  for (const object of allDocs) {
    if (isSopsKubernetesObject(object)) {
    //debuglog('trying to decrypt: apiVersion: %s, kind: %s, name: %s',
    //        object.apiVersion,
    //        object.kind,
    //        object.metadata.name);
      // this function
      await decryptAndInsertSopsKubernetesObject(args, configs, object);
    } else {
      // add it back
      configs.insert(object);
    }
  }
}

function detachConfigKubernetesIoAnnotations(object: SopsKubernetesObject): Map<string, string> {
  const annotations: string[] = [
    'config.kubernetes.io/index',
    'config.kubernetes.io/path',
    'config.k8s.io/id',
  ];

  let detached = new Map<string, string>();

  for (const annotation of annotations) {
    const value = getAnnotation(object, annotation);
    if (value !== undefined) {
      detached.set(annotation, value);
      removeAnnotation(object, annotation);
    }
  }
  return detached;
}

function attachDetachedAnnotations(object: SopsKubernetesObject, detached: Map<string, string>) {
  detached.forEach((value: string, key: string) => {
    addAnnotation(object, key, value);
  })
}

async function decryptAndInsertSopsKubernetesObject(args: string[], configs: Configs, object: SopsKubernetesObject) {
  let error;
  args.push(...['-d', temp_path]);

  // write encrypted file to the temp file
  const detached = detachConfigKubernetesIoAnnotations(object);
  const stringifiedObject = safeDump(object, YAML_STYLE);
  await writeFile(temp_path, stringifiedObject);
  attachDetachedAnnotations(object, detached);

  try {
    // run sops
    //debuglog('calling sops with args: %s', args.toString())
    const child = spawnSync('sops', args);
    error = child.stderr;
    // read the decrypted yaml from stdout and parse
    let decryptedObject = safeLoad(child.stdout);
    if (object && isKubernetesObject(decryptedObject)) {
      attachDetachedAnnotations(decryptedObject, detached);
      configs.insert(decryptedObject);
      return;
    }
  } catch (err) {
    //debuglog('got an error from exception: %s', err)
    configs.addResults(generalResult(`Exception for apiVersion: ${object.apiVersion}, kind: ${object.kind}, name: ${object.metadata.name}: ${err}`, 'error'));
  }
  if (error && error.length > 0) {
    //debuglog('got an error from stdout: %s', error)
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
    if (key === VERBOSE ||
        key === IGNORE_MAC) {
      args.push('--'+key);
    } else {
      args.push('--'+key);
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
ignore-mac: true [Optional: default empty]
verbose: true [Optional: default empty]
keyservice value [Optional: default empty]

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
        image: (put here the image url)
    config.kubernetes.io/local-config: "true"
data:
  verbose: true
`;
