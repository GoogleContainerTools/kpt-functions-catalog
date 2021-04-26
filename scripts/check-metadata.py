#!/usr/bin/env python

import os
import yaml

directories_to_scan = ['functions/go', 'functions/ts']
metadata_filename = 'metadata.yaml'
directories_to_skip = ['_template', 'dist', 'node_modules']
required_fields = ['image', 'description', 'tags', 'sourceURL', 'examplePackageURLs', 'emails', 'license']
kpt_team_email = 'kpt-team@google.com'


def validate_metadata(metadata, branch, path, fn):
    for required_field in required_fields:
        if required_field not in metadata:
            raise Exception(f'field {required_field} is required')
    desired_image_name = f'gcr.io/kpt-fn/{fn}'
    if metadata['image'] != desired_image_name:
        raise Exception(f'image name should be "{desired_image_name}"')
    if metadata['tags'] is None or len(metadata['tags']) == 0:
        raise Exception(f'"tags" must contain at least one tag')
    desired_source_url = f'https://github.com/GoogleContainerTools/kpt-functions-catalog/tree/{branch}/{path}'
    if metadata['sourceURL'] != desired_source_url:
        raise Exception(f'sourceURL should be "{desired_source_url}"')
    desired_example_pkg_url_prefix = f'https://github.com/GoogleContainerTools/kpt-functions-catalog/tree/{branch}/examples/{fn}/'
    for pkgURL in metadata['examplePackageURLs']:
        if not pkgURL.startswith(desired_example_pkg_url_prefix):
            raise Exception(f'"{pkgURL}" in examplePackageURLs should have prefix "{desired_example_pkg_url_prefix}"')
    if metadata['emails'] is None or kpt_team_email not in metadata['emails']:
        raise Exception(f'"{kpt_team_email}" should be in the emails list')


def main():
    branch_ref = os.getenv('GITHUB_REF')
    ref_prefix = 'refs/heads/'
    branch_name = 'master'
    if branch_ref is not None and branch_ref.startswith(ref_prefix):
        branch_name = branch_ref[len(ref_prefix):]

    for dir_to_scan in directories_to_scan:
        for file in os.listdir(dir_to_scan):
            path_name = os.path.join(dir_to_scan, file)
            if os.path.isdir(path_name):
                if file in directories_to_skip:
                    continue
                fn_name = file
                if metadata_filename not in os.listdir(path_name):
                    raise Exception(f'function {fn_name} directory must contain a {metadata_filename} file')
                meta = yaml.load(open(os.path.join(path_name, metadata_filename)), Loader=yaml.Loader)
                validate_metadata(meta, branch_name, path_name, fn_name)
    print("metadata files check succeeded!")


if __name__ == "__main__":
    main()
