# set the namespace on each resource
def run(r, ns_value):
  for resource in r:
    # mutate the resource
    resource["metadata"]["namespace"] = ns_value

# get the value of the annotation to add
ns_value = ctx.resource_list["functionConfig"]["spec"]["namespace_value"]

run(ctx.resource_list["items"], ns_value)
