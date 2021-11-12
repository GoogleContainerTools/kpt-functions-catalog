package docs

import (
	"bytes"
	"fmt"
	"html/template"
	"sort"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/list-setters/listsetters"
	kptfilev1 "github.com/GoogleContainerTools/kpt-functions-sdk/go/pkg/api/kptfile/v1"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// GenerateBlueprintReadme generates markdown readme for a blueprint
func GenerateBlueprintReadme(nodes []*yaml.RNode, repoPath string) (string, error) {
	r, err := newBlueprintReadme(nodes, repoPath)
	if err != nil {
		return "", err
	}
	// individual sections in the readme
	blueprintSections := []generateSection{
		generateHeaderSection,
		generateSetterTableSection,
		generateSubPkgSection,
		generateResourceTableSection,
		generateResourceRefsSection,
		generateUsageSection,
	}
	for _, section := range blueprintSections {
		r.addSection(section)
	}
	// render readme
	err = r.render()
	if err != nil {
		return "", err
	}
	return r.string(), nil

}

// generateHeaderSection generates the title header and description
func generateHeaderSection(r *blueprintReadme) error {
	r.write(getMdHeading(getBlueprintTitle(r.bp.rootKf), 1))
	// literal style description will include a newline
	if strings.HasSuffix(r.bp.rootKf.Info.Description, "\n") {
		r.write(r.bp.rootKf.Info.Description)
	} else {
		r.writeLn(r.bp.rootKf.Info.Description)
	}
	return nil
}

// generateSetterTableSection generates a markdown table of setters used in the package
func generateSetterTableSection(r *blueprintReadme) error {
	ls := listsetters.New()
	_, err := ls.Filter(r.bp.nodes)
	if err != nil {
		return err
	}
	setters := ls.GetResults()

	buf := &bytes.Buffer{}
	table := newMarkdownTable([]string{"Name", "Value", "Type", "Count"}, buf)
	for _, setter := range setters {
		table.Append([]string{setter.Name, setter.Value, setter.Type, fmt.Sprintf("%d", setter.Count)})
	}
	r.write(getMdHeading("Setters", 2))
	table.Render()
	r.write(buf.String())
	return nil
}

// generateResourceTableSection generates subpkg section with links to subpkgs if any
func generateSubPkgSection(r *blueprintReadme) error {
	subPkgLinks := []string{}
	for path, pkg := range r.bp.kfs {
		// ignore rootpkg
		if path == kptfilev1.KptFileName {
			continue
		}
		// path of the form subpkg/Kptfile, we only require subpkg/
		pkgPath := strings.TrimSuffix(path, fmt.Sprintf("/%s", kptfilev1.KptFileName))
		subPkgLinks = append(subPkgLinks, getMdLink(pkg.Name, pkgPath))
	}

	r.write(getMdHeading("Subpackages", 2))
	if len(subPkgLinks) == 0 {
		r.writeLn("This package has no sub-packages.")
	} else {
		sort.Strings(subPkgLinks)
		for _, link := range subPkgLinks {
			r.writeLn(getMdListItem(link))
		}
	}
	return nil
}

// generateResourceTableSection generates a markdown table of resources in the package
func generateResourceTableSection(r *blueprintReadme) error {
	buf := &bytes.Buffer{}
	table := newMarkdownTable([]string{"File", "APIVersion", "Kind", "Name", "Namespace"}, buf)
	for _, r := range r.filteredNodes {
		path, err := findResourcePath(r)
		if err != nil {
			return err
		}
		table.Append([]string{path, r.GetApiVersion(), r.GetKind(), r.GetName(), r.GetNamespace()})
	}
	r.write(getMdHeading("Resources", 2))
	table.Render()
	r.write(buf.String())
	return nil
}

// generateResourceRefsSection generates resource references with links to c.g.c or k8s docs
func generateResourceRefsSection(r *blueprintReadme) error {
	resourcesLinks := []string{}
	gvkSet := map[resid.Gvk]bool{}

	// dedupe multiple resources of same gvk
	for _, item := range r.filteredNodes {
		r := resid.GvkFromNode(item)
		_, exists := gvkSet[r]
		if !exists {
			gvkSet[r] = true
		}
	}

	// generate links for each gvk, if no link document kind
	for r := range gvkSet {
		link := getResourceDocsLink(r)
		if link != "" {
			resourcesLinks = append(resourcesLinks, link)
		} else {
			resourcesLinks = append(resourcesLinks, r.Kind)
		}

	}

	r.write(getMdHeading("Resource References", 2))
	if len(resourcesLinks) == 0 {
		r.writeLn("This package has no resources.")
	} else {
		sort.Strings(resourcesLinks)
		for _, l := range resourcesLinks {
			r.writeLn(getMdListItem(l))
		}
	}
	return nil

}

// generateUsageSection generates usage section describing how to use the package
func generateUsageSection(r *blueprintReadme) error {
	tmpl := strings.NewReplacer("¬", "`").Replace(`1.  Clone the package:
    ¬¬¬
    kpt pkg get {{.RepoPath}}{{.Pkgname}}@${VERSION}
    ¬¬¬
    Replace ¬${VERSION}¬ with the desired repo branch or tag
    (for example, ¬main¬).

1.  Move into the local package:
    ¬¬¬
    cd "./{{.Pkgname}}/"
    ¬¬¬

1.  Edit the function config file(s):
    - setters.yaml

1.  Execute the function pipeline
    ¬¬¬
    kpt fn render
    ¬¬¬

1.  Initialize the resource inventory
    ¬¬¬
    kpt live init --namespace ${NAMESPACE}"
    ¬¬¬
    Replace ¬${NAMESPACE}¬ with the namespace in which to manage
    the inventory ResourceGroup (for example, ¬config-control¬).

1.  Apply the package resources to your cluster
    ¬¬¬
    kpt live apply
    ¬¬¬

1.  Wait for the resources to be ready
    ¬¬¬
    kpt live status --output table --poll-until current
    ¬¬¬`)
	t, err := template.New("usage").Parse(tmpl)
	if err != nil {
		return err
	}
	r.write(getMdHeading("Usage", 2))
	err = t.Execute(r.content, struct {
		Pkgname  string
		RepoPath string
	}{
		Pkgname:  r.bp.rootKf.Name,
		RepoPath: r.bp.repoPath,
	},
	)
	if err != nil {
		return err
	}
	return nil
}
