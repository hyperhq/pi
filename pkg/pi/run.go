/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pi

import (
	"fmt"
	"strconv"
	"strings"

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	batchv2alpha1 "k8s.io/api/batch/v2alpha1"
	"k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
)

type DeploymentV1Beta1 struct{}

func (DeploymentV1Beta1) ParamNames() []GeneratorParam {
	return []GeneratorParam{
		{"labels", false},
		{"default-name", false},
		{"name", true},
		{"replicas", true},
		{"image", true},
		{"image-pull-policy", false},
		{"port", false},
		{"hostport", false},
		{"stdin", false},
		{"tty", false},
		{"command", false},
		{"args", false},
		{"env", false},
		{"requests", false},
		{"limits", false},
		{"serviceaccount", false},
	}
}

func (DeploymentV1Beta1) Generate(genericParams map[string]interface{}) (runtime.Object, error) {
	args, err := getArgs(genericParams)
	if err != nil {
		return nil, err
	}

	envs, err := getEnvs(genericParams)
	if err != nil {
		return nil, err
	}

	params, err := getParams(genericParams)
	if err != nil {
		return nil, err
	}

	name, err := getName(params)
	if err != nil {
		return nil, err
	}

	labels, err := getLabels(params, name)
	if err != nil {
		return nil, err
	}

	count, err := strconv.Atoi(params["replicas"])
	if err != nil {
		return nil, err
	}

	podSpec, err := makePodSpec(params, name)
	if err != nil {
		return nil, err
	}

	imagePullPolicy := v1.PullPolicy(params["image-pull-policy"])
	if err = updatePodContainers(params, args, envs, imagePullPolicy, podSpec); err != nil {
		return nil, err
	}

	if err := updatePodPorts(params, podSpec); err != nil {
		return nil, err
	}

	count32 := int32(count)
	deployment := extensionsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: extensionsv1beta1.DeploymentSpec{
			Replicas: &count32,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: *podSpec,
			},
		},
	}
	return &deployment, nil
}

type DeploymentAppsV1Beta1 struct{}

func (DeploymentAppsV1Beta1) ParamNames() []GeneratorParam {
	return []GeneratorParam{
		{"labels", false},
		{"default-name", false},
		{"name", true},
		{"replicas", true},
		{"image", true},
		{"image-pull-policy", false},
		{"port", false},
		{"hostport", false},
		{"stdin", false},
		{"tty", false},
		{"command", false},
		{"args", false},
		{"env", false},
		{"requests", false},
		{"limits", false},
		{"serviceaccount", false},
	}
}

func (DeploymentAppsV1Beta1) Generate(genericParams map[string]interface{}) (runtime.Object, error) {
	args, err := getArgs(genericParams)
	if err != nil {
		return nil, err
	}

	envs, err := getEnvs(genericParams)
	if err != nil {
		return nil, err
	}

	params, err := getParams(genericParams)
	if err != nil {
		return nil, err
	}

	name, err := getName(params)
	if err != nil {
		return nil, err
	}

	labels, err := getLabels(params, name)
	if err != nil {
		return nil, err
	}

	count, err := strconv.Atoi(params["replicas"])
	if err != nil {
		return nil, err
	}

	podSpec, err := makePodSpec(params, name)
	if err != nil {
		return nil, err
	}

	imagePullPolicy := v1.PullPolicy(params["image-pull-policy"])
	if err = updatePodContainers(params, args, envs, imagePullPolicy, podSpec); err != nil {
		return nil, err
	}

	if err := updatePodPorts(params, podSpec); err != nil {
		return nil, err
	}

	count32 := int32(count)
	deployment := appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: appsv1beta1.DeploymentSpec{
			Replicas: &count32,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: *podSpec,
			},
		},
	}
	return &deployment, nil
}

// getLabels returns map of labels.
func getLabels(params map[string]string, name string) (map[string]string, error) {
	labelString, found := params["labels"]
	var labels map[string]string
	var err error
	if found && len(labelString) > 0 {
		labels, err = ParseLabels(labelString)
		if err != nil {
			return nil, err
		}
	} else {
		labels = map[string]string{
			"run": name,
		}
	}
	return labels, nil
}

