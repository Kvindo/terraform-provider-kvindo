# Look up a folder by name (or by id instead — exactly one is required).
data "kvindo_folder" "example" {
  name = "default"
}

output "folder_id" {
  value = data.kvindo_folder.example.metadata.id
}
