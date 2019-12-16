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

// Info contains versioning information. how we'll want to distribute that information.
export class Info {
  public buildDate: string;

  public compiler: string;

  public gitCommit: string;

  public gitTreeState: string;

  public gitVersion: string;

  public goVersion: string;

  public major: string;

  public minor: string;

  public platform: string;

  constructor(desc: Info) {
    this.buildDate = desc.buildDate;
    this.compiler = desc.compiler;
    this.gitCommit = desc.gitCommit;
    this.gitTreeState = desc.gitTreeState;
    this.gitVersion = desc.gitVersion;
    this.goVersion = desc.goVersion;
    this.major = desc.major;
    this.minor = desc.minor;
    this.platform = desc.platform;
  }
}
