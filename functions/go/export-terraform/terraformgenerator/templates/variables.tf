{{range $variable := .Variables}}
variable "org_id" {
  description = "{{ $variable.Description }}"
  type        = string
  default     = "{{ $variable.Default }}"
}
{{end}}
