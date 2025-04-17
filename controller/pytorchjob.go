package controller

import (
	"encoding/json"
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	executorplugins "github.com/argoproj/argo-workflows/v3/pkg/plugins/executor"
	"github.com/gin-gonic/gin"
	pytorchjob "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	pytorchversioned "github.com/kubeflow/training-operator/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"net/http"
	"time"
)

const (
	LabelKeyWorkflow string = "workflows.argoproj.io/workflow"
)

type PytorchJobController struct {
	PytorchClient *pytorchversioned.Clientset
}

type PytorchPluginBody struct {
	PytorchJob *pytorchjob.PyTorchJob `json:"pytorch"`
}

func (ct *PytorchJobController) ExecutePytorchJob(ctx *gin.Context) {
	c := &executorplugins.ExecuteTemplateArgs{}
	err := ctx.BindJSON(&c)
	if err != nil {
		klog.Error(err)
		return
	}

	inputBody := &PytorchPluginBody{
		PytorchJob: &pytorchjob.PyTorchJob{},
	}

	pluginJson, _ := c.Template.Plugin.MarshalJSON()
	klog.Info("Receive: ", string(pluginJson))
	err = json.Unmarshal(pluginJson, &inputBody)
	if err != nil {
		klog.Error(err)
		ct.Response404(ctx)
		return
	}
	var msg string
	jobSpec := inputBody.PytorchJob.Spec
	if *(jobSpec.PyTorchReplicaSpecs["Master"].Replicas) < 0 || *(jobSpec.PyTorchReplicaSpecs["Worker"].Replicas) < 0 {
		msg = "job Replicas must be >= 0."
		klog.Error(msg)
		ct.ResponseMsg(ctx, wfv1.NodeFailed, msg)
		return
	}

	job := inputBody.PytorchJob

	if job.Name == "" {
		job.ObjectMeta.Name = c.Workflow.ObjectMeta.Name
	}

	if job.ObjectMeta.Namespace == "" {
		job.Namespace = "default"
	}

	var exists = false

	// 1. query job exists
	existsJob, err := ct.PytorchClient.KubeflowV1().PyTorchJobs(job.Namespace).Get(ctx, job.Name, metav1.GetOptions{})
	if err != nil {
		exists = false
	} else {
		exists = true
	}
	// 2. found and return
	if exists {
		klog.Info("# found exists Pytorch Job: ", job.Name, "returning Status...", job.Status)
		ct.ResponsePytorchJob(ctx, existsJob)
		return
	}

	// 3.Label keys with workflow Name
	InjectPytorchJobWithWorkflowName(job, c.Workflow.ObjectMeta.Name)

	newJob, err := ct.PytorchClient.KubeflowV1().PyTorchJobs(job.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		klog.Error("### " + err.Error())
		ct.ResponseMsg(ctx, wfv1.NodeFailed, err.Error())
		return
	}

	ct.ResponseCreated(ctx, newJob)

}

func (ct *PytorchJobController) ResponseCreated(ctx *gin.Context, job *pytorchjob.PyTorchJob) {
	message := ""
	if len(job.Status.Conditions) > 0 {
		message = job.Status.Conditions[len(job.Status.Conditions)-1].Message
	}
	ctx.JSON(http.StatusOK, &executorplugins.ExecuteTemplateReply{
		Node: &wfv1.NodeResult{
			Phase:   wfv1.NodePending,
			Message: message,
			Outputs: nil,
		},
		Requeue: &metav1.Duration{
			Duration: 10 * time.Second,
		},
	})
}

func (ct *PytorchJobController) ResponseMsg(ctx *gin.Context, status wfv1.NodePhase, msg string) {
	ctx.JSON(http.StatusOK, &executorplugins.ExecuteTemplateReply{
		Node: &wfv1.NodeResult{
			Phase:   status,
			Message: msg,
			Outputs: nil,
		},
	})
}

func (ct *PytorchJobController) ResponsePytorchJob(ctx *gin.Context, job *pytorchjob.PyTorchJob) {
	var status wfv1.NodePhase
	status = wfv1.NodeRunning
	if job.Status.StartTime == nil {
		status = wfv1.NodePending
	} else {
		masterFailed := &job.Status.ReplicaStatuses["Master"].Failed
		masterSucceeded := &job.Status.ReplicaStatuses["Master"].Succeeded
		workerFailed := &job.Status.ReplicaStatuses["Worker"].Failed
		workerSucceeded := &job.Status.ReplicaStatuses["Worker"].Succeeded
		if *masterFailed > 0 || *workerFailed > 0 {
			status = wfv1.NodeFailed
		} else if *masterSucceeded == *(job.Spec.PyTorchReplicaSpecs["Master"].Replicas) && *workerSucceeded == *(job.Spec.PyTorchReplicaSpecs["Worker"].Replicas) {
			status = wfv1.NodeSucceeded
		}
	}

	var requeue *metav1.Duration
	if status == wfv1.NodeRunning || status == wfv1.NodePending {
		requeue = &metav1.Duration{
			Duration: 10 * time.Second,
		}
	} else {
		requeue = nil
	}
	succeed := int32(0)
	total := int32(0)
	if job.Status.StartTime != nil {
		succeed = job.Status.ReplicaStatuses["Master"].Succeeded + job.Status.ReplicaStatuses["Worker"].Succeeded
		total = *(job.Spec.PyTorchReplicaSpecs["Master"].Replicas) + *(job.Spec.PyTorchReplicaSpecs["Worker"].Replicas)
	}
	progress, _ := wfv1.NewProgress(int64(succeed), int64(total))
	klog.Infof("### Job %v Phase "+", status: %v", job.Name, status)
	message := ""
	if len(job.Status.Conditions) > 0 {
		message = job.Status.Conditions[len(job.Status.Conditions)-1].Message
	}
	ctx.JSON(http.StatusOK, &executorplugins.ExecuteTemplateReply{
		Node: &wfv1.NodeResult{
			Phase:    status,
			Message:  message,
			Outputs:  nil,
			Progress: progress,
		},
		Requeue: requeue,
	})
}

func (ct *PytorchJobController) Response404(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotFound)
}

func InjectPytorchJobWithWorkflowName(job *pytorchjob.PyTorchJob, workflowName string) {
	masterJob := job.Spec.PyTorchReplicaSpecs["Master"]
	if masterJob.Template.ObjectMeta.Labels != nil {
		masterJob.Template.ObjectMeta.Labels[LabelKeyWorkflow] = workflowName
	} else {
		masterJob.Template.ObjectMeta.Labels = map[string]string{
			LabelKeyWorkflow: workflowName,
		}
	}
	slaveJob := job.Spec.PyTorchReplicaSpecs["Worker"]
	if slaveJob.Template.ObjectMeta.Labels != nil {
		slaveJob.Template.ObjectMeta.Labels[LabelKeyWorkflow] = workflowName
	} else {
		slaveJob.Template.ObjectMeta.Labels = map[string]string{
			LabelKeyWorkflow: workflowName,
		}
	}

	job.Spec.PyTorchReplicaSpecs["Master"] = masterJob
	job.Spec.PyTorchReplicaSpecs["Worker"] = slaveJob
}
