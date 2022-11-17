package kube

import (
	"context"
	"errors"
	"fmt"
	"kdiff/internal/helpers"
	"regexp"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type image struct {
	registry string
	name     string
	tag      string
	hash     string
}

type kContainer struct {
	name  string
	image image
}

type AppsV1Resource struct {
	name       string
	containers []kContainer
}

// GetName returns name of resource.
func (a *AppsV1Resource) GetName() string {
	return a.name
}

func (a *AppsV1Resource) GetContainers() []string {
	var containerNames []string
	for _, container := range a.containers {
		containerNames = append(containerNames, container.name)
	}
	return containerNames
}

// GetImage returns the image for a given container name.
func (a *AppsV1Resource) GetImage(containerName string, includeRegistryName bool, includeName bool, includeTag bool, includeHash bool) (string, error) {
	var returnString string

	for _, container := range a.containers {
		if container.name == containerName {
			if includeRegistryName && container.image.registry != "" {
				returnString = container.image.registry
			}
			if includeName {
				if len(returnString) > 0 {
					returnString = returnString + "/"
				}
				returnString = returnString + container.image.name
			}
			if includeTag && container.image.tag != "" {
				if len(returnString) > 0 {
					returnString = returnString + ":"
				}
				returnString = returnString + container.image.tag
			}
			if includeHash && container.image.hash != "" {
				if len(returnString) > 0 {
					returnString = returnString + "@sha256:"
				}
				returnString = returnString + container.image.hash
			}
		}
	}
	if len(returnString) > 0 {
		return returnString, nil
	}
	return returnString, fmt.Errorf("lookup failed for container '%s'", containerName)
}

// GetNamespaces returns a list of namespaces for a give context.
func GetNamespaces(ctx string) []string {
	namespaceList, err := clientSets[ctx].CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	helpers.HandleError(err)

	var listNamespaces []string
	for _, ns := range namespaceList.Items {
		listNamespaces = append(listNamespaces, ns.ObjectMeta.Name)
	}
	return listNamespaces
}

// GetNamespacesForContexts returns a list of namespaces for a give context.
func GetNamespacesForContextList(contexts []string) []string {
	var listNamespaces []string
	for _, ctx := range contexts {
		listNamespaces = append(listNamespaces, GetNamespaces(ctx)...)
	}

	sort.Strings(listNamespaces)
	return helpers.GetUniqueStrings(listNamespaces)
}

// GetDeployments returns a list of deployments for a given context & namespace.
func GetDeployments(ctx, namespace string) []*AppsV1Resource {
	deploymentList, err := clientSets[ctx].AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	helpers.HandleError(err)

	var returnVar []*AppsV1Resource
	for _, deployment := range deploymentList.Items {
		var resource AppsV1Resource
		resource.name = deployment.GetObjectMeta().GetName()
		for _, container := range deployment.Spec.Template.Spec.Containers {
			registryName, imageName, imageTag, imageHash, err := decomposeImage(container.Image)
			helpers.HandleError(err)
			resource.containers = append(resource.containers, kContainer{
				name: container.Name,
				image: image{
					registry: registryName,
					name:     imageName,
					tag:      imageTag,
					hash:     imageHash,
				},
			})
		}
		returnVar = append(returnVar, &resource)
	}
	return returnVar
}

// decomposeImage returns registry, name, tag and hash of a container image.
func decomposeImage(image string) (string, string, string, string, error) {
	re := regexp.MustCompile(`^((.*)/)?(.+?)(:([\w\.\d-]*))?((@sha256:)([\w\d]*?))?$`)
	if matches := re.FindAllStringSubmatch(image, -1); len(matches) >= 0 {
		return matches[0][2], matches[0][3], matches[0][5], matches[0][8], nil
	}
	return "", "", "", "", errors.New("could not parse image")
}

// GetDaemonSets returns a list of daemonSet for a given context & namespace.
func GetDaemonSets(ctx, namespace string) []*AppsV1Resource {
	daemonSetList, err := clientSets[ctx].AppsV1().DaemonSets(namespace).List(context.TODO(), metav1.ListOptions{})
	helpers.HandleError(err)

	var returnVar []*AppsV1Resource
	for _, daemonSet := range daemonSetList.Items {
		var resource AppsV1Resource
		resource.name = daemonSet.GetObjectMeta().GetName()
		for _, container := range daemonSet.Spec.Template.Spec.Containers {
			registryName, imageName, imageTag, imageHash, err := decomposeImage(container.Image)
			helpers.HandleError(err)
			resource.containers = append(resource.containers, kContainer{
				name: container.Name,
				image: image{
					registry: registryName,
					name:     imageName,
					tag:      imageTag,
					hash:     imageHash,
				},
			})
		}
		returnVar = append(returnVar, &resource)
	}
	return returnVar
}

// GetStatefulSets returns a list of statefulSet for a given context & namespace.
func GetStatefulSets(ctx, namespace string) []*AppsV1Resource {
	statefulSetList, err := clientSets[ctx].AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
	helpers.HandleError(err)

	var returnVar []*AppsV1Resource
	for _, statefulSet := range statefulSetList.Items {
		var resource AppsV1Resource
		resource.name = statefulSet.GetObjectMeta().GetName()
		for _, container := range statefulSet.Spec.Template.Spec.Containers {
			registryName, imageName, imageTag, imageHash, err := decomposeImage(container.Image)
			helpers.HandleError(err)
			resource.containers = append(resource.containers, kContainer{
				name: container.Name,
				image: image{
					registry: registryName,
					name:     imageName,
					tag:      imageTag,
					hash:     imageHash,
				},
			})
		}
		returnVar = append(returnVar, &resource)
	}
	return returnVar
}
