locals {
  subnets = {
    gke = {
      subnet_ip     = "10.0.0.0/20"
      subnet_name   = "subnet-gke"
      subnet_region = var.region
      description   = "GKE Cluster Subnet"
      secondary_ranges = [
        {
          range_name    = "subnet-pods"
          ip_cidr_range = "10.1.0.0/20"
        },
        {
          range_name    = "subnet-svc"
          ip_cidr_range = "10.2.0.0/24"
        }
      ]
    }
  }
}

module "network" {
  source  = "terraform-google-modules/network/google"
  version = "~> 9.0"

  project_id   = var.project_id
  network_name = "test-vpc-network"
  routing_mode = "GLOBAL"

  subnets = [
    {
      subnet_name           = local.subnets.gke.subnet_name
      subnet_ip             = local.subnets.gke.subnet_ip
      subnet_region         = local.subnets.gke.subnet_region
      subnet_private_access = "true"
      description           = local.subnets.gke.description
    },
  ]

  secondary_ranges = {
    (local.subnets.gke.subnet_name) = local.subnets.gke.secondary_ranges
  }
}
