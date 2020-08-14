# Kubeval

The `kubeval` KRM config function validates configuration using kubeval. This
example invokes the kubeval function using declarative configuration.

## Function invocation

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/kubeval .
kpt fn run kubeval --network --results-dir /tmp
```

## Expected Results

The `--results-dir` flag let us specify a destination to write function results
to. Check the results:

```sh
cat /tmp/results-0.yaml
```

They contain the following error:

```sh
Invalid type. Expected: [integer,null], given: string
```

The error comes from the `bob` resource in `configs/example-config.yaml`.
Replace the value of `spec.replicas` with an integer to pass validation and
rerun the command. This will return success (no output).
