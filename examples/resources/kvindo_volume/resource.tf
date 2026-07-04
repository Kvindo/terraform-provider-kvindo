resource "kvindo_volume" "example" {
  metadata = {
    name = "my-volume"
  }
  spec = {
    hosting_provider_id = "01abc123def456gh789012345"
    offer_id            = "01vol0ffr12345678901234"
    size_gib            = 50
  }
}

# Back up the volume to an image
resource "kvindo_image" "volume_backup" {
  metadata = {
    name = "my-volume-backup"
  }
  spec = {
    volume_id = kvindo_volume.example.id
  }
}

# Restore into a new volume from that backup
resource "kvindo_volume" "restored" {
  metadata = {
    name = "my-volume-restored"
  }
  spec = {
    hosting_provider_id = "01abc123def456gh789012345"
    offer_id            = "01vol0ffr12345678901234"
    image_id            = kvindo_image.volume_backup.id
  }
}
