/**
 * Copyright 2019 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { KubernetesObject } from '@googlecontainertools/kpt-functions';
import * as apisMetaV1 from './io.k8s.apimachinery.pkg.apis.meta.v1';

export class Team implements KubernetesObject {
  // APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources
  public apiVersion: string;

  // Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
  public kind: string;

  // Standard object's metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
  public metadata: apisMetaV1.ObjectMeta;

  public spec: Team.Spec;

  constructor(desc: Team.Interface) {
    this.apiVersion = Team.apiVersion;
    this.kind = Team.kind;
    this.metadata = desc.metadata;
    this.spec = desc.spec;
  }
}

export function isTeam(o: any): o is Team {
  return o && o.apiVersion === Team.apiVersion && o.kind === Team.kind;
}

export namespace Team {
  export const apiVersion = 'anthos.cft.dev/v1alpha1';
  export const group = 'anthos.cft.dev';
  export const version = 'v1alpha1';
  export const kind = 'Team';

  export interface Interface {
    // Standard object's metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
    metadata: apisMetaV1.ObjectMeta;

    spec: Team.Spec;
  }
  export class Spec {
    public roles?: Team.Spec.Item[];
  }

  export namespace Spec {
    export class Item {
      public groups?: string[];

      public role: string;

      public users?: string[];

      constructor(desc: Team.Spec.Item) {
        this.groups = desc.groups;
        this.role = desc.role;
        this.users = desc.users;
      }
    }
  }
}

// TeamList is a list of Team
export class TeamList {
  // APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources
  public apiVersion: string;

  // List of teams. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md
  public items: Team[];

  // Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
  public kind: string;

  // ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}.
  public metadata?: TeamList.Metadata;

  constructor(desc: TeamList) {
    this.apiVersion = TeamList.apiVersion;
    this.items = desc.items.map(i => new Team(i));
    this.kind = TeamList.kind;
    this.metadata = desc.metadata;
  }
}

export function isTeamList(o: any): o is TeamList {
  return o && o.apiVersion === TeamList.apiVersion && o.kind === TeamList.kind;
}

export namespace TeamList {
  export const apiVersion = 'anthos.cft.dev/v1alpha1';
  export const group = 'anthos.cft.dev';
  export const version = 'v1alpha1';
  export const kind = 'TeamList';

  // TeamList is a list of Team
  export interface Interface {
    // List of teams. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md
    items: Team[];

    // ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}.
    metadata?: TeamList.Metadata;
  }
  // ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}.
  export class Metadata {
    // continue may be set if the user set a limit on the number of items returned, and indicates that the server has more data available. The value is opaque and may be used to issue another request to the endpoint that served this list to retrieve the next set of available objects. Continuing a consistent list may not be possible if the server configuration has changed or more than a few minutes have passed. The resourceVersion field returned when using this continue value will be identical to the value in the first response, unless you have received this token from an error message.
    public continue?: string;

    // String that identifies the server's internal version of this object that can be used by clients to determine when objects have changed. Value must be treated as opaque by clients and passed unmodified back to the server. Populated by the system. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency
    public resourceVersion?: string;

    // selfLink is a URL representing this object. Populated by the system. Read-only.
    public selfLink?: string;
  }
}
