resource "pinecone_index" "test" {
  name      = "test"
  dimension = 1536
  metric    = "dotproduct"
}
