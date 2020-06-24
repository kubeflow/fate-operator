# fate-operator: Kubernetes Operator for FATE (Federated AI Technology Enabler)
## Overview

The FATE-Operator makes it easy to deploy and run federated machine learning (FML) jobs on Kubernetes. It is a early version, all suggestions and feature requests are valueables for us. Raising issues is appreciate. 

## Background

FML is a machine learning setting where many clients (e.g. mobile devices or organizations) collaboratively train a model under the coordination of a central server while keeping the training data decentralized. Only the encrypted mediate parameters are exchanged between clients with MPC or homomorphic encryption.
![Federated Machine Learning](doc/diagrams/fate-operator-fl.png)

FML has received significant interest recently, because of its effectiveness to solve data silos and data privacy preserving problems. Companies participated in federated machine learning include 4Paradigm, Ant Financial, Data Republic, Google, Huawei, Intel, JD.com, Microsoft, Nvidia, OpenMind, Pingan Technology, Sharemind, Tencent, VMware, Webank etc. 

Depending on the differences in features and sample data space, federated machine learning can be classified into _horizontal federated machine learning_, _vertical federated machine learning_ and _federated transfer learning_. Horizontal federated machine learning is also called sample-based federated machine learning, which means data sets share the same feature space but have different samples. With horizontal federated machine learning, we can gather the relatively small or partial datasets into a big one to increase the performance of trained models. Vertical federated machine learning is applicable to the cases where there are two datasets with different feature space but share same sample ID. With vertical federated machine learning we can train a model with attributes from different organizations for a full profile. Vertical federated machine learning is required to redesign most of machine learning algorithms. Federated transfer learning applies to scenarios where there are two datasets with different features space but also different samples. 

[FATE (Federated AI Technology Enabler)](https://fate.fedai.org) is an open source project initialized by Webank, [now hosted at the Linux Foundation](https://fate.fedai.org/2019/09/18/first-digital-only-bank-in-china-joins-linux-foundation/). FATE is the only open source FML framework that supports both horizontal and vertical FML currently. The architecture design of FATE is focused on providing FML platform for enterprises. [KubeFATE](https://github.com/FederatedAI/KubeFATE) is an open source project to deploy FATE on Kubernetes and is a proven effective solution for FML use cases. 

More technologies of Federated machine learning, please refer to [Reference section](#reference)

**The design proposal of fate-operator can be found in [kubeflow/community/fate-operator-proposal.md](https://github.com/kubeflow/community/blob/master/proposals/fate-operator-proposal.md).**

## Quick user guide

### Installation
The fate-operator can be installed by following steps.

#### Install CRDs to Kubernetes
```bash
$ make install 
```

#### Uninstall CRDs from Kubernetes
```bash
$ make uninstall 
```

#### Building controller images
The Docker images are built and pushed to Dockerhub. 
[fate-operator](https://hub.docker.com/r/federatedai/fate-controller)

Alternatively, we can build the images manually by commands,
```bash
$ make docker-build
```

#### Deploying controller
```bash
$ make deploy
```

### Deploying KubeFATE
KubeFATE is the infrastructure management service for multiple FATE clusters in one organization. It will only deploy once. To deploy a KubeFATE service, we use the YAML refer to [app_v1beta1_kubefate.yaml](./config/samples/app_v1beta1_kubefate.yaml) as an example,

```bash
$ kubectl create -f ./config/samples/app_v1beta1_kubefate.yaml
```

### Deploying FATE
FATE is the cluster we run FML jobs. To deploy a FATE cluster, we use YAML refer to [app_v1beta1_fatecluster.yaml](./config/samples/app_v1beta1_fatecluster.yaml) as an example,
```bash
$ kubectl create -f ./config/samples/app_v1beta1_fatecluster.yaml
```

### Submitting an FML Job
Once KubeFATE and FATE cluster deployed, we can submit a FML job with `FateJob` config file. In current version, the FATE Job is defined by `DSL pipeline` and `Module Config` two parts. The details refer to [app_v1beta1_fatejob.yaml](./config/samples/app_v1beta1_fatejob.yaml). 
```bash
kubectl create -f ./config/samples/app_v1beta1_fatejob.yaml
```
In this example, only a `secure add` operation will be processed. For more example, we can refer to [FATE Examples](https://github.com/FederatedAI/FATE/tree/master/examples/federatedml-1.x-examples). In each example, the files end with `_dsl`, e.g. https://github.com/FederatedAI/FATE/blob/master/examples/federatedml-1.x-examples/hetero_linear_regression/test_hetero_linr_cv_job_dsl.json is the job pipeline and what should be put in `pipeline` field; and the files end with `_conf`, e.g. https://github.com/FederatedAI/FATE/blob/master/examples/federatedml-1.x-examples/hetero_linear_regression/test_hetero_linr_cv_job_conf.json is the config of each components and what should be put in `modulesConf` field.

### Checking created resource status
The status of created resource can be monitor with `kubectl get` command,
```bash
$ kubectl get kubefate,fatecluster,fatejob -A -o yaml
```

## Reference
1. Qiang Yang, Yang Liu, Tianjian Chen, and Yongxin Tong. Federated machine learning: Concept and applications. CoRR, abs/1902.04885, 2019. URL http://arxiv.org/abs/1902.04885
2. Peter Kairouz et al. Advances and open problems in federated learning. arXiv preprint arXiv:1912.04977