// getImagePullSecrets returns map of imagePullSecrets.
func getImagePullSecrets(params map[string]string) ([]v1.LocalObjectReference, error) {
	str, found := params["image-pull-secrets"]
	var res []v1.LocalObjectReference
	var err error
	if found && len(str) > 0 {
		res, err = ParseImagePullSecrets(str)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

// getName returns the name of newly created resource.
func getName(params map[string]string) (string, error) {
	name, found := params["name"]
	if !found || len(name) == 0 {
		name, found = params["default-name"]
		if !found || len(name) == 0 {
			return "", fmt.Errorf("'name' is a required parameter.")
		}
	}
	return name, nil
}

// getParams returns map of generic parameters.
func getParams(genericParams map[string]interface{}) (map[string]string, error) {
	params := map[string]string{}
	for key, value := range genericParams {
		strVal, isString := value.(string)
		if !isString {
			return nil, fmt.Errorf("expected string, saw %v for '%s'", value, key)
		}
		params[key] = strVal
	}
	return params, nil
}

// getArgs returns arguments for the container command.
func getArgs(genericParams map[string]interface{}) ([]string, error) {
	args := []string{}
	val, found := genericParams["args"]
	if found {
		var isArray bool
		args, isArray = val.([]string)
		if !isArray {
			return nil, fmt.Errorf("expected []string, found: %v", val)
		}
		delete(genericParams, "args")
	}
	return args, nil
}

// getEnvs returns environment variables.
func getEnvs(genericParams map[string]interface{}) ([]v1.EnvVar, error) {
	var envs []v1.EnvVar
	envStrings, found := genericParams["env"]
	if found {
		if envStringArray, isArray := envStrings.([]string); isArray {
			var err error
			envs, err = parseEnvs(envStringArray)
			if err != nil {
				return nil, err
			}
			delete(genericParams, "env")
		} else {
			return nil, fmt.Errorf("expected []string, found: %v", envStrings)
		}
	}
	return envs, nil
}

// getVolumes returns volumes.
func getVolumes(genericParams map[string]interface{}) ([]v1.Volume, []v1.VolumeMount, error) {
	var volumes []v1.Volume
	var volumeMounts []v1.VolumeMount

	volumeStrings, found := genericParams["volume"]
	if found {
		if strArray, isArray := volumeStrings.([]string); isArray {
			var err error
			volumes, volumeMounts, err = parseVolumes(strArray)
			if err != nil {
				return nil, nil, err
			}
			delete(genericParams, "volume")
		} else {
			return nil, nil, fmt.Errorf("expected []string, found: %v", volumeStrings)
		}
	}
	return volumes, volumeMounts, nil
}

type JobV1 struct{}

func (JobV1) ParamNames() []GeneratorParam {
	return []GeneratorParam{
		{"labels", false},
		{"default-name", false},
		{"name", true},
		{"image", true},
		{"image-pull-policy", false},
		{"port", false},
		{"hostport", false},
		{"stdin", false},
		{"leave-stdin-open", false},
		{"tty", false},
		{"command", false},
		{"args", false},
		{"env", false},
		{"requests", false},
		{"limits", false},
		{"restart", false},
		{"serviceaccount", false},
		{"size", false},
		{"completions", false},
		{"parallelism", false},
		{"backoff-limit", false},
		{"active-deadline-seconds", false},
		{"image-pull-secrets", false},
		{"volume", false},
	}
}

func (JobV1) Generate(genericParams map[string]interface{}) (runtime.Object, error) {
	args, err := getArgs(genericParams)
	if err != nil {
		return nil, err
	}

	envs, err := getEnvs(genericParams)
	if err != nil {
		return nil, err
	}

	volumes, volumeMounts, err := getVolumes(genericParams)
	if err != nil {
		return nil, err
	}

	params, err := getParams(genericParams)
	if err != nil {
		return nil, err
	}

	name, err := getName(params)
	if err != nil {
		return nil, err
	}

	labels, err := getLabels(params, name)
	if err != nil {
		return nil, err
	}

	podSpec, err := makePodSpec(params, name)
	if err != nil {
		return nil, err
	}

	if err := updatePodVolumes(volumes, volumeMounts, podSpec); err != nil {
		return nil, err
	}

	imagePullPolicy := v1.PullPolicy(params["image-pull-policy"])
	if err = updatePodContainers(params, args, envs, imagePullPolicy, podSpec); err != nil {
		return nil, err
	}

	imagePullSecrets, err := getImagePullSecrets(params)
	if err != nil {
		return nil, err
	}
	if err := updatePodImagePullSecrets(imagePullSecrets, podSpec); err != nil {
		return nil, err
	}

	leaveStdinOpen, err := GetBool(params, "leave-stdin-open", false)
	if err != nil {
		return nil, err
	}
	podSpec.Containers[0].StdinOnce = !leaveStdinOpen && podSpec.Containers[0].Stdin

	if err := updatePodPorts(params, podSpec); err != nil {
		return nil, err
	}

	restartPolicy := v1.RestartPolicy(params["restart"])
	if len(restartPolicy) == 0 {
		restartPolicy = v1.RestartPolicyNever
	}
	podSpec.RestartPolicy = restartPolicy

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: *podSpec,
			},
		},
	}

	if _, ok := params["size"]; ok {
		if job.ObjectMeta.Annotations == nil {
			job.ObjectMeta.Annotations = map[string]string{}
		}
		job.ObjectMeta.Annotations["sh_hyper_instancetype"] = params["size"]
	}
	if params["completions"] != "" {
		if i, err := strconv.ParseInt(params["completions"], 10, 32); err != nil {
			return nil, fmt.Errorf("--completions should be a integer")
		} else {
			completions := int32(i)
			job.Spec.Completions = &completions
		}
	}
	if params["parallelism"] != "" {
		if i, err := strconv.ParseInt(params["parallelism"], 10, 32); err != nil {
			return nil, fmt.Errorf("--parallelism should be a integer")
		} else {
			parallelism := int32(i)
			job.Spec.Parallelism = &parallelism
		}
	}
	if params["backoff-limit"] != "" {
		if i, err := strconv.ParseInt(params["backoff-limit"], 10, 32); err != nil {
			return nil, fmt.Errorf("--parallelism should be a integer")
		} else {
			backoffLimit := int32(i)
			job.Spec.BackoffLimit = &backoffLimit
		}
	}
	if params["active-deadline-seconds"] != "" {
		if i, err := strconv.ParseInt(params["active-deadline-seconds"], 10, 64); err != nil {
			return nil, fmt.Errorf("--active-deadline-seconds should be a integer")
		} else {
			job.Spec.ActiveDeadlineSeconds = &i
		}
	}
	return &job, nil
}

