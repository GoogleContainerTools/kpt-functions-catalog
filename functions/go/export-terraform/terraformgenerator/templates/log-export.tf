{{range $logsink := .LoggingLogSink}}{{ if $logsink.ShouldCreate }}
module "logsink-{{ $logsink.GetResourceName }}" {
  source  = "terraform-google-modules/log-export/google"
  version = "~> 7.3.0"

  destination_uri      = module.{{with or $logsink.References.BigQueryDataset $logsink.References.PubSubTopic $logsink.References.StorageBucket }}{{.GetResourceName}}{{end}}-destination.destination_uri
  log_sink_name        = "{{ $logsink.GetDisplayName }}"
  parent_resource_id   = {{ $logsink.GetOrganization.GetTerraformId false }}
  parent_resource_type = "organization"
  include_children     = true
}
{{end}}{{with $logsink.References.BigQueryDataset }}
module "{{ .GetResourceName }}-destination" {
  source  = "terraform-google-modules/log-export/google//modules/bigquery"
  version = "~> 7.3.0"

  project_id               = module.{{ .Parent.GetResourceName }}.project_id
  dataset_name             = "{{ .GetResourceName }}"
  log_sink_writer_identity = module.logsink-{{ $logsink.GetResourceName }}.writer_identity
  expiration_days          = "365"
  location                 = "US"
}
{{end}}{{with $logsink.References.PubSubTopic }}
module "{{ .GetResourceName }}-destination" {
  source  = "terraform-google-modules/log-export/google//modules/pubsub"
  version = "~> 7.3.0"

  project_id               = module.{{ .Parent.GetResourceName }}.project_id
  topic_name               = "{{ .GetResourceName }}"
  log_sink_writer_identity = module.logsink-{{ $logsink.GetResourceName }}.writer_identity
}
{{end}}{{with $logsink.References.StorageBucket }}
module "{{ .GetResourceName }}-destination" {
  source  = "terraform-google-modules/log-export/google//modules/storage"
  version = "~> 7.3.0"

  project_id                  = module.{{ .Parent.GetResourceName }}.project_id
  storage_bucket_name         = "{{ .GetResourceName }}"
  log_sink_writer_identity    = module.logsink-{{ $logsink.GetResourceName }}.writer_identity
  uniform_bucket_level_access = true
  location                    = "US"
  storage_class               = "MULTI_REGIONAL"
  retention_policy            = { retention_period_days = 365, is_locked = false }
}
{{end}}{{end}}
