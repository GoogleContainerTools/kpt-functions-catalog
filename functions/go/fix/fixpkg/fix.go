package fixpkg

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-openapi/spec"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/sets"
	"sigs.k8s.io/kustomize/kyaml/setters2"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/fix/v1alpha1"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/fix/v1alpha2"
)

var _ kio.Filter = &Fix{}

// Fix migrates the resources in v1alpha1 format to v1alpha2 format
type Fix struct {
	// settersSchema is the schema equivalent of openAPI section in Kptfile
	settersSchema *spec.Schema
}

// Filter implements Fix as a yaml.Filter
func (s *Fix) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	kfFound := false
	// the input nodes are sorted based on package path depth with Kptfile first in each package
	for i := range nodes {
		meta, err := nodes[i].GetMeta()
		if err != nil {
			return nodes, err
		}

		// migrate Kptfile from v1alpha1 to v1alpha2
		if meta.Kind == v1alpha2.KptFileName {
			kfFound = true
			functions := FunctionsInPkg(nodes, i+1, filepath.Dir(meta.Annotations[kioutil.PathAnnotation]))
			kNode, err := s.FixKptfile(nodes[i], functions)
			if err != nil {
				return nodes, err
			}
			nodes[i] = kNode
			continue
		}

		// fix setter comments in each resource
		err = accept(s, nodes[i], s.settersSchema)
		if err != nil {
			return nil, errors.Wrap(err)
		}
	}
	if !kfFound {
		return nodes, fmt.Errorf("Kptfile not found, make sure you specify '--include-meta-resources' flag")
	}
	return nodes, nil
}

// FixKptfile migrates the input Kptfile node from v1alpha1 to v1alpha2
func (s *Fix) FixKptfile(node *yaml.RNode, functions []v1alpha2.Function) (*yaml.RNode, error) {
	var err error
	meta, err := node.GetMeta()
	if err != nil {
		return node, err
	}
	kfOld, err := v1alpha1.ReadFile(node)
	if err != nil {
		return node, err
	}

	kfNew := v1alpha2.KptFile{ResourceMeta: meta}
	kfNew.APIVersion = v1alpha2.KptFileAPIVersion

	// convert packageMetadata in v1alpha1 Kptfile to v1alpha2 info
	if kfOld.PackageMeta != nil {
		emails := []string{kfOld.PackageMeta.Email}
		if kfOld.PackageMeta.Email == "" {
			emails = nil
		}
		kfNew.Info = &v1alpha2.PackageInfo{
			Site:        kfOld.PackageMeta.URL,
			Emails:      emails,
			License:     kfOld.PackageMeta.License,
			Description: kfOld.PackageMeta.ShortDescription,
			Keywords:    kfOld.PackageMeta.Tags,
			Man:         kfOld.PackageMeta.Man,
		}
	}

	// convert upstream section
	if kfOld.Upstream != nil {
		kfNew.Upstream = &v1alpha2.Upstream{
			Type: v1alpha2.OriginType(kfOld.Upstream.Type),
			Git: &v1alpha2.Git{
				Repo:      kfOld.Upstream.Git.Repo,
				Directory: kfOld.Upstream.Git.Directory,
				Ref:       kfOld.Upstream.Git.Ref,
			},
			UpdateStrategy: v1alpha2.ResourceMerge,
		}

		kfNew.UpstreamLock = &v1alpha2.UpstreamLock{
			Type: v1alpha2.OriginType(kfOld.Upstream.Type),
			Git: &v1alpha2.GitLock{
				Repo:      kfOld.Upstream.Git.Repo,
				Directory: kfOld.Upstream.Git.Directory,
				Ref:       kfOld.Upstream.Git.Ref,
			},
		}
	}

	// convert OpenAPI section in v1alpha1 Kptfile to apply-setters
	s.settersSchema, err = schemaUsingField(node, openapi.SupplementaryOpenAPIFieldName)
	if err != nil {
		return node, err
	}
	setters, err := listSetters(node)
	if err != nil {
		return node, err
	}

	pl := &v1alpha2.Pipeline{}
	pl.Mutators = append(pl.Mutators, functions...)
	kfNew.Pipeline = pl

	if setters != nil && len(setters) > 0 {
		fn := v1alpha2.Function{
			Image:     "gcr.io/kpt-fn/apply-setters:v0.1",
			ConfigMap: setters,
		}
		pl.Mutators = append(pl.Mutators, fn)
	}

	// convert inventory section
	if kfOld.Inventory != nil {
		kfNew.Inventory = &v1alpha2.Inventory{
			Namespace:   kfOld.Inventory.Namespace,
			Name:        kfOld.Inventory.Name,
			InventoryID: kfOld.Inventory.InventoryID,
			Labels:      kfOld.Inventory.Labels,
			Annotations: kfOld.Inventory.Annotations,
		}
	}

	// convert kfNew to yaml node
	b, err := yaml.Marshal(kfNew)
	if err != nil {
		return node, err
	}
	kNode, err := yaml.Parse(string(b))
	if err != nil {
		return node, err
	}
	err = kNode.SetAnnotations(meta.Annotations)
	return kNode, err
}

