import {
  Configs,
  Result,
  KubernetesObject,
  kubernetesObjectResult,
  SOURCE_PATH_ANNOTATION,
  SOURCE_INDEX_ANNOTATION,
  ID_ANNOTATION,
  LEGACY_SOURCE_PATH_ANNOTATION,
  LEGACY_SOURCE_INDEX_ANNOTATION,
  LEGACY_ID_ANNOTATION,
} from 'kpt-functions';
import {
  isResourceHierarchy as isV3ResourceHierarchy,
  ResourceHierarchy as V3ResourceHierarchy,
} from './gen/com.google.cloud.blueprints.v1alpha3';
import {
  isResourceHierarchy as isV2ResourceHierarchy,
  ResourceHierarchy as V2ResourceHierarchy,
} from './gen/dev.cft.v1alpha2';
import {
  isResourceHierarchy as isV1ResourceHierarchy,
  ResourceHierarchy as V1ResourceHierarchy,
} from './gen/dev.cft.v1alpha1';
import { FolderList } from './gen/com.google.cloud.cnrm.resourcemanager.v1beta1';
import { ObjectMeta } from 'kpt-functions/dist/src/gen/io.k8s.apimachinery.pkg.apis.meta.v1';

// Representation of a node in the hierarchy tree
interface HierarchyNode {
  children: HierarchyNode[];
  parent?: HierarchyNode;
  config?: KubernetesObject;
  kind?: string;
  name: string;
}

export interface Annotations {
  [key: string]: string;
}

// Folder GK
export const FOLDER_GROUP = 'resourcemanager.cnrm.cloud.google.com';
export const FOLDER_KIND = 'Folder';

// Depends on anno
export const DEPENDS_ON_ANNOTATION = 'config.kubernetes.io/depends-on';

/**
 * Entrypoint for kpt function business logic. See `usage` field for more details.
 *
 * @param configs In-memory document store for Kubernetes objects
 */
export async function generateFolders(configs: Configs) {
  configs.get(isV1ResourceHierarchy).forEach((hierarchy) => {
    configs.addResults(oldHierarchyWarning(hierarchy));
    const layers: string[] = hierarchy.spec.layers;

    // Root node is the organization
    const root: HierarchyNode = {
      children: [],
      kind: 'Organization',
      name: `${hierarchy.spec.organization}`, // Annotation expects string type
    };

    // Represent results as a binary tree
    const errorResult = generateV1HierarchyTree(root, layers, 0, hierarchy, []);

    // Report any errors; create configs + delete ResourceHierarchy resource if
    // no errors reported.
    if (errorResult) {
      configs.addResults(errorResult);
    } else {
      insertConfigs(root, configs);
    }
  });
  // both v2 and v3 ResourceHierarchy generates the same folder hierarchy except v3 uses native KCC refs
  [
    ...configs.get(isV3ResourceHierarchy),
    ...configs.get(isV2ResourceHierarchy),
  ].forEach((hierarchy) => {
    // if v2 add warning to upgrade
    if (isV2ResourceHierarchy(hierarchy)) {
      configs.addResults(oldHierarchyWarning(hierarchy));
    }
    if (
      hierarchy.spec.parentRef === undefined ||
      hierarchy.spec.parentRef.external === undefined
    ) {
      configs.addResults(badParentErrorResult(hierarchy));
      return;
    }
    if (
      hierarchy.spec.parentRef.kind !== undefined &&
      !['Organization', 'Folder'].includes(hierarchy.spec.parentRef.kind)
    ) {
      configs.addResults(badParentKindErrorResult(hierarchy));
      return;
    }
    try {
      generateHierarchyTree(hierarchy, configs);
    } catch (e) {
      if (e instanceof MissingSubteeError) {
        configs.addResults(missingSubtreeErrorResult(e.message, hierarchy));
      } else {
        throw e;
      }
    }
  });
}

/**
 * Generate a warning for a v1/v2 hierarchy to upgrade to v3
 * @param hierarchy The old hierarchy
 */
