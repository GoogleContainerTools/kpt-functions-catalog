def setReplicas(resources, replicas):
    for r in resources:
        if r["apiVersion"] == "apps/v1" and r["kind"] == "Deployment":
            r["spec"]["replicas"] = replicas

replicas = ctx.resource_list["functionConfig"]["data"]["replicas"]
setReplicas(ctx.resource_list["items"], replicas)
