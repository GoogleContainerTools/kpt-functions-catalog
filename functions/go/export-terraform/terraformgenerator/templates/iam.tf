{{range $org := .Organization}}{{ if $org.HasIAMBindings }}
module "{{ $org.GetResourceName }}-iam" {
  source  = "terraform-google-modules/iam/google//modules/organizations_iam"
  version = "~> 7.4"

  organizations = ["{{ $org.Name }}"]

  bindings = {
    {{ range $role, $bindings := $org.GetIAMBindings }}
    "{{ $role }}" = [{{ range $binding := $bindings }}
      "{{ $binding.Member }}",{{ end }}
    ]
    {{ end }}
  }
}

{{end}}{{end}}{{range $ref := .Folder}}{{ if $ref.HasIAMBindings }}
module "{{ $ref.GetResourceName }}-iam" {
  source  = "terraform-google-modules/iam/google//modules/folders_iam"
  version = "~> 7.4"

  folders = [{{ $ref.GetTerraformId }}]

  bindings = {
    {{ range $role, $bindings := $ref.GetIAMBindings }}
    "{{ $role }}" = [{{ range $binding := $bindings }}
      "{{ $binding.Member }}",{{ end }}
    ]
    {{ end }}
  }
}

{{end}}{{end}}
