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
disallowed_kpt_commands = ['kpt fn run', 'kpt cfg', 'kpt pkg cat']
gcr_prefix = 'gcr.io/kpt-fn/'
git_url_prefix = 'https://github.com/GoogleContainerTools/kpt-functions-catalog.git'

def validate_master_branch():
    fn_name_to_examples = validate_functions_dir_for_master_branch()
    validate_examples_dir_for_master_branch(fn_name_to_examples)


def validate_examples_dir_for_master_branch(fn_name_to_examples):
    examples_seen = set()
    for fn_name, examples in fn_name_to_examples.items():
        for example_name in examples:
            examples_seen.add(example_name)
            if not example_name.startswith(fn_name):
                raise Exception(f'example name {example_name} must start with the function name {fn_name}')
            validate_example_md(fn_name, examples_directory, example_name, 'master')
            validate_example_kptfile(fn_name, examples_directory, example_name, 'master')

    for dir in os.listdir(examples_directory):
        dir_name = os.path.join(examples_directory, dir)
        if os.path.isdir(dir_name):
            if dir in examples_directories_to_skip:
                continue
            if not dir in examples_seen:
                raise Exception(f'directory {dir} is NOT in the metadata.yaml file of any functions')


def validate_functions_dir_for_master_branch():
    fn_name_to_examples = {}
    for dir_to_scan in directories_to_scan:
        for file in os.listdir(dir_to_scan):
            path_name = os.path.join(dir_to_scan, file)
            if os.path.isdir(path_name):
                if file in directories_to_skip:
                    continue
                fn_name = file
                print(f'verifying {fn_name}')
                if metadata_filename not in os.listdir(path_name):
                    raise Exception(f'function {fn_name} directory must contain a {metadata_filename} file')
                meta = yaml.load(open(os.path.join(path_name, metadata_filename)), Loader=yaml.Loader)
                example_list = [eg.split('/')[-1] for eg in meta['examplePackageURLs']]
                fn_name_to_examples[fn_name] = example_list
                validate_metadata(meta, 'master', path_name, fn_name, example_list)
    return fn_name_to_examples


def validate_release_branch(branch_name):
    items = branch_name.split('/')
    if len(items) != 2:
        raise Exception(f'the release branch name must be in the format of <fn-name>/v<major>.<minor>')
    fn_name = items[0]
    examples = validate_functions_dir_for_release_branch(branch_name, fn_name)
    validate_examples_dir_for_release_branch(branch_name, fn_name, examples)

def validate_examples_dir_for_release_branch(branch_name, fn_name, examples):
    if fn_name in os.listdir(os.path.join(examples_directory, 'contrib')):
        return
    for example_name in examples:
        validate_example_md(fn_name, examples_directory, example_name, branch_name)
        validate_example_kptfile(fn_name, examples_directory, example_name, branch_name)


def validate_functions_dir_for_release_branch(branch_name, fn_name):
    fn_dir = None
    for dir_to_scan in directories_to_scan:
        if fn_name in os.listdir(dir_to_scan):
            fn_dir = dir_to_scan
    print(f'verifying {fn_name}')
    path_name = os.path.join(fn_dir, fn_name)
    if metadata_filename not in os.listdir(path_name):
        raise Exception(f'function {fn_name} directory must contain a {metadata_filename} file')
    meta = yaml.load(open(os.path.join(path_name, metadata_filename)), Loader=yaml.Loader)
    examples = [eg.split('/')[-1] for eg in meta['examplePackageURLs']]
    validate_metadata(meta, branch_name, path_name, fn_name, examples)
    return examples


