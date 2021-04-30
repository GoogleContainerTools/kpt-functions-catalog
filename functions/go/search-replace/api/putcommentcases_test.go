package api

var putPatternCases = []test{
	{
		name: "put comment single setter",
		config: `
data:
  by-value: '3'
  put-comment: 'kpt-set: ${replicas}'
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
		out: `${filePath}
fieldPath: spec.replicas
value: 3 # kpt-set: ${replicas}

Mutated 1 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3 # kpt-set: ${replicas}
 `,
	},
	{
		name: "put comment group of setters",
		config: `
data:
  by-value: nginx-deployment
  put-comment: 'kpt-set: ${image}-${kind}'
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
 `,
		out: `${filePath}
fieldPath: metadata.name
value: nginx-deployment # kpt-set: ${image}-${kind}

Mutated 1 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment # kpt-set: ${image}-${kind}
spec:
  replicas: 3
 `,
	},
	{
		name: "put comment array setter",
		config: `
data:
  by-path: spec.images
  put-comment: 'kpt-set: ${image}'
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: 
    - nginx
    - ubuntu
 `,
		out: `${filePath}
fieldPath: spec.images # kpt-set: ${image}
value: [nginx, ubuntu]

Mutated 1 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: # kpt-set: ${image}
    - nginx
    - ubuntu
 `,
	},
	{
		name: "put comment array setter flow style",
		config: `
data:
  by-path: spec.images
  put-comment: 'kpt-set: ${image}'
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: [nginx, ubuntu]
  non-matching-list: [foo, bar]
 `,
		out: `${filePath}
fieldPath: spec.images # kpt-set: ${image}
value: [nginx, ubuntu]

Mutated 1 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: # kpt-set: ${image}
    - nginx
    - ubuntu
  non-matching-list: [foo, bar]
 `,
	},
	{
		name: "put comment by value",
		config: `
data:
  by-value: 'dev/my-project/nginx'
  put-comment: 'kpt-set: ${env}/${project}/${name}'
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dev/my-project/nginx
spec:
  replicas: 3
 `,
		out: `${filePath}
fieldPath: metadata.name
value: dev/my-project/nginx # kpt-set: ${env}/${project}/${name}

Mutated 1 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dev/my-project/nginx # kpt-set: ${env}/${project}/${name}
spec:
  replicas: 3
 `,
	},
	{
		name: "put comment by capture groups simple case",
		config: `
data:
  by-value-regex: 'my-project-(.*)'
  put-comment: 'kpt-set: ${project}-${1}'
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-project-deployment
  namespace: my-project-namespace
 `,
		out: `${filePath}
fieldPath: metadata.name
value: my-project-deployment # kpt-set: ${project}-deployment

${filePath}
fieldPath: metadata.namespace
value: my-project-namespace # kpt-set: ${project}-namespace

Mutated 2 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-project-deployment # kpt-set: ${project}-deployment
  namespace: my-project-namespace # kpt-set: ${project}-namespace
 `,
	},
	{
		name: "put value and comment by regex capture groups",
		config: `
data:
  by-value-regex: '(\w+)-dev-(\w+)-us-east-1-(\w+)'
  put-value: '${1}-prod-${2}-us-central-1-${3}'
  put-comment: 'kpt-set: ${1}-${environment}-${2}-${region}-${3}'
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo1-dev-bar1-us-east-1-baz1
  namespace: foo2-dev-bar2-us-east-1-baz2
 `,
		out: `${filePath}
fieldPath: metadata.name
value: foo1-prod-bar1-us-central-1-baz1 # kpt-set: foo1-${environment}-bar1-${region}-baz1

${filePath}
fieldPath: metadata.namespace
value: foo2-prod-bar2-us-central-1-baz2 # kpt-set: foo2-${environment}-bar2-${region}-baz2

Mutated 2 field(s)
`,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo1-prod-bar1-us-central-1-baz1 # kpt-set: foo1-${environment}-bar1-${region}-baz1
  namespace: foo2-prod-bar2-us-central-1-baz2 # kpt-set: foo2-${environment}-bar2-${region}-baz2
 `,
	},
	{
		name: "put value and comment by regex capture groups error",
		config: `
data:
  by-value-regex: '(\w+)-dev-(\w+)-us-east-1-(\w+)'
  put-value: '${1}-prod-${2}-us-central-1-${3}'
  put-comment: 'kpt-set: ${1}-${environment}-${2}-${region}-${3}-extra-${4}'
`,
		input: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo1-dev-bar1-us-east-1-baz1
  namespace: foo2-dev-bar2-us-east-1-baz2
 `,
		expectedResources: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo1-dev-bar1-us-east-1-baz1
  namespace: foo2-dev-bar2-us-east-1-baz2
 `,
		errMsg: "unable to resolve capture groups",
	},
}
