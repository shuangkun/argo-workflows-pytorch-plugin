# argo workflows pytorch plugin


A plugin lets Argo Workflows orchestrate PyTorch jobs.


## Why argo-workflows-pytorch-plugin

* Submit tasks using non-string methods and more flexibly control and observe the status of pytorch jobs.

* Save costs. In scenarios where a large number of PyTorch jobs are orchestrated, there is no need to generate an equal number of resource pods.

## Getting Started

1. Enable Plugin capability for controller
```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: workflow-controller
spec:
  template:
    spec:
      containers:
        - name: workflow-controller
          env:
            - name: ARGO_EXECUTOR_PLUGINS
              value: "true"
```
2. Build argo-pytorch-plugin image

```
git clone https://github.com/shuangkun/argo-workflows-pytorch-plugin.git
cd argo-workflows-pytorch-plugin
docker build -t argo-pytorch-plugin:v1 .
```
3. Deploy argo-pytorch-plugin
```
kubectl apply -f pytorch-executor-plugin-configmap.yaml
```

4. Permission to create PyTorchJob CRD

```
kubctl apply -f install/role-secret.yaml
```

4. Submit PyTorch jobs
```
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: pytorch-training-
spec:
  entrypoint: pytorch-demo
  templates:
    - name: pytorch-demo
      plugin:
        pytorch:
          # The complete definition of PyTorchJob (PyTorch Operator CRD needs to be installed in advance)
          apiVersion: kubeflow.org/v1
          kind: PyTorchJob
          metadata:
            name: pytorch-distributed-job
            namespace: argo
          spec:
            pytorchReplicaSpecs:
              Master:
                replicas: 1
                template:
                  spec:
                    containers:
                    - name: pytorch
                      image: pytorch/pytorch:2.0.1-cuda11.7-cudnn8-devel
                      command: ["python", "/workspace/train.py"]
                      resources:
                        limits:
                          nvidia.com/gpu: 1
                      volumeMounts:
                        - name: code
                          mountPath: /workspace
              Worker:
                replicas: 2  # worker number
                template:
                  spec:
                    containers:
                    - name: pytorch
                      image: pytorch/pytorch:2.0.1-cuda11.7-cudnn8-devel
                      command: ["python", "/workspace/train.py"]
                      resources:
                        limits:
                          cpu: 4
                          memory: 8Gi
            runPolicy:
              cleanPodPolicy: Running
     
      inputs:
        parameters:
          - name: learning-rate
            value: "0.001"
      volumes:
        - name: code
          persistentVolumeClaim:
            claimName: training-code-pvc
```
5. Check agent logs
```
tianshuangkun@MacBook-Pro local % kubectl logs pytorch-training-5hjjd-1340600742-agent -c pytorch-executor-plugin
I0417 09:16:46.513777       1 main.go:67] v1.31.2+k3s1 <nil>
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

[GIN-debug] POST   /api/v1/template.execute  --> github.com/shuangkun/argo-workflows-pytorch-plugin/controller.(*PytorchJobController).ExecutePytorchJob-fm (3 handlers)
[GIN-debug] [WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.
Please check https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies for details.
[GIN-debug] Listening and serving HTTP on :3008
I0417 09:16:54.476097       1 pytorchjob.go:41] Receive: {"pytorch":{"apiVersion":"kubeflow.org/v1","kind":"PyTorchJob","metadata":{"name":"pytorch-distributed-job","namespace":"argo"},"spec":{"pytorchReplicaSpecs":{"Master":{"replicas":1,"template":{"spec":{"containers":[{"command":["python","/workspace/train.py"],"image":"pytorch/pytorch:2.0.1-cuda11.7-cudnn8-devel","name":"pytorch","resources":{"limits":{"nvidia.com/gpu":1}},"volumeMounts":[{"mountPath":"/workspace","name":"code"}]}]}}},"Worker":{"replicas":2,"template":{"spec":{"containers":[{"command":["python","/workspace/train.py"],"image":"pytorch/pytorch:2.0.1-cuda11.7-cudnn8-devel","name":"pytorch","resources":{"limits":{"cpu":4,"memory":"8Gi"}}}]}}}},"runPolicy":{"cleanPodPolicy":"Running"}}}}
[GIN] 2025/04/17 - 09:16:54 | 200 |    7.070738ms |             ::1 | POST     "/api/v1/template.execute"
I0417 09:17:04.490596       1 pytorchjob.go:41] Receive: {"pytorch":{"apiVersion":"kubeflow.org/v1","kind":"PyTorchJob","metadata":{"name":"pytorch-distributed-job","namespace":"argo"},"spec":{"pytorchReplicaSpecs":{"Master":{"replicas":1,"template":{"spec":{"containers":[{"command":["python","/workspace/train.py"],"image":"pytorch/pytorch:2.0.1-cuda11.7-cudnn8-devel","name":"pytorch","resources":{"limits":{"nvidia.com/gpu":1}},"volumeMounts":[{"mountPath":"/workspace","name":"code"}]}]}}},"Worker":{"replicas":2,"template":{"spec":{"containers":[{"command":["python","/workspace/train.py"],"image":"pytorch/pytorch:2.0.1-cuda11.7-cudnn8-devel","name":"pytorch","resources":{"limits":{"cpu":4,"memory":"8Gi"}}}]}}}},"runPolicy":{"cleanPodPolicy":"Running"}}}}
I0417 09:17:04.492566       1 pytorchjob.go:78] # found exists Pytorch Job: pytorch-distributed-jobreturning Status...{[] map[] <nil> <nil> <nil>}
[GIN] 2025/04/17 - 09:17:04 | 200 |    2.517687ms |             ::1 | POST     "/api/v1/template.execute"
I0417 09:17:04.492742       1 pytorchjob.go:156] ### Job pytorch-distributed-job Phase , status: Pending
```