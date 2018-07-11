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

package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/hyperhq/pi/pkg/pi"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/resource"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"

	"github.com/docker/distribution/reference"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	api "k8s.io/kubernetes/pkg/apis/core"
	coreclient "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/core/internalversion"
	conditions "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/interrupt"
)

const (
	defaultPodAttachTimeout = 60 * time.Second
	defaultPodLogsTimeout   = 20 * time.Second
)

var (
	runLong = templates.LongDesc(i18n.T(`
		Create and run a pod with particular image.`))

	runExample = templates.Examples(i18n.T(`
		# Start a single instance of nginx.
		pi run nginx --image=nginx

		# Start a single instance of nginx and set environment variables "DNS_DOMAIN=cluster" and "POD_NAMESPACE=default" in the container.
		pi run nginx --image=nginx --env="DNS_DOMAIN=cluster" --env="POD_NAMESPACE=default"

		# Start a single instance of nginx and set labels "app=nginx" and "env=prod" in the container.
		pi run nginx --image=nginx --labels="app=nginx,env=prod"

		# Start a pod of busybox and keep it in the foreground, don't restart it if it exits.
		pi run -it busybox --image=busybox --restart=Never -- sh

		# Start the nginx container using a specified command and custom arguments.
		pi run nginx --image=nginx -- <cmd> <arg1> ... <argN>

		# Start the nginx container using a specified command and custom arguments.
		pi run nginx --rm --image=nginx -- echo hello world`))
)

type RunObject struct {
	Object  runtime.Object
	Kind    string
	Mapper  meta.RESTMapper
	Mapping *meta.RESTMapping
}

func NewCmdRun(f cmdutil.Factory, cmdIn io.Reader, cmdOut, cmdErr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run NAME [-i] [-t] --image=image [--env=\"key=value\"] -- [COMMAND] [args...]",
		Short:   i18n.T("Run a pod with particular image."),
		Long:    runLong,
		Example: runExample,
		Run: func(cmd *cobra.Command, args []string) {
			argsLenAtDash := cmd.ArgsLenAtDash()
			err := RunRun(f, cmdIn, cmdOut, cmdErr, cmd, args, argsLenAtDash)
			cmdutil.CheckErr(err)
		},
	}
	//cmdutil.AddPrinterFlags(cmd)
	AddRunFlags(cmd)
	//cmdutil.AddApplyAnnotationFlags(cmd)
	//cmdutil.AddRecordFlag(cmd)
	//cmdutil.AddInclude3rdPartyFlags(cmd)
	//cmdutil.AddPodRunningTimeoutFlag(cmd, defaultPodAttachTimeout)
	return cmd
}

