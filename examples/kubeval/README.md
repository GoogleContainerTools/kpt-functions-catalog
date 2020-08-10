# Kubeval

The `kubeval` config function validates configuration using the kubeval CLI.
This example invokes the kubeval function using declarative configuration.

## Function invocation

The function is invoked using the function configuration in
`config/example.yaml`.

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/kubeval .
kpt fn run kubeval/config --network --results-dir kubeval/results
```

The first command fetches this example. The last command:

* reads configs from the `kubeval/config` folder (from invoking `kpt fn run`
  on `kubeval/config`)
* runs the function config from `kubeval/config/example.yaml` (since it
  contains the `config.kubernetes.io/function` function annotation)
* validates the remaining config from `kubeval/config`
* writes configs back into the `kubeval/config` folder and validation results
  into the `kubeval/results` folder

Check the results:

```sh
kpt cfg cat kubeval/results
```

The command results in an error
`Invalid type. Expected: [integer,null], given: string in object 'v1/ReplicationController//bob' in file example.yaml`.
Fix the error and rerun the `kpt fn run` command. This will return success (no
output).
