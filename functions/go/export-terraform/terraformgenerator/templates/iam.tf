{{range $ref := .Organization}}{{ if $ref.HasIAMBindings }}
module "{{ $ref.GetResourceName }}-iam" {
  source  = "terraform-google-modules/iam/google//modules/organizations_iam"
  version = "~> 7.4"

  organizations = ["{{ $ref.Name }}"]

  bindings = {
    {{ range $role, $binding := $ref.GetIAMBindings }}
    "{{ $role }}" = [{{ range $member := $binding.Members }}
      "{{ $member.Member }}",{{ end }}
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
    {{ range $role, $binding := $ref.GetIAMBindings }}
    "{{ $role }}" = [{{ range $member := $binding.Members }}
      "{{ $member.Member }}",{{ end }}
    ]
    {{ end }}
  }
}

{{end}}{{end}}