type CronJobV2Alpha1 struct{}

func (CronJobV2Alpha1) ParamNames() []GeneratorParam {
	return []GeneratorParam{
		{"labels", false},
		{"default-name", false},
		{"name", true},
		{"image", true},
		{"image-pull-policy", false},
		{"port", false},
		{"hostport", false},
		{"stdin", false},
		{"leave-stdin-open", false},
		{"tty", false},
		{"command", false},
		{"args", false},
		{"env", false},
		{"requests", false},
		{"limits", false},
		{"restart", false},
		{"schedule", true},
		{"serviceaccount", false},
	}
}

func (CronJobV2Alpha1) Generate(genericParams map[string]interface{}) (runtime.Object, error) {
	args, err := getArgs(genericParams)
	if err != nil {
		return nil, err
	}

	envs, err := getEnvs(genericParams)
	if err != nil {
		return nil, err
	}

	params, err := getParams(genericParams)
	if err != nil {
		return nil, err
	}

	name, err := getName(params)
	if err != nil {
		return nil, err
	}

	labels, err := getLabels(params, name)
	if err != nil {
		return nil, err
	}

	podSpec, err := makePodSpec(params, name)
	if err != nil {
		return nil, err
	}

	imagePullPolicy := v1.PullPolicy(params["image-pull-policy"])
	if err = updatePodContainers(params, args, envs, imagePullPolicy, podSpec); err != nil {
		return nil, err
	}

	leaveStdinOpen, err := GetBool(params, "leave-stdin-open", false)
	if err != nil {
		return nil, err
	}
	podSpec.Containers[0].StdinOnce = !leaveStdinOpen && podSpec.Containers[0].Stdin

	if err := updatePodPorts(params, podSpec); err != nil {
		return nil, err
	}

	restartPolicy := v1.RestartPolicy(params["restart"])
	if len(restartPolicy) == 0 {
		restartPolicy = v1.RestartPolicyNever
	}
	podSpec.RestartPolicy = restartPolicy

	cronJob := batchv2alpha1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: batchv2alpha1.CronJobSpec{
			Schedule:          params["schedule"],
			ConcurrencyPolicy: batchv2alpha1.AllowConcurrent,
			JobTemplate: batchv2alpha1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: labels,
						},
						Spec: *podSpec,
					},
				},
			},
		},
	}

	return &cronJob, nil
}