func AddRunFlags(cmd *cobra.Command) {
	//cmdutil.AddDryRunFlag(cmd)
	cmd.Flags().String("generator", "", i18n.T("The name of the API generator to use, see http://kubernetes.io/docs/user-guide/pi-conventions/#generators for a list."))
	cmd.Flags().String("image", "", i18n.T("The image for the container to run."))
	cmd.MarkFlagRequired("image")
	cmd.Flags().String("image-pull-policy", "", i18n.T("The image pull policy for the container. If left empty, this value will not be specified by the client and defaulted by the server"))
	//cmd.Flags().IntP("replicas", "r", 1, "Number of replicas to create for this container. Default is 1.")
	cmd.Flags().Bool("rm", false, "If true, delete resources created in this command for attached containers.")
	//cmd.Flags().String("overrides", "", i18n.T("An inline JSON override for the generated object. If this is non-empty, it is used to override the generated object. Requires that the object supply a valid apiVersion field."))
	cmd.Flags().StringArray("env", []string{}, "Environment variables to set in the container")
	//cmd.Flags().String("serviceaccount", "", "Service account to set in the pod spec")
	//cmd.Flags().String("port", "", i18n.T("The port that this container exposes.  If --expose is true, this is also the port used by the service that is created."))
	//cmd.Flags().Int("hostport", -1, "The host port mapping for the container port. To demonstrate a single-machine container.")
	cmd.Flags().StringP("labels", "l", "", "Comma separated labels to apply to the pod(s). Will override previous values.")
	cmd.Flags().BoolP("stdin", "i", false, "Keep stdin open on the container(s) in the pod, even if nothing is attached.")
	cmd.Flags().BoolP("tty", "t", false, "Allocated a TTY for each container in the pod.")
	//cmd.Flags().Bool("attach", false, "If true, wait for the Pod to start running, and then attach to the Pod as if 'pi attach ...' were called.  Default false, unless '-i/--stdin' is set, in which case the default is true. With '--restart=Never' the exit code of the container process is returned.")
	//cmd.Flags().Bool("leave-stdin-open", false, "If the pod is started in interactive mode or with stdin, leave stdin open after the first attach completes. By default, stdin will be closed after the first attach completes.")
	cmd.Flags().String("restart", "Always", i18n.T("The restart policy for this Pod.  Legal values [Always, OnFailure, Never]. if set to 'Never', a regular pod is created. Default 'Always'"))
	//cmd.Flags().Bool("command", false, "If true and extra arguments are present, use them as the 'command' field in the container, rather than the 'args' field which is the default.")
	//cmd.Flags().String("requests", "", i18n.T("The resource requirement requests for this container.  For example, 'cpu=100m,memory=256Mi'.  Note that server side components may assign requests depending on the server configuration, such as limit ranges."))
	cmd.Flags().String("limits", "", i18n.T("The resource requirement limits for this container.  For example, 'cpu=200m,memory=512Mi'.  Note that server side components may assign limits depending on the server configuration, such as limit ranges."))
	//cmd.Flags().Bool("expose", false, "If true, a public, external service is created for the container(s) which are run")
	//cmd.Flags().String("service-generator", "service/v2", i18n.T("The name of the generator to use for creating a service.  Only used if --expose is true"))
	//cmd.Flags().String("service-overrides", "", i18n.T("An inline JSON override for the generated service object. If this is non-empty, it is used to override the generated object. Requires that the object supply a valid apiVersion field.  Only used if --expose is true."))
	//cmd.Flags().Bool("quiet", false, "If true, suppress prompt messages.")
	//cmd.Flags().String("schedule", "", i18n.T("A schedule in the Cron format the job should be run with."))
	cmd.Flags().StringP("image-pull-secrets", "", "", i18n.T("The secret for the private docker registry, comma separated."))
	cmd.Flags().StringP("active-deadline-seconds", "", "", i18n.T("Optional duration in seconds the pod may be active on the node relative to StartTime before the system will actively try to mark it failed and kill associated containers. Value must be a positive integer."))
	cmd.Flags().StringP("size", "", "s4", i18n.T("The size for the pod (e.g. s1, s2, s3, s4, m1, m2, m3, l1, l2, l3, l4, l5, l6)"))
	cmd.Flags().StringArray("volume", []string{}, "Pod volumes to mount into the container's filesystem. format '<volname>:<path>'")
}

