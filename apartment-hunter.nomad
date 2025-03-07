job "projects.apartment-hunter" {
  datacenters = ["dc1"]
  type        = "service"
  constraint {
    attribute = "${meta.role}"
    operator  = "="
    value     = "public_cloud_node"
  }

  group "apartment-hunter" {
    count = 1
    update {
      canary       = 1
      max_parallel = 1
      auto_revert  = true
      auto_promote = true
    }
    task "hunter" {
      driver = "docker"
      config {
        image = "d.isotronic.de/project/apartmenthunter:latest"
        force_pull = true
      }
      restart {
        attempts = 5
        interval = "10m"
        delay    = "30s"
        mode     = "fail"
      }
      resources {
        cpu    = 100
        memory = 100
      }
    }
  }
}