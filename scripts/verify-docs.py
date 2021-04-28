#!/usr/bin/env python

# Copyright 2021 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script does the following things:
# - Enforce several rules (e.g. required fields) in metadata.yaml for each curated function.
# - Ensure the URLs in the examples are correct.


import os
import yaml
import subprocess

metadata_filename = 'metadata.yaml'
directories_to_skip = ['_template', 'dist', 'node_modules']
examples_directory = 'examples'
functions_directory = 'functions'
directories_to_scan = [os.path.join(functions_directory, 'go'), os.path.join(functions_directory, 'ts')]
examples_directories_to_skip = ['_template', 'contrib']
required_fields = ['image', 'description', 'tags', 'sourceURL', 'examplePackageURLs', 'emails', 'license']
kpt_team_email = 'kpt-team@google.com'


def validate_master_branch():
    fn_name_to_examples = validate_examples_dir_for_master_branch()
    validate_functions_dir_for_master_branch(fn_name_to_examples)


def validate_examples_dir_for_master_branch():
    fn_name_to_examples = {}
    for dir in os.listdir(examples_directory):
        dir_name = os.path.join(examples_directory, dir)
        if os.path.isdir(dir_name):
            if dir in examples_directories_to_skip:
                continue
            fn_name = dir
            example_list = []
            for example_name in os.listdir(dir_name):
                example_list.append(example_name)
                validate_example_md(os.path.join(dir_name, example_name), 'master')
            fn_name_to_examples[fn_name] = example_list
    return fn_name_to_examples


def validate_functions_dir_for_master_branch(fn_name_to_examples):
    for dir_to_scan in directories_to_scan:
        for file in os.listdir(dir_to_scan):
            path_name = os.path.join(dir_to_scan, file)
            if os.path.isdir(path_name):
                if file in directories_to_skip:
                    continue
                fn_name = file
                print(f'verifying {fn_name}')
                if fn_name not in fn_name_to_examples:
                    raise Exception(f'function {fn_name} must have at least one example in the examples/ directory')
                if metadata_filename not in os.listdir(path_name):
                    raise Exception(f'function {fn_name} directory must contain a {metadata_filename} file')
                meta = yaml.load(open(os.path.join(path_name, metadata_filename)), Loader=yaml.Loader)
                validate_metadata(meta, 'master', path_name, fn_name, fn_name_to_examples[fn_name])


def validate_release_branch(branch_name):
    items = branch_name.split('/')
    if len(items) != 2:
        raise Exception(f'the release branch name must be in the format of <fn-name>/v<major>.<minor>')
    fn_name = items[0]
    examples = validate_examples_dir_for_release_branch(branch_name, fn_name)
    validate_functions_dir_for_release_branch(examples, branch_name, fn_name)


def validate_examples_dir_for_release_branch(branch_name, fn_name):
    if fn_name in os.listdir(os.path.join(examples_directory, 'contrib')):
        return None
    if fn_name not in os.listdir(examples_directory):
        raise Exception(f'{fn_name} must have at least one example in the examples/ directory')
    examples = []
    dir_name = os.path.join(examples_directory, fn_name)
    for example_name in os.listdir(dir_name):
        examples.append(example_name)
        validate_example_md(os.path.join(dir_name, example_name), branch_name)
    return examples


def validate_functions_dir_for_release_branch(examples, branch_name, fn_name):
    fn_dir = None
    for dir_to_scan in directories_to_scan:
        if fn_name in os.listdir(dir_to_scan):
            fn_dir = dir_to_scan
    print(f'verifying {fn_name}')
    path_name = os.path.join(fn_dir, fn_name)
    if metadata_filename not in os.listdir(path_name):
        raise Exception(f'function {fn_name} directory must contain a {metadata_filename} file')
    meta = yaml.load(open(os.path.join(path_name, metadata_filename)), Loader=yaml.Loader)
    validate_metadata(meta, branch_name, path_name, fn_name, examples)


def validate_example_md(example_path, branch):
    md_file_path = os.path.join(example_path, 'README.md')
    process = subprocess.Popen(['mdrip', '--label', 'test', md_file_path],
                               stdout=subprocess.PIPE,
                               stderr=subprocess.PIPE)
    stdout, stderr = process.communicate()
    if len(stderr) > 0:
        print(f'stderr of mdrip: {str(stderr)}')
    found_pkg_url = False
    for line in str(stdout).splitlines():
        if line.startswith('#') or line.startswith('echo'):
            continue
        for item in line.split():
            git_url_prefix = 'https://github.com/GoogleContainerTools/kpt-functions-catalog.git'
            if item.startswith(git_url_prefix):
                found_pkg_url = True
                desired_pkg_url = f'{git_url_prefix}/{example_path}'
                if branch != 'master':
                    desired_pkg_url = desired_pkg_url + f'@{branch}'
                if item != desired_pkg_url:
                    raise Exception(f'the desired package url in {md_file_path} is {desired_pkg_url}, but found {item}')
    if not found_pkg_url:
        raise Exception(f'at least one fenced code block should be marked with "<!-- @yourComment @test -->" in {md_file_path}')


def validate_metadata(metadata, branch, path, fn, examples_list):
    for required_field in required_fields:
        if required_field not in metadata:
            raise Exception(f'{fn}: field {required_field} is required')
    desired_image_name = f'gcr.io/kpt-fn/{fn}'
    if metadata['image'] != desired_image_name:
        raise Exception(f'{fn}: image name should be "{desired_image_name}"')
    if metadata['tags'] is None or len(metadata['tags']) == 0:
        raise Exception(f'{fn}: "tags" must contain at least one tag')
    desired_source_url = f'https://github.com/GoogleContainerTools/kpt-functions-catalog/tree/{branch}/{path}'
    if metadata['sourceURL'] != desired_source_url:
        raise Exception(f'{fn}: "sourceURL" should be "{desired_source_url}"')
    if len(metadata['examplePackageURLs']) != len(examples_list):
        raise Exception(f"{fn}: examplePackageURLs have {len(metadata['examplePackageURLs'])} examples, but the examples directory has {len(examples_list)}")
    for example in examples_list:
        desired_example_pkg_url = f'https://github.com/GoogleContainerTools/kpt-functions-catalog/tree/{branch}/examples/{fn}/{example}'
        if desired_example_pkg_url not in metadata['examplePackageURLs']:
            raise Exception(f'"{desired_example_pkg_url}" is not listed in examplePackageURLs')
    if metadata['emails'] is None or kpt_team_email not in metadata['emails']:
        raise Exception(f'"{kpt_team_email}" should be in the emails list')


def main():
    branch_name = os.getenv('GITHUB_BASE_REF')
    if branch_name is None or len(branch_name) == 0:
        branch_name = 'master'

    if branch_name == 'master':
        validate_master_branch()
    else:
        validate_release_branch(branch_name)
    print("Docs verification succeeded!")


if __name__ == "__main__":
    main()
