# ABOUT
This is a simple program to make the migration of [Pulp Operator](https://docs.pulpproject.org/pulp_operator/) from [Ansible](https://github.com/pulp/pulp-operator/tree/ansible) to [Golang](https://github.com/pulp/pulp-operator/tree/main) easier.


# RUNNING
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


# WHAT IT DOES?

* it verifies the current database PVC, SVC, and STS names
* it gathers the current subscription's CSV name
* with the above information it will delete the current Pulp operator subscription and csv associated with it
* after that it will delete the current deployments, downscale database replicas, and update the database service to use the new database pods as endpoints
* as a last step it will subscribe to the new operator version and migrate the current CR to match the new CRD specification