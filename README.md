# Wildfly Operator

`Wildfly` Operator let's you describe and deploy your JEE application on `Wildfly` server by creating a custom resource definition 
in `Kubernetes`: `WildflyAppServer`.
Once you deploy a `WildflyAppServer` resource, the operator starts up a `WildFly` server cluster with a given number of nodes running from the provided container image in `standalone` mode.
JEE application to deploy must be in the provided image in `/opt/jboss/wildfly/standalone/deployments` folder. It will also start a `Load balancer` service so that Wildfly Management Console and deployed Web application are reachable from outside.
Nodes are joining together with JGroups configured to use `KUBE_PING` protocol which finds each `Pod` running a Wildfly server based on labels and namespace.
You can use the default standalone Full HA configuration for `Kubernetes` or you can provide your own in a `ConfigMap`. In this case you had to specify name of the config map and the key containing standalone.xml configuration.
Default `Wildfly` username/password is `admin/wildfly`, you can set up your own in a `Secret` which should have the same name 
as your `WildflyAppServer` resource. 

Example `WildflyServerApp`:

```
apiVersion: "wildfly.banzaicloud.com/v1alpha1"
kind: "WildflyAppServer"
metadata:
  name: "wildfly-example"
  labels:
    app: my-label
spec:
  nodeCount: 2
  image: "banzaicloud/wildfly-demo:0.0.51"
  applicationPath: "applicationPetstore"
  labels:
      app: my-label
  configMapName: standalone
  standaloneConfigKey: standalone-full-ha-k8s.xml
  dataSourceConfig:
    mariadb:
      hostName: "demo-mariadb"
      databaseName: "petstore"
      jndiName: "java:jboss/datasources/ExampleDS"
      user: "xxxxxxxx"
      password: "xxxxxxxx"
```

Example `Secret` containing credentials for Wildfly server:

```
apiVersion: v1
kind: Secret
metadata:
  name: "wildfly-example"
type: Opaque
data:
  wildfly-admin-user: "YOUR_USERNAME_IN_BASE_64_ENCODED_FROM"
  wildfly-admin-password: "YOUR_PASSWORD_IN_BASE_64_ENCODED_FROM"
```

NOTE: In this example we are using an example image, which contains a [Application - Petstore Java EE 7](https://github.com/banzaicloud/agoncal-application-petstore-ee7/tree/master-k8s).

Resource params:

    - spec.nodeCount : number of server nodes to run
    - spec.configMapName : name of ConfigMap containing standalone.xml config (optional) 
    - spec.standaloneConfigKey : key name containing standalone.xml config in ConfigMap (optional) 
    - spec.applicationPath : web application path, used for HTTP liveness probe 
    - spec.dataSourceConfig : optional datasource config, which will be set up in case no custom config and `ConfigMap` is provided in `spec.configMapName`, `spec.standaloneConfigKey`. 
        Currently only mariadb is supported.
    
    
***Deploy operator and custom resource***

```
kubectl apply -f deploy
```

***Check resource is deployed***

```
kubectl get WildflyAppServer
```


***Check number of nodes running***

```
kubectl describe WildflyAppServer
```

```
Status:
  External Addresses:
    Application:  35.232.66.116:8080
    Management:   35.232.66.116:9990
  Nodes:
    wildfly-example-84b5986dbf-gxhbv
    wildfly-example-84b5986dbf-zvwml
```

In status you should find external addresses, once the `Loadbalancer` has been created.
Applications and Wildfly Management Console are reachable through listed external IP/port.







