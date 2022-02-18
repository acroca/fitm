terraform {
  required_providers {
    restapi = {
      source  = "Mastercard/restapi"
      version = "1.16.1"
    }
  }
}

provider "restapi" {
  uri = "http://127.0.0.1:4000/"
  write_returns_object = true
}

resource "restapi_object" "default_bucket" {
  path = "/buckets"
  data = jsonencode({
    id = "default"
  })
}

resource "restapi_object" "bucket1" {
  path = "/buckets"
  data = jsonencode({
    id = "bucket1"
  })
}

resource "restapi_object" "bucket2" {
  path = "/buckets"
  data = jsonencode({
    id = "bucket2"
  })
}

resource "restapi_object" "admin" {
  path = "/users"
  data = jsonencode({
    id = "admin"
    tokens = [
      "admin"
    ]
    buckets = [
      "default",
    ]
  })
}

resource "random_uuid" "user1" {
}

resource "restapi_object" "user1" {
  path = "/users"
  data = jsonencode({
    id = "user1"
    tokens = [
      random_uuid.user1.result
    ]
    buckets = [
      "default",
      "bucket2",
    ]
  })
}

output "user1_token" {
  value = random_uuid.user1.result
}
