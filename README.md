# kubespiffe

An experimental Kubernetes-native implementation of the [SPIFFE standard](https://spiffe.io/).

## Aims

`kubespiffe` should be able to:

* Issue SVIDs (both X.509 and JWT) to workloads on a Kubernetes cluster
* Manage SVID issuance with Kubernetes resources
* Support federation between two Kubernetes clusters

with the constraints:

* It should only have a single deployed component (i.e. a marked reduction in complexity from implementations like SPIRE, or `cert-manager` + `csi-driver-spiffe`)
* It should not have a completely disqualifying security posture

## Obtaining an SVID

```sequenceDiagram
    participant W as Workload
    participant K8S as Kubernetes API
    participant KS as kubespiffed

    W->>K8S: Request PSAT (projected volume)
    K8S-->>W: Returns PSAT (JWT signed by Kubernetes)

    W->>KS: Present PSAT
    KS->>K8S: Validate PSAT via TokenReview API
    K8S-->>KS: TokenReview response
    KS-->>W: Return SVID (X.509 or JWT-SVID)
```

