# set the namespace on each resource
def run(r, ns_value):
  for resource in r:
    # mutate the resource
    resource["metadata"]["namespace"] = ns_value
# get the value to add
ns_value = ctx.resource_list["functionConfig"]["keyValues"]["foo"]
run(ctx.resource_list["items"], ns_value)