// FunctionsInPkg gets the v1alpha2 functions list for functions in package
// nodes is list of input nodes which are sorted according to the package depth
// i is the index of the Kptfile of the package, all the files till i hits next Kptfile
// are the files of the package
// pkgPath is the package path relative to the root package directory
func FunctionsInPkg(nodes []*yaml.RNode, i int, pkgPath string) []v1alpha2.Function {
	var res []v1alpha2.Function
	for i < len(nodes) {
		node := nodes[i]
		meta, err := node.GetMeta()
		if err != nil {
			return res
		}
		if meta.Kind == v1alpha2.KptFileName {
			// we hit the Kptfile of next package, so break here
			break
		}
		fnSpec := runtimeutil.GetFunctionSpec(node)
		if fnSpec != nil {
			// in v1alpha2, fn-config must be present in the package directory
			// so configPath must be just the file name
			fnFileName := filepath.Base(meta.Annotations[kioutil.PathAnnotation])
			res = append(res, v1alpha2.Function{
				Image:      fnSpec.Container.Image,
				ConfigPath: fnFileName,
			})
			// move the fn-config to the top level directory of the package
			meta.Annotations[kioutil.PathAnnotation] = filepath.Join(pkgPath, fnFileName)
			delete(meta.Annotations, runtimeutil.FunctionAnnotationKey)
			delete(meta.Annotations, "config.k8s.io/function")
			node.SetAnnotations(meta.Annotations)
		}
		i++
	}
	return res
}

// visitMapping visits mapping node to convert the comments for a array setters
func (s *Fix) visitMapping(object *yaml.RNode) error {
	return object.VisitFields(func(node *yaml.MapNode) error {
		if node.IsNilOrEmpty() {
			return nil
		}
		if node.Value.YNode().Kind != yaml.SequenceNode {
			// return if it is not a sequence node
			return nil
		}

		comment := node.Key.YNode().LineComment
		// # {"$kpt-set":"foo"} must be converted to # kpt-set: ${foo}
		if strings.Contains(comment, "$kpt-set") {
			comment := strings.TrimPrefix(comment, `# {"$kpt-set":"`)
			comment = strings.TrimSuffix(comment, `"}`)
			node.Key.YNode().LineComment = fmt.Sprintf("kpt-set: ${%s}", comment)
		}
		return nil
	})
}

// visitScalar visits scalar nodes and converts the comments to v1alpha2 format
func (s *Fix) visitScalar(object *yaml.RNode, setterSchema *openapi.ResourceSchema) error {
	ext, err := getExtFromComment(setterSchema)
	if err != nil {
		return err
	}
	if ext == nil {
		return nil
	}

	ok, err := s.fixSetter(object, ext)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	_, err = s.fixSubst(object, ext)
	if err != nil {
		return err
	}
	return nil
}

// fixSetter converts the setter comment
func (s *Fix) fixSetter(field *yaml.RNode, ext *setters2.CliExtension) (bool, error) {
	// check full setter
	if ext == nil || ext.Setter == nil {
		return false, nil
	}

	field.YNode().LineComment = fmt.Sprintf("kpt-set: ${%s}", ext.Setter.Name)
	return true, nil
}

// fixSubst converts the substitution comment to expanded setter comment pattern
func (s *Fix) fixSubst(field *yaml.RNode, ext *setters2.CliExtension) (bool, error) {
	if ext.Substitution == nil {
		return false, nil
	}

	// track the visited nodes to detect cycles in nested substitutions
	visited := sets.String{}

	res, err := s.substituteUtil(ext, visited)
	if err != nil {
		return false, err
	}

	field.YNode().LineComment = fmt.Sprintf("kpt-set: %s", res)

	return true, nil
}