type CronJobV1Beta1 struct{}

func (CronJobV1Beta1) ParamNames() []GeneratorParam {
	return []GeneratorParam{
		{"labels", false},
		{"default-name", false},
		{"name", true},
		{"image", true},
		{"image-pull-policy", false},
		{"port", false},
		{"hostport", false},
		{"stdin", false},
		{"leave-stdin-open", false},
		{"tty", false},
		{"command", false},
		{"args", false},
		{"env", false},
		{"requests", false},
		{"limits", false},
		{"restart", false},
		{"schedule", true},
		{"serviceaccount", false},
	}
}

func (CronJobV1Beta1) Generate(genericParams map[string]interface{}) (runtime.Object, error) {
	args, err := getArgs(genericParams)
	if err != nil {
		return nil, err
	}

	envs, err := getEnvs(genericParams)
	if err != nil {
		return nil, err
	}

	params, err := getParams(genericParams)
	if err != nil {
		return nil, err
	}

	name, err := getName(params)
	if err != nil {
		return nil, err
	}

	labels, err := getLabels(params, name)
	if err != nil {
		return nil, err
	}

	podSpec, err := makePodSpec(params, name)
	if err != nil {
		return nil, err
	}

	imagePullPolicy := v1.PullPolicy(params["image-pull-policy"])
	if err = updatePodContainers(params, args, envs, imagePullPolicy, podSpec); err != nil {
		return nil, err
	}

	leaveStdinOpen, err := GetBool(params, "leave-stdin-open", false)
	if err != nil {
		return nil, err
	}
	podSpec.Containers[0].StdinOnce = !leaveStdinOpen && podSpec.Containers[0].Stdin

	if err := updatePodPorts(params, podSpec); err != nil {
		return nil, err
	}

	restartPolicy := v1.RestartPolicy(params["restart"])
	if len(restartPolicy) == 0 {
		restartPolicy = v1.RestartPolicyNever
	}
	podSpec.RestartPolicy = restartPolicy

	cronJob := batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: batchv1beta1.CronJobSpec{
			Schedule:          params["schedule"],
			ConcurrencyPolicy: batchv1beta1.AllowConcurrent,
			JobTemplate: batchv1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: labels,
						},
						Spec: *podSpec,
					},
				},
			},
		},
	}

	return &cronJob, nil
}

type BasicReplicationController struct{}

func (BasicReplicationController) ParamNames() []GeneratorParam {
	return []GeneratorParam{
		{"labels", false},
		{"default-name", false},
		{"name", true},
		{"replicas", true},
		{"image", true},
		{"image-pull-policy", false},
		{"port", false},
		{"hostport", false},
		{"stdin", false},
		{"tty", false},
		{"command", false},
		{"args", false},
		{"env", false},
		{"requests", false},
		{"limits", false},
		{"serviceaccount", false},
	}
}

