---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "pinecone_index Resource - pinecone"
subcategory: ""
description: |-
  Manage an index.
---

# pinecone_index (Resource)

Manage an index.

## Example Usage

```terraform
resource "pinecone_index" "test" {
  name      = "test"
  dimension = 1536
  metric    = "dotproduct"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `dimension` (Number) The dimension of the index.
- `name` (String) The name of the index.

### Optional

- `metadata_config` (Attributes) The metadata config of the index. (see [below for nested schema](#nestedatt--metadata_config))
- `metric` (String) The metric of the index.
- `pod_type` (String) The pod type of the index.
- `pods` (Number) The number of pods of the index.
- `replicas` (Number) The number of replicas of the index.

### Read-Only

- `id` (String) The ID of the index.
- `last_updated` (String) The last updated time of the index.

<a id="nestedatt--metadata_config"></a>
### Nested Schema for `metadata_config`

Optional:

- `indexed` (List of String) The indexed fields of the index.

## Import

Import is supported using the following syntax:

```shell
terraform import pinecone_index.test test
```