func RunRun(f cmdutil.Factory, cmdIn io.Reader, cmdOut, cmdErr io.Writer, cmd *cobra.Command, args []string, argsLenAtDash int) error {
	// Let pi run follow rules for `--`, see #13004 issue
	if len(args) == 0 || argsLenAtDash == 0 {
		return cmdutil.UsageErrorf(cmd, "NAME is required for run")
	}

	// validate image name
	imageName := cmdutil.GetFlagString(cmd, "image")
	if imageName == "" {
		return fmt.Errorf("--image is required")
	}
	validImageRef := reference.ReferenceRegexp.MatchString(imageName)
	if !validImageRef {
		return fmt.Errorf("Invalid image name %q: %v", imageName, reference.ErrReferenceInvalidFormat)
	}

	interactive := cmdutil.GetFlagBool(cmd, "stdin")
	tty := cmdutil.GetFlagBool(cmd, "tty")
	if tty && !interactive {
		return cmdutil.UsageErrorf(cmd, "-i/--stdin is required for containers with -t/--tty=true")
	}

	namespace, _, err := f.DefaultNamespace()
	if err != nil {
		return err
	}
	restartPolicy, err := getRestartPolicy(cmd, interactive)
	if err != nil {
		return err
	}

	remove := cmdutil.GetFlagBool(cmd, "rm")

	if err := verifyImagePullPolicy(cmd); err != nil {
		return err
	}

	clientset, err := f.ClientSet()
	if err != nil {
		return err
	}

	generatorName := cmdutil.GetFlagString(cmd, "generator")
	//generatorName := cmdutil.RunPodV1GeneratorName

	//schedule := cmdutil.GetFlagString(cmd, "schedule")
	schedule := ""

	if len(schedule) != 0 && len(generatorName) == 0 {
		hasResource, err := cmdutil.HasResource(clientset.Discovery(), batchv1beta1.SchemeGroupVersion.WithResource("cronjobs"))
		if err != nil {
			return err
		}
		if hasResource {
			generatorName = cmdutil.CronJobV1Beta1GeneratorName
		} else {
			generatorName = cmdutil.CronJobV2Alpha1GeneratorName
		}
	}
	if len(generatorName) == 0 {
		switch restartPolicy {
		case api.RestartPolicyAlways:
			// TODO: we need to deprecate this along with extensions/v1beta1.Deployments
			// in favor of the new generator for apps/v1beta1.Deployments
			hasResource, err := cmdutil.HasResource(clientset.Discovery(), extensionsv1beta1.SchemeGroupVersion.WithResource("deployments"))
			if err != nil {
				return err
			}
			if hasResource {
				generatorName = cmdutil.DeploymentV1Beta1GeneratorName
			} else {
				generatorName = cmdutil.RunV1GeneratorName
			}
			generatorName = cmdutil.RunPodV1GeneratorName
		case api.RestartPolicyOnFailure:
			hasResource, err := cmdutil.HasResource(clientset.Discovery(), batchv1.SchemeGroupVersion.WithResource("jobs"))
			if err != nil {
				return err
			}
			if hasResource {
				generatorName = cmdutil.JobV1GeneratorName
			} else {
				generatorName = cmdutil.RunPodV1GeneratorName
			}
		case api.RestartPolicyNever:
			generatorName = cmdutil.RunPodV1GeneratorName
		}
	}

	generatorName, err = cmdutil.FallbackGeneratorNameIfNecessary(generatorName, clientset.Discovery(), cmdErr)
	if err != nil {
		return err
	}

	generators := f.Generators("run")
	if generatorName == cmdutil.JobV1GeneratorName {
		generators = f.Generators("job")
		if tty || interactive {
			return fmt.Errorf("--tty and --stdin are not supported for job")
		}
	}

	//fmt.Printf("generatorName:%v generators:%v\n", generatorName, generators)
	generator, found := generators[generatorName]
	if !found {
		return cmdutil.UsageErrorf(cmd, "generator %q not found", generatorName)
	}
	names := generator.ParamNames()
	params := pi.MakeParams(cmd, names)
	params["name"] = args[0]
	if len(args) > 1 {
		params["args"] = args[1:]
	}
	command, err := getArgs(params)
	if err != nil {
		return err
	}

	params["limits"] = cmdutil.GetFlagString(cmd, "limits")
	params["size"] = cmdutil.GetFlagString(cmd, "size")
	if params["size"] != "" && params["limits"] != "" {
		return cmdutil.UsageErrorf(cmd, "--size and --limits can not be used together")
	}

	params["env"] = cmdutil.GetFlagStringArray(cmd, "env")
	params["image-pull-secrets"] = cmdutil.GetFlagString(cmd, "image-pull-secrets")
	params["active-deadline-seconds"] = cmdutil.GetFlagString(cmd, "active-deadline-seconds")

	params["volume"] = cmdutil.GetFlagStringArray(cmd, "volume")

	podClient := clientset.Core()
	podName := params["name"].(string)

	var runObjectMap = map[string]*RunObject{}
	runObject, err := createGeneratedObject(f, cmd, generator, names, params, "", namespace)
	if err != nil {
		return err
	}
	runObjectMap[generatorName] = runObject

	if len(command) > 0 {
		pod, err := podClient.Pods("default").Get(podName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if remove {
			defer deletePod(pod.Name, podClient)
		}

		if pod.Status.Phase == api.PodSucceeded || pod.Status.Phase == api.PodFailed {
			return fmt.Errorf("cannot exec into a container in a completed pod; current phase is %s", pod.Status.Phase)
		}
		for i := 0; i <= 20; i++ {
			if pod.Status.Phase == api.PodPending {
				glog.V(4).Infof("%v/20 waiting for pod start", i)
				time.Sleep(time.Duration(1 * time.Second))
				pod, err = podClient.Pods("default").Get(podName, metav1.GetOptions{})
			} else {
				glog.V(4).Infof("pod started:%v", string(pod.Status.Phase))
				break
			}
		}
		options := &ExecOptions{
			StreamOptions: StreamOptions{
				In:        cmdIn,
				Out:       cmdOut,
				Err:       cmdErr,
				PodName:   pod.Name,
				Namespace: "default",
				Quiet:     false,
				TTY:       tty,
				Stdin:     interactive,
			},
			Executor: &DefaultRemoteExecutor{},
			Command:  command,
		}
		options.PodClient = podClient
		cmdutil.CheckErr(options.RunHyper(f))
	} else {
		f.PrintSuccess(runObject.Mapper, false, cmdOut, runObject.Mapping.Resource, args[0], false, "created")
	}

	return nil
}

func deletePod(podName string, podClient coreclient.CoreInterface) {
	glog.V(4).Infof("deletel pod %v due to --rm", podName)
	var gracePeriodSeconds int64 = 0
	err := podClient.Pods("default").Delete(podName, &metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds})
	if err != nil {
		fmt.Printf("failed to delete pod \"%v\", error:%v\n", podName, err)
	} else {
		fmt.Printf("pod \"%v\" deleted\n", podName)
	}
}

