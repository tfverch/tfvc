module "cluster" {
  count = local.enable_k8s_cluster ? 1 : 0

  source = "git::ssh://git@gitlab.developers.cam.ac.uk/uis/devops/infra/terraform/gke-cluster.git?ref=2.0.0"

  project  = module.project.project_id
  location = local.gke_region

  enable_autopilot = !local.is_production

  machine_type = local.gke_node_size

  autoscaling = local.gke_node_autoscaling

  enable_workload_identity = true
}