// substituteUtil recursively parses nested substitutions in ext and sets the setter value
// returns error if cyclic substitution is detected or any other unexpected errors
func (s *Fix) substituteUtil(ext *setters2.CliExtension, visited sets.String) (string, error) {
	// check if the substitution has already been visited and throw error as cycles
	// are not allowed in nested substitutions
	if visited.Has(ext.Substitution.Name) {
		return "", errors.Errorf(
			"cyclic substitution detected with name " + ext.Substitution.Name)
	}

	visited.Insert(ext.Substitution.Name)
	pattern := ext.Substitution.Pattern

	// substitute each setter into the pattern to get the new value
	// if substitution references to another substitution, recursively
	// process the nested substitutions to replace the pattern with setter values
	for _, v := range ext.Substitution.Values {
		if v.Ref == "" {
			return "", errors.Errorf(
				"missing reference on substitution " + ext.Substitution.Name)
		}
		ref, err := spec.NewRef(v.Ref)
		if err != nil {
			return "", errors.Wrap(err)
		}
		def, err := openapi.Resolve(&ref, s.settersSchema) // resolve the def to its openAPI def
		if err != nil {
			return "", errors.Wrap(err)
		}
		defExt, err := setters2.GetExtFromSchema(def) // parse the extension out of the openAPI
		if err != nil {
			return "", errors.Wrap(err)
		}

		if defExt.Substitution != nil {
			// parse recursively if it reference is substitution
			substVal, err := s.substituteUtil(defExt, visited)
			if err != nil {
				return "", err
			}
			pattern = strings.ReplaceAll(pattern, v.Marker, substVal)
			continue
		}

		if val, found := defExt.Setter.EnumValues[defExt.Setter.Value]; found {
			// the setter has an enum-map. we should replace the marker with the
			// enum value looked up from the map rather than the enum key
			pattern = strings.ReplaceAll(pattern, v.Marker, val)
		} else {
			pattern = strings.ReplaceAll(pattern, v.Marker, fmt.Sprintf("${%s}", defExt.Setter.Name))
		}
	}
	return pattern, nil
}

// listSetters extracts the setters information from the input Kptfile node
func listSetters(object *yaml.RNode) (map[string]string, error) {
	setters := make(map[string]string)
	// read the OpenAPI definitions in Kptfile
	def, err := object.Pipe(yaml.LookupCreate(yaml.MappingNode, "openAPI", "definitions"))
	if err != nil {
		return nil, err
	}
	if yaml.IsMissingOrNull(def) {
		return nil, nil
	}

	// iterate over definitions -- find those that are setters
	err = def.VisitFields(func(node *yaml.MapNode) error {
		setter := setters2.SetterDefinition{}

		// the definition key -- contains the setter name
		key := node.Key.YNode().Value

		if !strings.HasPrefix(key, fieldmeta.SetterDefinitionPrefix) {
			// not a setter -- doesn't have the right prefix
			return nil
		}

		setterNode, err := node.Value.Pipe(yaml.Lookup(setters2.K8sCliExtensionKey, "setter"))
		if err != nil {
			return err
		}
		if yaml.IsMissingOrNull(setterNode) {
			// has the setter prefix, but missing the setter extension
			return errors.Errorf("missing x-k8s-cli.setter for %s", key)
		}

		// unmarshal the yaml for the setter extension into the definition struct
		b, err := setterNode.String()
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal([]byte(b), &setter); err != nil {
			return err
		}

		// the description is not part of the extension, and should be pulled out
		// separately from the extension values.
		description := node.Value.Field("description")
		if description != nil {
			setter.Description = description.Value.YNode().Value
		}
		if len(setter.ListValues) > 0 {
			var vals string
			for _, val := range setter.ListValues {
				vals = vals + ", " + val
			}
			vals = strings.TrimPrefix(vals, ", ")
			setters[setter.Name] = fmt.Sprintf("[%s]", vals)
		} else {
			setters[setter.Name] = setter.Value
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return setters, nil
}

// getExtFromComment returns the cliExtension openAPI extension if it is present as
// a comment on the field.
func getExtFromComment(schema *openapi.ResourceSchema) (*setters2.CliExtension, error) {
	if schema == nil {
		// no schema found
		// TODO(pwittrock): should this be an error if it doesn't resolve?
		return nil, nil
	}

	// get the cli extension from the openapi (contains setter information)
	ext, err := setters2.GetExtFromSchema(schema.Schema)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return ext, nil
}

// schemaUsingField returns the schema object for the openAPI section in input Kptfile object node
func schemaUsingField(object *yaml.RNode, field string) (*spec.Schema, error) {
	if field != "" {
		// get the field containing the openAPI
		m := object.Field(field)
		if m.IsNilOrEmpty() {
			// doesn't contain openAPI definitions
			return nil, nil
		}
		object = m.Value
	}

	oAPI, err := object.String()
	if err != nil {
		return nil, err
	}

	// convert the yaml openAPI to a JSON string by unmarshalling it to an
	// interface{} and the marshalling it to a string
	var o interface{}
	err = yaml.Unmarshal([]byte(oAPI), &o)
	if err != nil {
		return nil, err
	}
	j, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	var sc spec.Schema
	err = sc.UnmarshalJSON(j)
	if err != nil {
		return nil, err
	}

	return &sc, nil
}