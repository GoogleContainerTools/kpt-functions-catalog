module "project-in-external" {
  source  = "terraform-google-modules/project-factory/google"
  version = "~> 12.0"

  name       = "project-in-external"
  folder_id = "folders/335620346181"

  billing_account = "AAAAAA-AAAAAA-AAAAAA"
}

module "project-in-folder" {
  source  = "terraform-google-modules/project-factory/google"
  version = "~> 12.0"

  name       = "project-name"
  project_id = "project-in-folder"
  folder_id = google_folder.test.name

  billing_account = "AAAAAA-AAAAAA-AAAAAA"
}

module "project-in-org" {
  source  = "terraform-google-modules/project-factory/google"
  version = "~> 12.0"

  name       = "project-in-org"
  org_id = "organizations/123456789012"

  billing_account = "AAAAAA-AAAAAA-AAAAAA"
  auto_create_network = true
}
