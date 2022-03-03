{{range $project := .Project}}{{ if $project.ShouldCreate }}
module "{{ $project.GetResourceName }}" {
  source  = "terraform-google-modules/project-factory/google"
  version = "~> 12.0"

  name       = "{{ $project.GetDisplayName }}"{{ if ne $project.GetDisplayName $project.GetResourceName }}
  project_id = "{{ $project.GetResourceName }}"{{end}}{{if eq $project.Parent.Kind "Folder"}}
  folder_id = {{ $project.Parent.GetTerraformId }}{{end}}{{if eq $project.Parent.Kind "Organization"}}
  org_id = {{ $project.Parent.GetTerraformId }}{{end}}

  billing_account = "{{ $project.GetStringFromObject "spec" "billingAccountRef" "external" }}"{{if $project.GetBool "metadata" "annotations" "cnrm.cloud.google.com/auto-create-network"}}
  auto_create_network = true{{end}}
}
{{end}}{{end}}

