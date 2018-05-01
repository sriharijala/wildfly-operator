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
  labels:
      app: my-label
  configMapName: standalone
  standaloneConfigKey: standalone-full-ha-k8s.xml
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

***Check LoadBalancer is available***

```
kubectl get svc
```

There should be a `wildfly-example` LoadBalancer with an external IP associated. Applications and Wildfly Management Console are reachable through this external IP.






