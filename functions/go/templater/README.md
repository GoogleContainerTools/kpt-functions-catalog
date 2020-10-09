# Templater function

This is an example of implementing a templater function.
Templater or aka Go-Template Generator/Transformer is
a multitool function that allows to perform 2 things:
* transform the existing documents using KYAML Go-Template library in compbination with [Sprig library](http://masterminds.github.io/sprig/). Transformation may include deletion.
* generate documents. It's not recommended, but sometimes it still may be needed. E.g. if it's necessary to add inventory document.


The configuarion is implemented compatible with ConfigMap. There is only 1 field in ConfigMap that must be present: `entrypoint`, that contains go-template expression that will be executed.

The expression will get  2 arguments as Values:

* .Data - contains all ConfigMap `.data` literals as map, except entrypoint
* .Items - list of RNodes that the function got from the pipeline. It's possible to modify this value, e.g. using `set` funciton.

If go-template engine returns some yaml-text resources, items will be appended by that resources.

## KYAML Go-Template library

The libary is based on [kyaml module](https://github.com/kubernetes-sigs/kustomize/tree/master/kyaml) and consists of the following functions:

``` go
/* 
  Creates yaml.Filter based on the yaml configuration
  e.g.:
  example 1:
    kind: PathGetter
    path:
    - data
  example 2:
    kind: FieldSetter
    name: value
    stringValue: someValue
  see https://github.com/kubernetes-sigs/kustomize/blob/master/kyaml/yaml/filters.go#L14
  for the list of filters and see the type definition of each filter to get the names of fields.
*/
func YFilter(cfg string) yaml.Filter 
/*
  Executes the list of yfilters for the input yaml.
  Similar to function yaml.Pipe with the only difference that it swallows the error (but logs it)
*/
func YPipe(input *yaml.RNode, yfilters []yaml.Filter) *yaml.RNode
/*
  Converts RNode to go interface by
  marshalling RNode and unmarhsalling to the interface{}.
  That will create string if RNode was Scalar and should create 
  maps,arrays and etc if RNode had a complex type.
*/
func YValue(input *yaml.RNode) interface{} 
/*
Creates kio.Filter based on the yaml configuration
  e.g.:
    kind: GrepFilter
    path: 
    - metadata
    - name
    value: ^map1$
  see https://github.com/kubernetes-sigs/kustomize/blob/master/kyaml/kio/filters/filters.go#L18
  for the list of filters and see the type definition of each filter to get the names of fields.
*/
func KFilter(cfg string) kio.Filter
/* 
Creates a special type of kio.Filter that will execute the list of yaml.Filters.
It's convenient if it's necessary to modify all yamls is pipeline.
*/
func KYFilter(yfilters []yaml.Filter) kio.Filter
/*
Exectutes all filters for all input RNodes
Swallows (logs) errors
*/
func KPipe(input []*yaml.RNode, kfilters []kio.Filter) []*yaml.RNode
```
## KYAML Go-Template library usage examples
### Filtering/Deletion of some documents

Here is the example of possible configuration that will filter-out (delete) all resources that have annotation `test-annotation` = `x`

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: notImportantHere
data:
  grep: |
    kind: GrepFilter
    path:
    - metadata
    - annotations
    - test-annotation
    value: ^x$
    invertMatch: true
  entrypoint: |
    {{- $_ := (set . "Items" (KPipe .Items (list (KFilter .Data.grep)))) -}}
```

*How it works*: We're performing kio.Pipe for all input items and applies kio.filters.GrepFilter, that was initialized with configuraion from `grep` literal.
KPipe will return all Rnodes that don't have a needed annotation. The expression `$_ := (set . "Items" ...)` is needed to return this filtered list back, so the function can pass it further.

### Substitution of value in one resource with the value from another resource

Here is the example of how to:
* get value from the desired resource field
* set it to another resource field

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: notImportantHere
data:
  map1grep: |
    kind: GrepFilter
    path: 
    - metadata
    - name
    value: ^map1$
  pathGet1: |
    kind: PathGetter
    path:
    - data
    - value
  map2grep: |
    kind: GrepFilter
    path:
    - metadata
    - name
    value: ^map2$
  map2PathGet: |
    kind: PathGetter
    path:
    - data
  fieldSet: |
    kind: FieldSetter
    name: value
    stringValue: %s
  entrypoint: |
    {{- $map1 := KPipe .Items (list (KFilter .Data.map1grep)) -}}
    {{- $map1value := YValue (YPipe (index $map1 0) (list (YFilter .Data.pathGet1))) -}}
    {{- $_ := KPipe .Items (list (KFilter .Data.map2grep) (KYFilter (list (YFilter .Data.map2PathGet) (YFilter (printf .Data.fieldSet $map1value))))) -}}
```

*How it works*: the first line looks for the resource using the filter `map1grep`. Note: it's possible to set several GrepFilters in the list - it will emulate `AND` operaion.
The second line returns the value of the field `data.value` of the first resource filtered with the first string. That will be a string in case of ConfigMap.
The last string Filters all resources that match to `map2grep` and sets their field `data.value` with the value taken on the previous step. Note: if the grepFilter return severl objects - all of the will be modified.

Note: Of course it was possible to make a one-liner, but it woudn't be readable:

```
{{- $_ := KPipe .Items (list (KFilter .Data.map2grep) (KYFilter (list (YFilter .Data.map2PathGet) (YFilter (printf .Data.fieldSet YValue (YPipe (index KPipe .Items (list (KFilter .Data.map1grep)) 0) (list (YFilter .Data.pathGet1)))))))) -}}
```

### Stamping all resources with check date

Sometimes it may be necessary to set a label or annotation with the current date to some or may be all of resources. Here is the example for all documents:

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: notImportantHere
data:
  annotationSet: |
    kind: AnnotationSetter
    key: check-datestamp
    value: %s
  entrypoint: |
    {{- $_ := KPipe .Items (list (KYFilter (list (YFilter (printf .Data.annotationSet (now | date "2006-01-02")))))) -}}
```

*How it works*: KPipe executes for input resources yaml.Filter called AnnotationSetter that has been initialized with the key `check-datestamp` and value of the current date in the set format.

## Function implementation

The function is implemented as an image (see Dockerfile).

## Function invocation

The function is invoked by authoring a [local Resource](local-resource)
with `metadata.annotations.[config.kubernetes.io/function]` and running:

    kpt fn run local-resource/

This exits non-zero if there is an error.

## Running the Example

Run the function with:

    kpt fn run local-resource/

The generated resources will appear in local-resource/ and will contain the current date in check-datestamp field:

```
$ cat local-resource/data.yaml

apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
  annotations:
    check-datestamp: '2020-10-12'
data:
  value: value1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: map2
  annotations:
    check-datestamp: '2020-10-12'
data:
  value: value2
```
