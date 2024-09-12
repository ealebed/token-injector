locals {
  cluster_type = "regional"

  gke_node_pools = [
    {
      name               = "app-node-pool"
      machine_type       = var.gke_node_machine_type
      auto_upgrade       = true
      auto_repair        = true
      autoscaling        = true
      min_count          = 1
      max_count          = 3
      disk_size_gb       = 100
      disk_type          = "pd-balanced"
      image_type         = "COS_CONTAINERD"
      enable_gcfs        = true
      preemtible         = true
      initial_node_count = 1
    },
  ]
}

module "gke" {
  source  = "terraform-google-modules/kubernetes-engine/google"
  version = "~> 33.0"

  project_id        = var.project_id
  name              = "${local.cluster_type}-cluster-test"
  region            = var.region
  release_channel   = var.gke_release_channel
  network           = module.network.network_name
  subnetwork        = module.network.subnets["${var.region}/subnet-gke"].name
  ip_range_pods     = module.network.subnets["${var.region}/subnet-gke"].secondary_ip_range[0].range_name
  ip_range_services = module.network.subnets["${var.region}/subnet-gke"].secondary_ip_range[1].range_name

  add_master_webhook_firewall_rules = true
  firewall_inbound_ports            = ["443", "8443", "9443", "15017"]
  remove_default_node_pool          = true
  service_account                   = "create"
  node_metadata                     = "GKE_METADATA"
  deletion_protection               = false
  grant_registry_access             = true
  node_pools                        = local.gke_node_pools
}