// waitForPod watches the given pod until the exitCondition is true
func waitForPod(podClient coreclient.PodsGetter, ns, name string, exitCondition watch.ConditionFunc) (*api.Pod, error) {
	w, err := podClient.Pods(ns).Watch(metav1.SingleObject(metav1.ObjectMeta{Name: name}))
	if err != nil {
		return nil, err
	}

	intr := interrupt.New(nil, w.Stop)
	var result *api.Pod
	err = intr.Run(func() error {
		ev, err := watch.Until(0, w, func(ev watch.Event) (bool, error) {
			return exitCondition(ev)
		})
		if ev != nil {
			result = ev.Object.(*api.Pod)
		}
		return err
	})

	// Fix generic not found error.
	if err != nil && errors.IsNotFound(err) {
		err = errors.NewNotFound(api.Resource("pods"), name)
	}

	return result, err
}

func handleAttachPod(f cmdutil.Factory, podClient coreclient.PodsGetter, ns, name string, opts *AttachOptions) error {
	pod, err := waitForPod(podClient, ns, name, conditions.PodRunningAndReady)
	if err != nil && err != conditions.ErrPodCompleted {
		return err
	}

	if pod.Status.Phase == api.PodSucceeded || pod.Status.Phase == api.PodFailed {
		return logOpts(f, pod, opts)
	}

	opts.PodClient = podClient
	opts.PodName = name
	opts.Namespace = ns

	// TODO: opts.Run sets opts.Err to nil, we need to find a better way
	stderr := opts.Err
	if err := opts.Run(); err != nil {
		fmt.Fprintf(stderr, "Error attaching, falling back to logs: %v\n", err)
		return logOpts(f, pod, opts)
	}
	return nil
}