export function oldHierarchyWarning(hierarchy: KubernetesObject): Result {
  return kubernetesObjectResult(
    `ResourceHierarchy ${hierarchy.metadata.name} references an older Resource Hierarchy GroupVersion. Latest GroupVersion is blueprints.cloud.google.com/v1alpha3.`,
    hierarchy,
    undefined,
    'warn'
  );
}

/**
 * Generate an error for a hierarchy with an undefined parentRef
 * @param hierarchy The hierarchy to yield an error for
 */
export function badParentErrorResult(hierarchy: KubernetesObject): Result {
  return kubernetesObjectResult(
    `ResourceHierarchy ${hierarchy.metadata.name} has an undefined parentRef`,
    hierarchy,
    undefined,
    'error'
  );
}

/**
 * Generate an error for a hierarchy with an unsupported parentRef kind
 * @param hierarchy The hierarchy to yield an error for
 */
export function badParentKindErrorResult(hierarchy: KubernetesObject): Result {
  return kubernetesObjectResult(
    `ResourceHierarchy ${hierarchy.metadata.name} has an unsupported parentRef kind`,
    hierarchy,
    undefined,
    'error'
  );
}

/**
 * Generate an error for a hierarchy with a missing subtree
 * @param hierarchy The hierarchy to yield a missing subtree
 */
export function missingSubtreeErrorResult(
  subtree: string,
  hierarchy: KubernetesObject
): Result {
  return kubernetesObjectResult(
    `ResourceHierarchy ${hierarchy.metadata.name} references non-existent subtree "${subtree}"`,
    hierarchy,
    undefined,
    'error'
  );
}

class MissingSubteeError extends Error {
  constructor(message?: string) {
    super(message);
    Object.setPrototypeOf(this, new.target.prototype);
    this.name = MissingSubteeError.name;
  }
}

/**
 * Creates a copy of the annotations object and removes annotations that generated
 * resources do not inherit from the ResourceHierarchy resource.
 * @param annotations The KRM annotations structure.
 * @returns The copy of annotations with non-inheritable keys removed.
 */
function filterNonInheritableAnnotations(
  annotations: Annotations
): Annotations {
  const copy: Annotations = { ...annotations };
  delete copy['config.kubernetes.io/local-config'];
  delete copy['config.k8s.io/function'];
  // Do not inherit kpt SDK's internal annotations.
  delete copy[ID_ANNOTATION];
  delete copy[LEGACY_ID_ANNOTATION];
  delete copy[SOURCE_PATH_ANNOTATION];
  delete copy[LEGACY_SOURCE_PATH_ANNOTATION];
  delete copy[SOURCE_INDEX_ANNOTATION];
  delete copy[LEGACY_SOURCE_INDEX_ANNOTATION];
  return copy;
}

/**
 * Creates a representation of the resulting folder hierarchy from the
 * ResourceHierarchy object in a tree data structure. Each node contains the
 * corresponding config to generate.
 *
 * @param hierarchy The ResourceHierarchy to generate configs for
 * @param configs The Config list to insert folders into
 */
function generateHierarchyTree(
  hierarchy: V2ResourceHierarchy | V3ResourceHierarchy,
  configs: Configs
): Result | undefined {
  const root: HierarchyNode = {
    children: [],
    kind: `${hierarchy.spec.parentRef.kind ?? 'Organization'}`, // if no kind is specified, default to Organization
    name: `${hierarchy.spec.parentRef.external}`, // Annotation expects string type
  };
  const namespace = hierarchy.metadata.namespace;
  const annotations: Annotations = filterNonInheritableAnnotations(
    hierarchy.metadata.annotations || {}
  );

  const subtrees: { [key: string]: HierarchyNode } = {};

  if (hierarchy.spec.subtrees !== undefined) {
    for (const name in hierarchy.spec.subtrees) {
      const node: HierarchyNode = {
        children: [],
        kind: 'Subtree',
        name,
      };
      const children = hierarchy.spec.subtrees[name];
      const subtree = generateTree(node, children as any[], subtrees);
      node.children = subtree.children;
      subtrees[name] = node;
    }
  }

  const tree = generateTree(root, hierarchy.spec.config, subtrees);

  const generateConfigs = (node: HierarchyNode, path: string[]) => {
    for (const child of node.children) {
      configs.insert(
        generateManifest(
          child.name,
          path,
          node,
          annotations,
          namespace,
          isV3ResourceHierarchy(hierarchy)
        )
      );
      generateConfigs(child, [...path, child.name]);
    }
  };

  generateConfigs(tree, []);
  return undefined;
}

