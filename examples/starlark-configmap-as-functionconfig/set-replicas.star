def setReplicas(resources, replicas):
    for r in resources:
        if r["apiVersion"] == "apps/v1" and r["kind"] == "Deployment":
            r["spec"]["replicas"] = replicas

# The functionConfig is a ConfigMap, so the replicas we got from field
# functionConfig.data.replicas is a string. We need to convert it to an int.
replicas = int(ctx.resource_list["functionConfig"]["data"]["replicas"])
setReplicas(ctx.resource_list["items"], replicas)
