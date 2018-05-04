package controller

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/yarntime/aiops/pkg/client"
	"github.com/yarntime/aiops/pkg/types"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	batch "k8s.io/client-go/pkg/apis/batch/v2alpha1"
	"k8s.io/client-go/pkg/util"
)

const (
	AIOpsJobs           = "skyform-ai-job"
	ContainerNamePrefix = "training-job"
)

type JobController struct {
	k8sClient *k8s.Clientset
	config    *types.Config
}

func NewJobController(c *types.Config) *JobController {
	return &JobController{
		k8sClient: client.NewK8sClint(c.Host),
		config:    c,
	}
}

func componentCronJob(obj *types.MonitorObject, customConf types.CustomConfig, appConf types.Application) *batch.CronJob {
	labels := map[string]string{
		"type":     AIOpsJobs,
		"tier":     appConf.Application,
		"host":     obj.Host,
		"instance": obj.InstanceName,
		"metric":   obj.Metric,
	}
	objParams := []string{
		fmt.Sprintf("--host=%s", obj.Host),
		fmt.Sprintf("--instance_name=%s", obj.InstanceName),
		fmt.Sprintf("--kpi=%s", obj.Metric),
	}
	allParams := append(appConf.Params, objParams...)
	return &batch.CronJob{
		ObjectMeta: meta_v1.ObjectMeta{
			GenerateName: "train-",
			Namespace:    customConf.Global.Namespace,
			Labels:       labels,
		},
		Spec: batch.CronJobSpec{
			Schedule:                   appConf.Cron,
			ConcurrencyPolicy:          batch.ForbidConcurrent,
			SuccessfulJobsHistoryLimit: util.Int32Ptr(5),
			FailedJobsHistoryLimit:     util.Int32Ptr(10),
			JobTemplate: batch.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Parallelism: util.Int32Ptr(1),
					Completions: util.Int32Ptr(1),
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:      ContainerNamePrefix,
									Image:     customConf.Global.Image,
									Command:   appConf.Cmd,
									Args:      allParams,
									Resources: componentResources("500m"),
								},
							},
							RestartPolicy: v1.RestartPolicyOnFailure,
						},
					},
				},
			},
		},
	}
}

func componentResources(cpu string) v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceName(v1.ResourceCPU): resource.MustParse(cpu),
		},
	}
}

func (jc *JobController) CreateTrainingJob(obj *types.MonitorObject, customConf types.CustomConfig, appConf types.Application) {
	job := componentCronJob(obj, customConf, appConf)
	selector := labels.Set(job.Labels).AsSelector()
	listOptions := meta_v1.ListOptions{
		LabelSelector: selector.String(),
	}
	previousJobs, _ := jc.k8sClient.BatchV2alpha1().CronJobs(job.Namespace).List(listOptions)

	for _, previousJob := range previousJobs.Items {
		jc.k8sClient.BatchV2alpha1().CronJobs(previousJob.Namespace).Delete(previousJob.Name, &meta_v1.DeleteOptions{})
	}

	_, err := jc.k8sClient.BatchV2alpha1().CronJobs(job.Namespace).Create(job)
	if err != nil {
		glog.Errorf("Failed to create training job: %s/%s, %s", job.Namespace, job.Name, err.Error())
	}
}