/**
 * Add a child into the parent
 *
 * @param parent The parent node to attach the child to
 * @param child The child to append
 * @param subtrees A map of subtrees which can be referenced
 */
function addChild(
  parent: HierarchyNode,
  child: any,
  subtrees: { [key: string]: HierarchyNode }
) {
  if (child === null) {
    return;
  }
  if (typeof child === 'string') {
    parent.children.push({
      name: child,
      children: [],
    });
    return;
  }
  if (typeof child === 'object') {
    const name = Object.keys(child)[0];
    const node: HierarchyNode = {
      name,
      children: [],
    };
    const children = child[name];
    if (Array.isArray(children)) {
      generateTree(node, children, subtrees);
    } else if (typeof children === 'object') {
      const subtree = children['$subtree'];
      if (subtrees[subtree] === undefined) {
        throw new MissingSubteeError(subtree);
      }
      node.children = subtrees[subtree].children;
    }
    parent.children.push(node);
  }
}

/**
 * Generate a folder tree
 *
 * @param root The root node to build the tree from
 * @param children Top-level children to attach on the root
 * @param subtrees A map of subtrees which can be referenced
 */
function generateTree(
  root: HierarchyNode,
  children: any[],
  subtrees: { [key: string]: HierarchyNode }
): HierarchyNode {
  for (const child of children) {
    addChild(root, child, subtrees);
  }
  return root;
}

/**
 * Creates a representation of the resulting folder hierarchy from the
 * ResourceHierarchy object in a tree data structure. Each node contains the
 * corresponding config to generate.
 *
 * @param node The root node of the tree to generate. This is the organization.
 * @param layers The list of names of layers to create (levels of the tree).
 * @param layerIndex The index of which layer to process. Used for recursion.
 * @param hierarchy The object representing the ResourceHierarchy custom resource.
 * @param path The name of folders preceeding the current layer. Used to
 *  generate the unique name of the k8s resource.
 */
function generateV1HierarchyTree(
  node: HierarchyNode,
  layers: string[],
  layerIndex: number,
  hierarchy: V1ResourceHierarchy,
  path: string[]
): Result | undefined {
  if (layerIndex >= layers.length) {
    return undefined;
  }

  const layer = layers[layerIndex];
  const folders = hierarchy.spec.config[layer];

  // No layer config entry
  if (folders === undefined) {
    return {
      severity: 'error',
      message: `Layer "${layer}" has no corresponding entry config entry. Either
      add to spec.config.${layer} or remove it from spec.layers
      `,
    };
  }

  // Do not support annotation inheritance for v1;
  const annotations: Annotations = {};

  for (const folder of folders) {
    const child = {
      name: folder,
      children: [],
      parent: node,
      config: generateManifest(
        folder,
        path,
        node,
        annotations,
        hierarchy.metadata.namespace
      ),
    };

    const errorResult = generateV1HierarchyTree(
      child,
      layers,
      layerIndex + 1,
      hierarchy,
      [...path, folder]
    );

    if (errorResult) {
      return errorResult;
    }

    node.children.push(child);
  }

  return undefined;
}

/**
 * Crafts a k8s manifest based on the input data and node info.
 *
 * @param name The name of the folder
 * @param path A list of names of the ancestors of the current folder. Used to
 *             generate k8s resource name.
 * @param parent The parent node of the current folder.
 * @param namespace Namespace to generate the resource in.
 */
