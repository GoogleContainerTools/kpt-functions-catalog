# Helm Template

This is an example of implementing a helm template function using declarative configuration.

This example uses the simplest approach for building abstractions.

## Function invocation

The function is invoked by authoring a [local Resource](local-resource)
with `metadata.annotations.[config.kubernetes.io/function]` and running:

    kpt fn run local-resource/

This should expand the templates and output configs but runs into an error sourcing yaml.
