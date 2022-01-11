resource "google_folder" "{{ . | resourceName }}" {
  display_name = "{{. | displayName}}"
  parent       = {{.Parent | reference}}
}

