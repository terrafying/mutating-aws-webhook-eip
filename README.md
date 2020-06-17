# Kubernetes Mutating Webhook for EIP Allocation

This readme shows how to build and deploy this [AdmissionWebhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#admission-webhooks).

In our mutating webhook we make an annotation for `service.beta.kubernetes.io/aws-load-balancer-eip-allocations`, which instructs the AWS provider to choose an EIP Allocation based on its allocation ID.  The feature we add is that these can now be specified as `ip.brivo.com/ip: 1.2.3.4`, which are transformed into the service annotations.

## Prerequisites

Kubernetes 1.9.0 or above with the `admissionregistration.k8s.io/v1beta1` API enabled. Verify that by the following command:
```
kubectl api-versions | grep admissionregistration.k8s.io/v1beta1
```
The result should be:
```
admissionregistration.k8s.io/v1beta1
```

In addition, the `MutatingAdmissionWebhook` and `ValidatingAdmissionWebhook` admission controllers should be added and listed in the correct order in the admission-control flag of kube-apiserver.

## Build

- Build and push docker image

```
make build image
```

## How does it work?

Here's a blog post that explains webhooks in depth with the help of a similar example. Check [it](https://brivo.com/blog/k8s-admission-webhooks/) out!