def validate_example_kptfile(fn_name, dir_name, example_name, branch):
    example_path = os.path.join(dir_name, example_name)
    kptfile_path = os.path.join(example_path, 'Kptfile')
    if not os.path.exists(kptfile_path):
        return

    tag = branch
    if branch == 'master':
        tag = 'unstable'
    else:
        splits = branch.split('/')
        if len(splits) != 2:
            raise Exception(f'the release branch {branch} must has format <fn-name>/vX.Y')
        tag = splits[1]

    kptfile = yaml.load(open(kptfile_path), Loader=yaml.Loader)
    if kptfile['apiVersion'] != 'kpt.dev/v1alpha2' and kptfile['apiVersion'] != 'kpt.dev/v1':
        return
    pipeline = kptfile['pipeline']
    if 'mutators' in pipeline:
        for mutator in pipeline['mutators']:
            actual_image = mutator['image']
            desired_image_name = f'gcr.io/kpt-fn/{fn_name}:{tag}'
            if actual_image != desired_image_name:
                raise Exception(f'expect Kptfile to contain {desired_image_name} but find {actual_image}')
    if 'validators' in pipeline:
        for mutator in pipeline['validators']:
            actual_image = mutator['image']
            desired_image_name = f'gcr.io/kpt-fn/{fn_name}:{tag}'
            if actual_image != desired_image_name:
                raise Exception(f'expect Kptfile to contain {desired_image_name} but find {actual_image}')


def validate_example_md(fn_name, dir_name, example_name, branch):
    example_path = os.path.join(dir_name, example_name)
    md_file_path = os.path.join(example_path, 'README.md')

    if (fn_name + "-") not in example_name:
        raise Exception(f'example directory "{example_name}" must have the function name "{fn_name}-" as prefix')

    with open(md_file_path) as f:
        first_line = f.readline().strip()
        if not first_line.startswith('# '):
            raise Exception(f'title must be in the 1st line and starts with one "#"')
        if fn_name not in first_line:
            raise Exception(f'title "{first_line}" must be in format "<fn-name>: Example Name" and contains "{fn_name}"')
        shorter_example_name = example_name
        if example_name.startswith(fn_name):
            shorter_example_name = example_name[len(fn_name):]
        if shorter_example_name.replace('-', ' ') not in first_line.lower():
            raise Exception(f'title "{first_line}" must be in format "<fn-name>: Example Name" and contains "{shorter_example_name.replace("-", " ")}"')

    process = subprocess.Popen(['mdrip', md_file_path],
                               stdout=subprocess.PIPE,
                               stderr=subprocess.PIPE)
    stdout, stderr = process.communicate()
    if len(stderr) > 0:
        print(f'stderr of mdrip: {str(stderr)}')

    process = subprocess.Popen(['mdrip', '--label', 'skip', md_file_path],
                               stdout=subprocess.PIPE,
                               stderr=subprocess.PIPE)
    stdout2, stderr2 = process.communicate()
    if len(stderr2) > 0:
        print(f'stderr of mdrip: {str(stderr2)}')

    tag = branch
    if branch == 'master':
        tag = 'unstable'
    else:
        splits = branch.split('/')
        if len(splits) != 2:
            raise Exception(f'the release branch {branch} must has format <fn-name>/vX.Y')
        tag = splits[1]

    lines = stdout.decode("utf-8").splitlines()
    lines2 = stdout2.decode("utf-8").splitlines()

    for line in lines:
        if line.startswith('#') or line.startswith('echo'):
            continue
        for disallowed in disallowed_kpt_commands:
            if disallowed in line:
                raise Exception(f'command {disallowed} is not allowed in the desired package url in {md_file_path}')

        if line in lines2:
            continue

        for item in line.split():
            if item.startswith(git_url_prefix):
                desired_pkg_url = f'{git_url_prefix}/{example_path}'
                if branch != 'master':
                    desired_pkg_url = desired_pkg_url + f'@{branch}'
                if not item.startswith(desired_pkg_url):
                    raise Exception(f'the desired package url in {md_file_path} is {desired_pkg_url}, but found {item}')

            if gcr_prefix in item:
                desired_image_name = f'{gcr_prefix}{fn_name}:{tag}'
                if desired_image_name not in item:
                    raise Exception(f'expect "{line}" to contain "{desired_image_name}" in {md_file_path}')


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
    if len(examples_list) == 0:
        raise Exception(f'{fn}: there must be at least one example listed in metadata.yaml file')
    for example in examples_list:
        desired_example_pkg_url = f'https://github.com/GoogleContainerTools/kpt-functions-catalog/tree/{branch}/examples/{example}'
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