// logOpts logs output from opts to the pods log.
func logOpts(f cmdutil.Factory, pod *api.Pod, opts *AttachOptions) error {
	ctrName, err := opts.GetContainerName(pod)
	if err != nil {
		return err
	}

	req, err := f.LogsForObject(pod, &api.PodLogOptions{Container: ctrName}, opts.GetPodTimeout)
	if err != nil {
		return err
	}

	readCloser, err := req.Stream()
	if err != nil {
		return err
	}
	defer readCloser.Close()

	_, err = io.Copy(opts.Out, readCloser)
	if err != nil {
		return err
	}
	return nil
}

func getRestartPolicy(cmd *cobra.Command, interactive bool) (api.RestartPolicy, error) {
	restart := cmdutil.GetFlagString(cmd, "restart")
	if len(restart) == 0 {
		if interactive {
			return api.RestartPolicyOnFailure, nil
		} else {
			return api.RestartPolicyAlways, nil
		}
	}
	switch api.RestartPolicy(restart) {
	case api.RestartPolicyAlways:
		return api.RestartPolicyAlways, nil
	case api.RestartPolicyOnFailure:
		return api.RestartPolicyOnFailure, nil
	case api.RestartPolicyNever:
		return api.RestartPolicyNever, nil
	}
	return "", cmdutil.UsageErrorf(cmd, "invalid restart policy: %s", restart)
}

func verifyImagePullPolicy(cmd *cobra.Command) error {
	pullPolicy := cmdutil.GetFlagString(cmd, "image-pull-policy")
	switch api.PullPolicy(pullPolicy) {
	case api.PullAlways, api.PullIfNotPresent, api.PullNever:
		return nil
	case "":
		return nil
	}
	return cmdutil.UsageErrorf(cmd, "invalid image pull policy: %s", pullPolicy)
}

func createGeneratedObject(f cmdutil.Factory, cmd *cobra.Command, generator pi.Generator, names []pi.GeneratorParam, params map[string]interface{}, overrides, namespace string) (*RunObject, error) {
	err := pi.ValidateParams(names, params)
	if err != nil {
		return nil, err
	}

	// TODO: Validate flag usage against selected generator. More tricky since --expose was added.
	obj, err := generator.Generate(params)
	if err != nil {
		return nil, err
	}

	mapper, typer := f.Object()
	groupVersionKinds, _, err := typer.ObjectKinds(obj)
	if err != nil {
		return nil, err
	}
	groupVersionKind := groupVersionKinds[0]

	if len(overrides) > 0 {
		codec := runtime.NewCodec(f.JSONEncoder(), f.Decoder(true))
		obj, err = cmdutil.Merge(codec, obj, overrides)
		if err != nil {
			return nil, err
		}
	}

	mapping, err := mapper.RESTMapping(groupVersionKind.GroupKind(), groupVersionKind.Version)
	if err != nil {
		return nil, err
	}
	client, err := f.ClientForMapping(mapping)
	if err != nil {
		return nil, err
	}

	resourceMapper := &resource.Mapper{
		ObjectTyper:  typer,
		RESTMapper:   mapper,
		ClientMapper: resource.ClientMapperFunc(f.ClientForMapping),
		Decoder:      f.Decoder(true),
	}
	info, err := resourceMapper.InfoForObject(obj, nil)
	if err != nil {
		return nil, err
	}

	obj, err = resource.NewHelper(client, mapping).Create(namespace, false, info.Object)
	if err != nil {
		return nil, err
	}

	return &RunObject{
		Object:  obj,
		Kind:    groupVersionKind.Kind,
		Mapper:  mapper,
		Mapping: mapping,
	}, nil
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