// populateResourceListV1 takes strings of form <resourceName1>=<value1>,<resourceName1>=<value2>
// and returns ResourceList.
func populateResourceListV1(spec string) (v1.ResourceList, error) {
	// empty input gets a nil response to preserve generator test expected behaviors
	if spec == "" {
		return nil, nil
	}

	result := v1.ResourceList{}
	resourceStatements := strings.Split(spec, ",")
	for _, resourceStatement := range resourceStatements {
		parts := strings.Split(resourceStatement, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("Invalid argument syntax %v, expected <resource>=<value>", resourceStatement)
		}
		resourceName := v1.ResourceName(parts[0])
		resourceQuantity, err := resource.ParseQuantity(parts[1])
		if err != nil {
			return nil, err
		}
		result[resourceName] = resourceQuantity
	}
	return result, nil
}

// HandleResourceRequirementsV1 parses the limits and requests parameters if specified
// and returns ResourceRequirements.
func HandleResourceRequirementsV1(params map[string]string) (v1.ResourceRequirements, error) {
	result := v1.ResourceRequirements{}
	limits, err := populateResourceListV1(params["limits"])
	if err != nil {
		return result, err
	}
	result.Limits = limits
	requests, err := populateResourceListV1(params["requests"])
	if err != nil {
		return result, err
	}
	result.Requests = requests
	return result, nil
}

// makePodSpec returns PodSpec filled with passed parameters.
func makePodSpec(params map[string]string, name string) (*v1.PodSpec, error) {
	stdin, err := GetBool(params, "stdin", false)
	if err != nil {
		return nil, err
	}

	tty, err := GetBool(params, "tty", false)
	if err != nil {
		return nil, err
	}

	resourceRequirements, err := HandleResourceRequirementsV1(params)
	if err != nil {
		return nil, err
	}

	spec := v1.PodSpec{
		ServiceAccountName: params["serviceaccount"],
		Containers: []v1.Container{
			{
				Name:      name,
				Image:     params["image"],
				Stdin:     stdin,
				TTY:       tty,
				Resources: resourceRequirements,
			},
		},
	}
	return &spec, nil
}

func (BasicReplicationController) Generate(genericParams map[string]interface{}) (runtime.Object, error) {
	args, err := getArgs(genericParams)
	if err != nil {
		return nil, err
	}

	envs, err := getEnvs(genericParams)
	if err != nil {
		return nil, err
	}

	params, err := getParams(genericParams)
	if err != nil {
		return nil, err
	}

	name, err := getName(params)
	if err != nil {
		return nil, err
	}

	labels, err := getLabels(params, name)
	if err != nil {
		return nil, err
	}

	count, err := strconv.Atoi(params["replicas"])
	if err != nil {
		return nil, err
	}

	podSpec, err := makePodSpec(params, name)
	if err != nil {
		return nil, err
	}

	imagePullPolicy := v1.PullPolicy(params["image-pull-policy"])
	if err = updatePodContainers(params, args, envs, imagePullPolicy, podSpec); err != nil {
		return nil, err
	}

	if err := updatePodPorts(params, podSpec); err != nil {
		return nil, err
	}

	count32 := int32(count)
	controller := v1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: v1.ReplicationControllerSpec{
			Replicas: &count32,
			Selector: labels,
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: *podSpec,
			},
		},
	}
	return &controller, nil
}

// updatePodContainers updates PodSpec.Containers with passed parameters.
func updatePodContainers(params map[string]string, args []string, envs []v1.EnvVar, imagePullPolicy v1.PullPolicy, podSpec *v1.PodSpec) error {
	if len(args) > 0 {
		command, err := GetBool(params, "command", false)
		if err != nil {
			return err
		}
		if command {
			podSpec.Containers[0].Command = args
		} else {
			podSpec.Containers[0].Args = args
		}
	}

	if len(envs) > 0 {
		podSpec.Containers[0].Env = envs
	}

	if len(imagePullPolicy) > 0 {
		// imagePullPolicy should be valid here since we have verified it before.
		podSpec.Containers[0].ImagePullPolicy = imagePullPolicy
	}
	return nil
}