function generateManifest(
  name: string,
  path: string[],
  parent: HierarchyNode,
  annotations: Annotations,
  namespace?: string,
  nativeRef = false
): KubernetesObject {
  // Parent name is the metadata name
  const parentName = path.join('.') || parent.name;

  // hold annotationRef if any
  let annotationRef = {};
  // hold dependsOnAnno if any
  let annotationDependsOn = {};
  // hold nativeRefs if any
  let ref = {};

  if (nativeRef) {
    // root node has no parent and both org/folder ref is external
    if (path.length === 0) {
      ref =
        parent.kind === 'Organization'
          ? { organizationRef: { external: parentName } }
          : { folderRef: { external: parentName } };
    } else {
      ref = { folderRef: { name: normalize(parentName) } };
      if (namespace !== undefined) {
        // https://kpt.dev/reference/annotations/depends-on/?id=resource-reference
        annotationDependsOn = {
          [DEPENDS_ON_ANNOTATION]: `${FOLDER_GROUP}/namespaces/${namespace}/${FOLDER_KIND}/${normalize(
            parentName
          )}`,
        };
      }
    }
  } else {
    // generate annotation based ref
    const annotationName =
      parent.kind === 'Organization'
        ? 'cnrm.cloud.google.com/organization-id'
        : 'cnrm.cloud.google.com/folder-ref';
    annotationRef = { [annotationName]: normalize(parentName) };
  }

  let combinedAnnotations = {};
  if (
    Object.keys(annotations).length > 0 ||
    Object.keys(annotationRef).length > 0 ||
    Object.keys(annotationDependsOn).length > 0
  ) {
    combinedAnnotations = {
      annotations: { ...annotations, ...annotationRef, ...annotationDependsOn },
    };
  }

  const config = {
    apiVersion: FolderList.apiVersion,
    kind: 'Folder',
    metadata: {
      // TODO(jcwc): This only works up to 253 char (k8s name limit). Figure out
      //   how to handle the edge cases beyond the character limit.
      name: normalize([...path, name].join('.')),
      ...combinedAnnotations,
    } as ObjectMeta,
    spec: {
      displayName: name,
      ...ref,
    },
  };

  // Add namespace if provided
  if (namespace !== undefined) {
    config.metadata.namespace = namespace;
  }

  return config;
}

/**
 * Normalizes name to fit the K8s DNS subdomain naming requirements
 *
 * @param name Non-normalized name
 */
export function normalize(name: string) {
  name = name.toLowerCase();
  name = name.replace(/['"]/g, '');
  name = name.replace(/[_ ]/g, '-');
  name = name.replace(/[^a-z0-9\.\- ]/g, '');
  return name;
}

/**
 * Iterates through the tree of configs and inserts them into the output config
 * result.
 *
 * @param root The root node of the tree
 * @param configs In-memory document store for Kubernetes objects
 */
function insertConfigs(root: HierarchyNode, configs: Configs): void {
  if (root === undefined) return;

  if (root.config !== undefined) {
    configs.insert(root.config);
  }

  for (const child of root.children) {
    insertConfigs(child, configs);
  }
}

generateFolders.usage = `
This function translates the "ResourceHierarchy" custom resource and transforms
it to the resulting "Folder" custom resources constituting the hierarchy. Post
translation, it'll be necessary to use the "kpt-folder-parent" function to
translate the results into Cork configs.

Example configuration:

# hierarchy.yaml
# The config below will generate a folder structure of the following
#         [org: 123456789012]
#           [dev]    [prod]
# [retail, finance] [retail, finance]
apiVersion: blueprints.cloud.google.com/v1alpha3
kind: ResourceHierarchy
metadata:
  annotations:
    config.k8s.io/function: |
      container:
        image: gcr.io/krm-blueprints/generate-folders:dev
    config.kubernetes.io/local-config: "true"
  name: root-hierarchy
spec:
  parentRef:
    type: Organization
    external: 123456789012
  config:
    - dev:
      - retail
      - finance
    - prod:
      - retail
      - finance
`;
