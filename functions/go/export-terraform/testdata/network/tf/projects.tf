module "prj-network" {
  source  = "terraform-google-modules/project-factory/google"
  version = "~> 12.0"

  name       = "prj-network"
  org_id     = var.org_id

  billing_account = var.billing_account
}