// updatePodContainers updates PodSpec.Containers.Ports with passed parameters.
func updatePodPorts(params map[string]string, podSpec *v1.PodSpec) (err error) {
	port := -1
	hostPort := -1
	if len(params["port"]) > 0 {
		port, err = strconv.Atoi(params["port"])
		if err != nil {
			return err
		}
	}

	if len(params["hostport"]) > 0 {
		hostPort, err = strconv.Atoi(params["hostport"])
		if err != nil {
			return err
		}
		if hostPort > 0 && port < 0 {
			return fmt.Errorf("--hostport requires --port to be specified")
		}
	}

	// Don't include the port if it was not specified.
	if len(params["port"]) > 0 {
		podSpec.Containers[0].Ports = []v1.ContainerPort{
			{
				ContainerPort: int32(port),
			},
		}
		if hostPort > 0 {
			podSpec.Containers[0].Ports[0].HostPort = int32(hostPort)
		}
	}
	return nil
}

// updatePodImagePullSecrets updates PodSpec.ImagePullSecrets with passed parameters.
func updatePodImagePullSecrets(imagePullSecrets []v1.LocalObjectReference, podSpec *v1.PodSpec) (err error) {
	if len(imagePullSecrets) > 0 {
		podSpec.ImagePullSecrets = imagePullSecrets
	}
	return nil
}

// updatePodActiveDeadlineSeconds updates PodSpec.ActiveDeadlineSeconds with passed parameters.
func updatePodActiveDeadlineSeconds(params map[string]string, podSpec *v1.PodSpec) (err error) {
	if len(params["active-deadline-seconds"]) > 0 {
		activeDeadlineSeconds, err := strconv.ParseInt(params["active-deadline-seconds"], 10, 64)
		if err != nil {
			return err
		}
		podSpec.ActiveDeadlineSeconds = &activeDeadlineSeconds
	}
	return nil
}

// updatePodVolumes updates PodSpec.Volumes and PodSpec.Containers.VolumeMounts with passed parameters.
func updatePodVolumes(volumes []v1.Volume, volumeMounts []v1.VolumeMount, podSpec *v1.PodSpec) error {
	if len(volumes) > 0 && len(volumeMounts) > 0 {
		podSpec.Volumes = volumes
		podSpec.Containers[0].VolumeMounts = volumeMounts
	}

	return nil
}

type BasicPod struct{}

func (BasicPod) ParamNames() []GeneratorParam {
	return []GeneratorParam{
		{"labels", false},
		{"default-name", false},
		{"name", true},
		{"image", true},
		{"image-pull-policy", false},
		{"port", false},
		{"hostport", false},
		{"stdin", false},
		{"leave-stdin-open", false},
		{"tty", false},
		{"restart", false},
		{"command", false},
		{"args", false},
		{"env", false},
		{"requests", false},
		{"limits", false},
		{"serviceaccount", false},
		{"image-pull-secrets", false},
		{"active-deadline-seconds", false},
		{"size", false},
		{"volume", false},
	}
}

