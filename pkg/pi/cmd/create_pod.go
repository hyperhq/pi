/*
Copyright 2015 The Kubernetes Authors.

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

	"github.com/docker/distribution/reference"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/kubernetes/pkg/apis/core"

	"github.com/hyperhq/pi/pkg/pi"
	"github.com/hyperhq/pi/pkg/pi/cmd/templates"
	cmdutil "github.com/hyperhq/pi/pkg/pi/cmd/util"
	"github.com/hyperhq/pi/pkg/pi/util/i18n"
)

var (
	createPodLong = templates.LongDesc(i18n.T(`Create and run a pod with particular image.`))

	createPodExample = templates.Examples(i18n.T(`
		# Start a single instance of nginx.
		pi create pod nginx --image=nginx

		# Start a single instance of nginx and set environment variables "DNS_DOMAIN=cluster" and "POD_NAMESPACE=default" in the container.
		pi create pod nginx --image=nginx --env="DNS_DOMAIN=cluster" --env="POD_NAMESPACE=default"

		# Start a single instance of nginx and set labels "app=nginx" and "env=prod" in the container.
		pi create pod nginx --image=nginx --labels="app=nginx,env=prod"

		# Start a pod of busybox and keep it in the foreground, don't restart it if it exits.
		pi create pod -it busybox --image=busybox --restart=Never -- sh

		# Start the nginx container using a specified command and custom arguments.
		pi create pod nginx --image=nginx -- <cmd> <arg1> ... <argN>

		# Start the nginx container using a specified command and custom arguments.
		pi create pod nginx --rm --image=nginx -- echo hello world`))
)

// NewCmdCreatePod groups subcommands to create pod
func NewCmdCreatePod(f cmdutil.Factory, cmdOut, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pod NAME --image=image [options] [COMMAND]",
		Short:   i18n.T("Create and run a pod with particular image."),
		Long:    createPodLong,
		Example: createPodExample,
		Run: func(cmd *cobra.Command, args []string) {
			argsLenAtDash := cmd.ArgsLenAtDash()
			err := RunCreatePod(f, cmdOut, errOut, cmd, args, argsLenAtDash)
			// RunRun(f, cmdOut, cmdErr, cmd, args, argsLenAtDash)
			cmdutil.CheckErr(err)
		},
	}

	addFlags(cmd)

	return cmd
}

func addFlags(cmd *cobra.Command) {
	//cmdutil.AddDryRunFlag(cmd)
	//cmd.Flags().String("generator", "", i18n.T("The name of the API generator to use, see http://kubernetes.io/docs/user-guide/pi-conventions/#generators for a list."))
	cmd.Flags().String("image", "", i18n.T("The image for the container to run."))
	cmd.MarkFlagRequired("image")
	//cmd.Flags().String("image-pull-policy", "", i18n.T("The image pull policy for the container. If left empty, this value will not be specified by the client and defaulted by the server"))
	//cmd.Flags().IntP("replicas", "r", 1, "Number of replicas to create for this container. Default is 1.")
	cmd.Flags().Bool("rm", false, "If true, delete resources created in this command for attached containers.")
	//cmd.Flags().String("overrides", "", i18n.T("An inline JSON override for the generated object. If this is non-empty, it is used to override the generated object. Requires that the object supply a valid apiVersion field."))
	cmd.Flags().StringArray("env", []string{}, "Environment variables to set in the container")
	//cmd.Flags().String("serviceaccount", "", "Service account to set in the pod spec")
	//cmd.Flags().String("port", "", i18n.T("The port that this container exposes.  If --expose is true, this is also the port used by the service that is created."))
	//cmd.Flags().Int("hostport", -1, "The host port mapping for the container port. To demonstrate a single-machine container.")
	cmd.Flags().StringP("labels", "l", "", "Comma separated labels to apply to the pod(s). Will override previous values.")
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
	cmd.Flags().StringArray("volume", []string{}, "Pod volumes to mount into the container's filesystem.")
}

// Run create-pod
func RunCreatePod(f cmdutil.Factory, cmdOut, cmdErr io.Writer, cmd *cobra.Command, args []string, argsLenAtDash int) error {
	// Let pi create pod follow rules for `--`, see #13004 issue
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

	namespace, _, err := f.DefaultNamespace()
	if err != nil {
		return err
	}
	restartPolicy, err := getRestartPolicy(cmd, false)
	if err != nil {
		return err
	}

	//if err := verifyImagePullPolicy(cmd); err != nil {
	//	return err
	//}

	clientset, err := f.ClientSet()
	if err != nil {
		return err
	}

	//generatorName := cmdutil.GetFlagString(cmd, "generator")
	generatorName := cmdutil.RunPodV1GeneratorName

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

	params["env"] = cmdutil.GetFlagStringArray(cmd, "env")
	params["image-pull-secrets"] = cmdutil.GetFlagString(cmd, "image-pull-secrets")
	params["active-deadline-seconds"] = cmdutil.GetFlagString(cmd, "active-deadline-seconds")
	params["size"] = cmdutil.GetFlagString(cmd, "size")
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
				Out:       cmdOut,
				Err:       cmdErr,
				PodName:   pod.Name,
				Namespace: "default",
				Quiet:     false,
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
