module "organization-iam" {
  source  = "terraform-google-modules/iam/google//modules/organizations_iam"
  version = "~> 7.4"

  organizations = ["11111111111"]

  bindings = {
    
    "roles/editor" = [
      "group:gcp-organization-admins@example.com",
      "group:gcp-developers@example.com",
    ]
    
    "roles/orgpolicy.policyAdmin" = [
      "group:gcp-organization-admins@example.com",
    ]
    
  }
}


module "folder-1-iam" {
  source  = "terraform-google-modules/iam/google//modules/folders_iam"
  version = "~> 7.4"

  folders = ["folders/335620346181"]

  bindings = {
    
    "roles/viewer" = [
      "group:gcp-developers@example.com",
    ]
    
  }
}


module "test-iam" {
  source  = "terraform-google-modules/iam/google//modules/folders_iam"
  version = "~> 7.4"

  folders = [google_folder.test.name]

  bindings = {
    
    "roles/viewer" = [
      "group:gcp-developers@example.com",
    ]
    
  }
}
