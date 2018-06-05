package controller

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/yarntime/aiops/pkg/client"
	"github.com/yarntime/aiops/pkg/types"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	batch "k8s.io/client-go/pkg/apis/batch/v2alpha1"
	"k8s.io/client-go/pkg/util"
	"k8s.io/client-go/util/workqueue"
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
	// Jobs that need to be updated
	queue         workqueue.RateLimitingInterface
	currentCount  int32
	countLimit    int32
	processPeriod time.Duration
}

func NewJobController(c *types.Config) *JobController {
	return &JobController{
		k8sClient:     client.NewK8sClint(c.Host),
		config:        c,
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "job"),
		currentCount:  0,
		countLimit:    c.CustomCfg.Global.JobCountLimit,
		processPeriod: c.CustomCfg.Global.JobProcessPeriod,
	}
}

func (jc *JobController) Run(stopCh <-chan struct{}) {
	glog.Infof("Starting job worker")
	defer glog.Infof("Shutting down job worker")
	go wait.Until(jc.worker, time.Second, stopCh)
	<-stopCh
}

func (jc *JobController) worker() {
	for jc.processNextWorkItem() {
		jc.currentCount++
		if jc.currentCount == jc.countLimit {
			jc.currentCount = 0
			time.Sleep(jc.processPeriod)
		}
	}
}

func (jc *JobController) processNextWorkItem() bool {
	key, quit := jc.queue.Get()
	if quit {
		return false
	}
	defer jc.queue.Done(key)

	job := key.(*batchv1.Job)

	glog.V(4).Info("Creating job %s/%s", job.Namespace, job.Name)
	_, err := jc.k8sClient.BatchV1().Jobs(job.Namespace).Create(job)
	if err != nil {
		glog.Errorf("Failed to create job: %s/%s, %s", job.Namespace, job.Name, err.Error())
	}
	glog.V(4).Info("Creating job %s/%s", job.Namespace, job.Name)
	return true
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

	glog.V(4).Info("Putting job %s/%s in the queue", job.Namespace, job.Name)

	jc.queue.Add(job)
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
