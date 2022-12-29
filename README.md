This is a simple program to make the migration of [Pulp Operator](https://docs.pulpproject.org/pulp_operator/) from [Ansible](https://github.com/pulp/pulp-operator/tree/ansible) to [Golang](https://github.com/pulp/pulp-operator/tree/main) easier.


> :warning: MAKE SURE TO HAVE A BACKUP BEFORE PROCEEDING

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
