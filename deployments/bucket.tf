resource "google_storage_bucket_iam_member" "viewer" {
  bucket = google_storage_bucket.bucket.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

resource "google_storage_bucket" "bucket" {
  name                        = var.gcp_bucket_name
  uniform_bucket_level_access = true
  cors {
    origin          = ["*"]
    method          = ["GET", "HEAD"]
    response_header = ["*"]
    max_age_seconds = 3600
  }
}

# In an ideal world we could just reference the swagger file in gitlab. 
# Example: https://gitlab.com/Blockdaemon/bpm-sdk/-/raw/v0.14.1/swagger.yaml
# Unfortunately gitlab doesn't allow CORS so as a workaround we put the
# file in a Google Bucket
resource "google_storage_bucket_object" "swagger" {
  name   = "swagger.yaml"
  source = "../swagger.yaml"
  bucket = google_storage_bucket.bucket.name
}

resource "google_storage_bucket_object" "image1" {
  name   = "docs/install.png"
  source = "../docs/install.png"
  bucket = google_storage_bucket.bucket.name
}

resource "google_storage_bucket_object" "image2" {
  name   = "docs/start.png"
  source = "../docs/start.png"
  bucket = google_storage_bucket.bucket.name
}

resource "google_storage_bucket_object" "image3" {
  name   = "docs/configure.png"
  source = "../docs/configure.png"
  bucket = google_storage_bucket.bucket.name
}
