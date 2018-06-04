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
	"strconv"
	"time"
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

func componentCronJob(obj *types.MonitorObject, customConf types.CustomConfig, appConf types.Application, objParams []string) *batch.CronJob {
	labels := map[string]string{
		"type":         AIOpsJobs,
		"tier":         appConf.Application,
		"id":           strconv.Itoa(obj.ID),
		"monitor_id":   strconv.Itoa(appConf.Id),
		"monitor_type": appConf.Application,
	}
	objParams = append(objParams, []string{
		fmt.Sprintf("--host=%s", obj.Host),
		fmt.Sprintf("--instance_name=%s", obj.InstanceName),
		fmt.Sprintf("--kpi=%s", obj.Metric),
		fmt.Sprintf("--index=%s", obj.ESIndex),
		fmt.Sprintf("--doc_type=%s", obj.ESType),
	}...)
	allParams := append(appConf.Params, objParams...)
	return &batch.CronJob{
		ObjectMeta: meta_v1.ObjectMeta{
			GenerateName: "train-",
			Namespace:    customConf.Global.Namespace,
			Labels:       labels,
		},
		Spec: batch.CronJobSpec{
			Schedule:                   appConf.Cron,
			ConcurrencyPolicy:          batch.ConcurrencyPolicy(customConf.Global.ConcurrencyPolicy),
			SuccessfulJobsHistoryLimit: util.Int32Ptr(customConf.Global.SuccessfulJobsHistoryLimit),
			FailedJobsHistoryLimit:     util.Int32Ptr(customConf.Global.FailedJobsHistoryLimit),
			JobTemplate: batch.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Parallelism: util.Int32Ptr(1),
					Completions: util.Int32Ptr(1),
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:            ContainerNamePrefix,
									Image:           appConf.Image,
									ImagePullPolicy: v1.PullPolicy(customConf.Global.ImagePullPolicy),
									Command:         appConf.Cmd,
									Args:            allParams,
									Resources:       componentResources(appConf.CpuRequest, appConf.MemoryRequest),
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

func componentResources(cpu string, mem string) v1.ResourceRequirements {
	result := v1.ResourceRequirements{
		Requests: v1.ResourceList{},
	}

	if len(cpu) != 0 {
		result.Requests[v1.ResourceName(v1.ResourceCPU)] = resource.MustParse(cpu)
	}

	if len(mem) != 0 {
		result.Requests[v1.ResourceName(v1.ResourceMemory)] = resource.MustParse(mem)
	}
	return result
}

func (jc *JobController) CreateCronJob(obj *types.MonitorObject, customConf types.CustomConfig, appConf types.Application, objParams []string) (*batch.CronJob, error) {
	job := componentCronJob(obj, customConf, appConf, objParams)
	job, err := jc.k8sClient.BatchV2alpha1().CronJobs(job.Namespace).Create(job)
	if err != nil {
		glog.Errorf("Failed to create cron job: %s/%s, %s", job.Namespace, job.Name, err.Error())
	}
	return job, err
}

func (jc *JobController) CreateJobFromCronJob(cj *batch.CronJob) {
	suffix := "-" + strconv.Itoa(int(time.Now().Unix()))
	job := &batchv1.Job{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      cj.Name + suffix,
			Namespace: cj.Namespace,
		},
		Spec: cj.Spec.JobTemplate.Spec,
	}
	_, err := jc.k8sClient.BatchV1().Jobs(job.Namespace).Create(job)
	if err != nil {
		glog.Errorf("Failed to create job: %s/%s, %s", job.Namespace, job.Name, err.Error())
	}
}

func (jc *JobController) DeleteCronJob(customConf types.CustomConfig) {
	selector := labels.Set(map[string]string{"type": AIOpsJobs}).AsSelector()
	listOptions := meta_v1.ListOptions{
		LabelSelector: selector.String(),
	}
	allJobs, _ := jc.k8sClient.BatchV2alpha1().CronJobs(customConf.Global.Namespace).List(listOptions)
	for _, job := range allJobs.Items {
		jc.k8sClient.BatchV2alpha1().CronJobs(job.Namespace).Delete(job.Name, &meta_v1.DeleteOptions{})
	}
}
