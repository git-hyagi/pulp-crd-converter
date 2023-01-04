# ABOUT
This is a simple program to make the migration of [Pulp Operator](https://docs.pulpproject.org/pulp_operator/) installations made through [OLM](https://operatorhub.io/about#How-does-OperatorHub.io-work?) from [Ansible](https://github.com/pulp/pulp-operator/tree/ansible) to [Golang](https://github.com/pulp/pulp-operator/tree/main) easier.


# QUICKSTART
> :warning: MAKE SURE TO HAVE A BACKUP BEFORE PROCEEDING :warning:

> :warning: MAKE SURE TO HAVE A BACKUP BEFORE PROCEEDING :warning:

> :warning: MAKE SURE TO HAVE A BACKUP BEFORE PROCEEDING :warning:


Apply `migrator-job.yaml` to:
* create a new `serviceAccount` to run the commands
* since we need to delete/recreate a `subscription` we will use a `cluster-admin` clusterrole
* create the job that will run the migration
```
export PULP_RESOURCE_NAME=example-pulp
export PULP_NAMESPACE=pulp
envsubst < migrator-job.yaml |oc apply -f-
```

* when the migration finishes, we can remove the clusterrole from `serviceAccount` and delete it
```
oc adm policy remove-cluster-role-from-user cluster-admin -z migrator
oc delete sa migrator
```

# CONFIGURING

All the configurations are made through environment variables.
The `envsubst` command will fill the `migrator-job.yaml` with the values.

| Env var | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| PULP_NAMESPACE | Namespace where Pulp Operator is installed. | string | true |
| PULP_RESOURCE_NAME | Name of ansible operator's custom resource. Can be retrieved through: `oc get pulps.pulp` | string | true |
| NEW_PULP_RESOURCE_NAME | Name of the golang operator's custom resource. If not provided will use the same value as `PULP_RESOURCE_NAME`. | string | false |
| PULP_SUBSCRIPTION_NAME | Name of ansible Pulp Operator subscription. Default: `pulp-operator` | string | false |
| NEW_PULP_SUBSCRIPTION_NAME | Name of golang Pulp Operator subscription. If not provided will use the same value as `PULP_SUBSCRIPTION_NAME` | string | false |
| NEW_SUBSCRIPTION_CHANNEL | Golang Operator subscription channel ("release version"). Default: `beta` | string | false |
| NEW_SUBSCRIPTION_INSTALL_PLAN_APPROVAL | Approval is the user approval policy for an InstallPlan. It must be one of "Automatic" or "Manual". Default: `Automatic` | string | false |
| NEW_SUBSCRIPTION_SOURCE | CatalogSource ("repository") of golang Operator. Default: `community-operators` | string | false |
| NEW_SUBSCRIPTION_SOURCE_NAMESPACE | Namespace of CatalogSource ("repository") of golang Operator. Default: `openshift-marketplace` | string | false |
| NEW_SUBSCRIPTION_STARTING_CSV | Version of golang Pulp Operator to install. Default: `pulp-operator.v1.0.0-alpha.5` | string | false |
| NEW_PULP_API | Golang Pulp Operator APIVersion. Default: `repo-manager.pulpproject.org/v1alpha1` | string | false |
| NEW_PULP_KIND | Golang Pulp Operator Kind. Default: `Pulp` | string | false |
| PULP_RESOURCE | Ansible Pulp Operator Resource Type. Default: `pulps` | string | false |
| NEW_PULP_RESOURCE | Golang Pulp Operator Resource Type. If not provided will use the same value as `PULP_RESOURCE` | string | false |
| CONVERTION_ONLY | Define if the job should run only the convertion of Pulp CR from ansible to golang. Default: `false` | string | false |


By default, the `migrator-job` will run a lot of [steps](#what-does-it-do), but it is also possible to instruct it to only run the convertion procedure by setting the `CONVERTION_ONLY` env var to `true`.  
The convertion procedure will create a new `golang Pulp CR` with the data collected from `ansible Pulp CR`. In this case, all of the [other steps](#what-does-it-do) done by `migrator-job` should be run manually if needed.

# ROLLBACK
To rollback the changes, just remove the resources created by `migrator` and, in case of any, from `go-based` version:
```
$ oc -npulp delete csv,sub -l operators.coreos.com/pulp-operator.pulp=
$ oc -npulp delete deployments,svc,sts -l app.kubernetes.io/managed-by=<deployment type>-operator
```

Reinstall the ansible version of the operator, for example:
```
$ oc apply -f-<<EOF
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  labels:
    operators.coreos.com/pulp-operator.pulp: ""
  name: pulp-operator
  namespace: pulp
spec:
  channel: alpha
  installPlanApproval: Automatic
  name: pulp-operator
  source: pulp-operator-cs
  sourceNamespace: openshift-marketplace
  startingCSV: pulp-operator.v1.0.0-alpha
EOF
```
> :blue_book: Since ansible version is not available in the operatorhub catalog anymore, to reinstall it through OLM we need to create a custom catalogsource with it: https://gist.github.com/git-hyagi/00de7b25e2f935b83ea91969cc24eea1

It is also possible to reinstall it using `make deploy`:
```
$ git clone -b ansible git@github.com:pulp/pulp-operator.git
$ cd pulp-operator
$ make deploy NAMESPACE=pulp
```


# WHAT DOES IT DO?

* it verifies the current database PVC, SVC, and STS names
* it gathers the current subscription's CSV name
* with the above information it will delete the current Pulp operator subscription and csv associated with it
* after that it will delete the current deployments, downscale database replicas, and update the database service to use the new database pods as endpoints
* as a last step it will subscribe to the new operator version and migrate the current CR to match the new CRD specification

> :blue_book: All `PVCs`, `Secrets`, and `ConfigMaps` will remain the same (they will **not**  be modified by `migrator`), which allows to do a rollback or **manually** retry a migration in case of failure.


```
$ oc logs -f jobs/pulp-migrator
🔎 Retrieving the current Database PVC ...
Migrator will use the following PVC to the database pods: postgres-example-pulp-postgres-13-0
🔎 Retrieving the current Database Service ...
Migrator will use the following SVC to the database pods: example-pulp-postgres-13
🔎 Retrieving the current Database StatefulSet ...
Migrator will downscale the following StatefulSet to 0 replica pods: example-pulp-postgres-13
🔎 Retrieving the current csv from subscription pulp-operator ...
Current CSV Name: pulp-operator.v1.0.0-alpha
🗑️  Deleting pulp-operator subscription ...
{"kind":"Status","apiVersion":"v1","metadata":{},"status":"Success","details":{"name":"pulp-operator","group":"operators.coreos.com","kind":"subscriptions","uid":"75d1ad18-fdc7-42ac-a3fe-95bcf9fb8f7b"}}

🗑️  Deleting pulp-operator.v1.0.0-alpha CSV ...
{"kind":"Status","apiVersion":"v1","metadata":{},"status":"Success","details":{"name":"pulp-operator.v1.0.0-alpha","group":"operators.coreos.com","kind":"clusterserviceversions","uid":"d1309154-8e0a-4972-a8e7-ce457b21fb22"}}

🗑️  Deleting api deployment ...
🗑️  Deleting content-server deployment ...
🗑️  Deleting worker deployment ...
🗑️  Deleting webserver deployment ...
Scaling old Database STS to 0 replicas ...
Updating example-pulp-postgres-13 Database Service ...
Subscribing to the new Operator version ...
{"kind":"Subscription","apiVersion":"operators.coreos.com/v1alpha1","metadata":{"name":"pulp-operator","namespace":"pulp","creationTimestamp":null},"spec":{"source":"community-operators","sourceNamespace":"openshift-marketplace","name":"pulp-operator","channel":"beta","startingCSV":"pulp-operator.v1.0.0-alpha.4","installPlanApproval":"Automatic"},"status":{"lastUpdated":null}}
{"apiVersion":"operators.coreos.com/v1alpha1","kind":"Subscription","metadata":{"creationTimestamp":"2022-12-30T15:25:05Z","generation":1,"managedFields":[{"apiVersion":"operators.coreos.com/v1alpha1","fieldsType":"FieldsV1","fieldsV1":{"f:spec":{".":{},"f:channel":{},"f:installPlanApproval":{},"f:name":{},"f:source":{},"f:sourceNamespace":{},"f:startingCSV":{}}},"manager":"pulp-migrator","operation":"Update","time":"2022-12-30T15:25:05Z"}],"name":"pulp-operator","namespace":"pulp","resourceVersion":"9366641","uid":"e72be29b-a7ca-4b87-b12a-05e181d03d18"},"spec":{"channel":"beta","installPlanApproval":"Automatic","name":"pulp-operator","source":"community-operators","sourceNamespace":"openshift-marketplace","startingCSV":"pulp-operator.v1.0.0-alpha.4"}}

Converting Pulp CR to the new CRD ...
Create new CR: {"kind":"Pulp","apiVersion":"repo-manager.pulpproject.org/v1alpha1","metadata":{"name":"example-pulp","namespace":"pulp","creationTimestamp":null},"spec":{"file_storage_size":"10Gi","file_storage_access_mode":"ReadWriteMany","storage_type":"File","ingress_type":"nodeport","haproxy_timeout":"180s","nginx_client_max_body_size":"10m","nginx_proxy_body_size":"10m","nginx_proxy_read_timeout":"120s","nginx_proxy_connect_timeout":"120s","nginx_proxy_send_timeout":"120s","image_version":"nightly","image_pull_policy":"IfNotPresent","api":{"replicas":1,"gunicorn_timeout":90,"gunicorn_workers":2,"resource_requirements":{},"strategy":{}},"database":{"postgres_resource_requirements":{},"pvc":"postgres-example-pulp-postgres-13-0"},"content":{"replicas":1,"resource_requirements":{},"gunicorn_timeout":90,"gunicorn_workers":2,"strategy":{}},"worker":{"replicas":1,"resource_requirements":{},"strategy":{}},"web":{"replicas":1,"resource_requirements":{}},"cache":{"redis_resource_requirements":{},"strategy":{}},"pulp_settings":{"allowed_export_paths":["/tmp"],"allowed_import_paths":["/tmp"],"telemetry":false},"image_web_version":"nightly","admin_password_secret":"example-pulp-admin-password"},"status":{"conditions":null}}
Waiting for new CRD be created ... : the server could not find the requested resource
CRD: {"kind":"APIResourceList","apiVersion":"v1","groupVersion":"repo-manager.pulpproject.org/v1alpha1","resources":[{"name":"pulpbackups","singularName":"pulpbackup","namespaced":true,"kind":"PulpBackup","verbs":["delete","deletecollection","get","list","patch","create","update","watch"],"storageVersionHash":"aAreXaOGRJ0="},{"name":"pulpbackups/status","singularName":"","namespaced":true,"kind":"PulpBackup","verbs":["get","patch","update"]},{"name":"pulprestores","singularName":"pulprestore","namespaced":true,"kind":"PulpRestore","verbs":["delete","deletecollection","get","list","patch","create","update","watch"],"storageVersionHash":"aHYzRhXqFe8="},{"name":"pulprestores/status","singularName":"","namespaced":true,"kind":"PulpRestore","verbs":["get","patch","update"]},{"name":"pulps","singularName":"pulp","namespaced":true,"kind":"Pulp","verbs":["delete","deletecollection","get","list","patch","create","update","watch"],"storageVersionHash":"M1rgAm1eJDo="},{"name":"pulps/status","singularName":"","namespaced":true,"kind":"Pulp","verbs":["get","patch","update"]}]}

✅ Migration finished
```