func (BasicPod) Generate(genericParams map[string]interface{}) (runtime.Object, error) {
	args, err := getArgs(genericParams)
	if err != nil {
		return nil, err
	}

	envs, err := getEnvs(genericParams)
	if err != nil {
		return nil, err
	}

	volumes, volumeMounts, err := getVolumes(genericParams)
	if err != nil {
		return nil, err
	}

	params, err := getParams(genericParams)
	if err != nil {
		return nil, err
	}

	name, err := getName(params)
	if err != nil {
		return nil, err
	}

	labels, err := getLabels(params, name)
	if err != nil {
		return nil, err
	}

	imagePullSecrets, err := getImagePullSecrets(params)
	if err != nil {
		return nil, err
	}

	stdin, err := GetBool(params, "stdin", false)
	if err != nil {
		return nil, err
	}
	leaveStdinOpen, err := GetBool(params, "leave-stdin-open", false)
	if err != nil {
		return nil, err
	}

	tty, err := GetBool(params, "tty", false)
	if err != nil {
		return nil, err
	}

	resourceRequirements, err := HandleResourceRequirementsV1(params)
	if err != nil {
		return nil, err
	}

	restartPolicy := v1.RestartPolicy(params["restart"])
	if len(restartPolicy) == 0 {
		restartPolicy = v1.RestartPolicyAlways
	}
	// TODO: Figure out why we set ImagePullPolicy here, whether we can make it
	// consistent with the other places imagePullPolicy is set using flag.
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
			Annotations: map[string]string{
				"sh_hyper_instancetype": params["size"],
			},
		},
		Spec: v1.PodSpec{
			ServiceAccountName: params["serviceaccount"],
			Containers: []v1.Container{
				{
					Name:            name,
					Image:           params["image"],
					ImagePullPolicy: v1.PullIfNotPresent,
					Stdin:           stdin,
					StdinOnce:       !leaveStdinOpen && stdin,
					TTY:             tty,
					Resources:       resourceRequirements,
				},
			},
			DNSPolicy:     v1.DNSClusterFirst,
			RestartPolicy: restartPolicy,
		},
	}
	imagePullPolicy := v1.PullPolicy(params["image-pull-policy"])
	if err = updatePodContainers(params, args, envs, imagePullPolicy, &pod.Spec); err != nil {
		return nil, err
	}

	if err := updatePodPorts(params, &pod.Spec); err != nil {
		return nil, err
	}

	if err := updatePodActiveDeadlineSeconds(params, &pod.Spec); err != nil {
		return nil, err
	}

	if err := updatePodImagePullSecrets(imagePullSecrets, &pod.Spec); err != nil {
		return nil, err
	}

	if err := updatePodVolumes(volumes, volumeMounts, &pod.Spec); err != nil {
		return nil, err
	}

	return &pod, nil
}

// parseEnvs converts string into EnvVar objects.
func parseEnvs(envArray []string) ([]v1.EnvVar, error) {
	envs := make([]v1.EnvVar, 0, len(envArray))
	for _, env := range envArray {
		pos := strings.Index(env, "=")
		if pos == -1 {
			return nil, fmt.Errorf("invalid env: %v", env)
		}
		name := env[:pos]
		value := env[pos+1:]
		if len(name) == 0 {
			return nil, fmt.Errorf("invalid env: %v", env)
		}
		if len(validation.IsEnvVarName(name)) != 0 {
			return nil, fmt.Errorf("invalid env: %v", env)
		}
		envVar := v1.EnvVar{Name: name, Value: value}
		envs = append(envs, envVar)
	}
	return envs, nil
}

// parseVolumes converts string into Volume and VolumeMount objects.
func parseVolumes(volumeArray []string) ([]v1.Volume, []v1.VolumeMount, error) {
	volumes := make([]v1.Volume, 0, len(volumeArray))
	volumeMounts := make([]v1.VolumeMount, 0, len(volumeArray))

	for _, str := range volumeArray {
		pos := strings.Index(str, ":")
		if pos == -1 {
			return nil, nil, fmt.Errorf("invalid volume: %v", str)
		}
		name := str[:pos]
		path := str[pos+1:]
		if len(name) == 0 {
			return nil, nil, fmt.Errorf("invalid volume: %v", str)
		}
		volumes = append(volumes, v1.Volume{
			Name: name,
			VolumeSource: v1.VolumeSource{
				FlexVolume: &v1.FlexVolumeSource{
					Options: map[string]string{
						"volumeID": name,
					},
				},
			},
		})
		volumeMounts = append(volumeMounts, v1.VolumeMount{Name: name, MountPath: path})
	}
	return volumes, volumeMounts, nil
}
