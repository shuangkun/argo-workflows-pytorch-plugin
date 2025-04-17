// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"

	v1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	kubefloworgv1 "github.com/kubeflow/training-operator/pkg/client/applyconfiguration/kubeflow.org/v1"
	scheme "github.com/kubeflow/training-operator/pkg/client/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// JAXJobsGetter has a method to return a JAXJobInterface.
// A group's client should implement this interface.
type JAXJobsGetter interface {
	JAXJobs(namespace string) JAXJobInterface
}

// JAXJobInterface has methods to work with JAXJob resources.
type JAXJobInterface interface {
	Create(ctx context.Context, jAXJob *v1.JAXJob, opts metav1.CreateOptions) (*v1.JAXJob, error)
	Update(ctx context.Context, jAXJob *v1.JAXJob, opts metav1.UpdateOptions) (*v1.JAXJob, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, jAXJob *v1.JAXJob, opts metav1.UpdateOptions) (*v1.JAXJob, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.JAXJob, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.JAXJobList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.JAXJob, err error)
	Apply(ctx context.Context, jAXJob *kubefloworgv1.JAXJobApplyConfiguration, opts metav1.ApplyOptions) (result *v1.JAXJob, err error)
	// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
	ApplyStatus(ctx context.Context, jAXJob *kubefloworgv1.JAXJobApplyConfiguration, opts metav1.ApplyOptions) (result *v1.JAXJob, err error)
	JAXJobExpansion
}

// jAXJobs implements JAXJobInterface
type jAXJobs struct {
	*gentype.ClientWithListAndApply[*v1.JAXJob, *v1.JAXJobList, *kubefloworgv1.JAXJobApplyConfiguration]
}

// newJAXJobs returns a JAXJobs
func newJAXJobs(c *KubeflowV1Client, namespace string) *jAXJobs {
	return &jAXJobs{
		gentype.NewClientWithListAndApply[*v1.JAXJob, *v1.JAXJobList, *kubefloworgv1.JAXJobApplyConfiguration](
			"jaxjobs",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1.JAXJob { return &v1.JAXJob{} },
			func() *v1.JAXJobList { return &v1.JAXJobList{} }),
	}
}
