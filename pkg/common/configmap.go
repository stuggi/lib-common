/*
Copyright 2020 Red Hat

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

package common

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
)

// createOrPatchConfigMap -
func createOrPatchConfigMap(
	ctx context.Context,
	r ReconcilerCommon,
	obj client.Object,
	cm Template,
) (string, controllerutil.OperationResult, error) {
	data := make(map[string]string)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cm.Name,
			Namespace:   cm.Namespace,
			Annotations: cm.Annotations,
		},
		Data: data,
	}

	// create or update the CM
	op, err := controllerutil.CreateOrPatch(ctx, r.GetClient(), configMap, func() error {

		configMap.Labels = cm.Labels
		// add data from templates
		renderedTemplateData, err := GetTemplateData(cm)
		if err != nil {
			return err
		}
		configMap.Data = renderedTemplateData
		// add provided custom data to configMap.Data
		// Note: this can overwrite data rendered from GetTemplateData() if key is same
		if len(cm.CustomData) > 0 {
			for k, v := range cm.CustomData {
				configMap.Data[k] = v
			}
		}

		if !cm.SkipSetOwner {
			err := controllerutil.SetControllerReference(obj, configMap, r.GetScheme())
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return "", op, fmt.Errorf("error create/updating configmap: %v", err)
	}

	configMapHash, err := ObjectHash(configMap)
	if err != nil {
		return "", op, fmt.Errorf("error calculating configuration hash: %v", err)
	}

	return configMapHash, op, nil
}

// createOrGetCustomConfigMap -
func createOrGetCustomConfigMap(
	ctx context.Context,
	r ReconcilerCommon,
	obj client.Object,
	cm Template,
) (string, error) {
	// Check if this configMap already exists
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cm.Name,
			Namespace:   cm.Namespace,
			Labels:      cm.Labels,
			Annotations: cm.Annotations,
		},
		Data: map[string]string{},
	}
	foundConfigMap := &corev1.ConfigMap{}
	err := r.GetClient().Get(ctx, types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, foundConfigMap)
	if err != nil && k8s_errors.IsNotFound(err) {
		if !cm.SkipSetOwner {
			err := controllerutil.SetControllerReference(obj, configMap, r.GetScheme())
			if err != nil {
				return "", err
			}
		}

		r.GetLogger().Info(fmt.Sprintf("Creating a new ConfigMap %s in namespace %s", cm.Namespace, cm.Name))
		err = r.GetClient().Create(ctx, configMap)
		if err != nil {
			return "", err
		}
	} else {
		// use data from already existing custom configmap
		configMap.Data = foundConfigMap.Data
	}

	configMapHash, err := ObjectHash(configMap)
	if err != nil {
		return "", fmt.Errorf("error calculating configuration hash: %v", err)
	}

	return configMapHash, nil
}

// EnsureConfigMaps - get all configmaps required, verify they exist and add the hash to env and status
func EnsureConfigMaps(
	ctx context.Context,
	r ReconcilerCommon,
	obj client.Object,
	cms []Template,
	envVars *map[string]EnvSetter,
) error {
	var err error

	for _, cm := range cms {
		var hash string
		var op controllerutil.OperationResult

		if cm.Type != TemplateTypeCustom {
			hash, op, err = createOrPatchConfigMap(ctx, r, obj, cm)
		} else {
			hash, err = createOrGetCustomConfigMap(ctx, r, obj, cm)
			// set op to OperationResultNone because createOrGetCustomConfigMap does not return an op
			// and it will add log entries bellow with none operation
			op = controllerutil.OperationResult(controllerutil.OperationResultNone)
		}
		if err != nil {
			return err
		}
		if op != controllerutil.OperationResultNone {
			r.GetLogger().Info(fmt.Sprintf("ConfigMap %s successfully reconciled - operation: %s", cm.Name, string(op)))
		}
		if envVars != nil {
			(*envVars)[cm.Name] = EnvValue(hash)
		}
	}

	return nil
}

// GetConfigMaps - get all configmaps required, verify they exist and add the hash to env and status
func GetConfigMaps(
	ctx context.Context,
	r ReconcilerCommon,
	obj client.Object,
	configMaps []string,
	namespace string,
	envVars *map[string]EnvSetter,
) ([]Hash, error) {
	hashes := []Hash{}

	for _, cm := range configMaps {
		_, hash, err := GetConfigMapAndHashWithName(ctx, r, cm, namespace)
		if err != nil {
			return nil, err
		}
		(*envVars)[cm] = EnvValue(hash)
		hashes = append(hashes, Hash{Name: cm, Hash: hash})
	}

	return hashes, nil
}

// CreateOrGetCustomConfigMap -
func CreateOrGetCustomConfigMap(
	ctx context.Context,
	r ReconcilerCommon,
	configMap *corev1.ConfigMap,
) (string, error) {
	// Check if this configMap already exists
	foundConfigMap := &corev1.ConfigMap{}
	err := r.GetClient().Get(ctx, types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, foundConfigMap)
	if err != nil && k8s_errors.IsNotFound(err) {
		r.GetLogger().Info("Creating a new ConfigMap", "ConfigMap.Namespace", configMap.Namespace, "ConfigMap.Name", configMap.Name)
		err = r.GetClient().Create(ctx, configMap)
		if err != nil {
			return "", err
		}
	} else {
		// use data from already existing custom configmap
		configMap.Data = foundConfigMap.Data
	}

	configMapHash, err := ObjectHash(configMap)
	if err != nil {
		return "", fmt.Errorf("error calculating configuration hash: %v", err)
	}

	return configMapHash, nil
}

// GetConfigMapAndHashWithName -
func GetConfigMapAndHashWithName(
	ctx context.Context,
	r ReconcilerCommon,
	configMapName string,
	namespace string,
) (*corev1.ConfigMap, string, error) {

	configMap := &corev1.ConfigMap{}
	err := r.GetClient().Get(ctx, types.NamespacedName{Name: configMapName, Namespace: namespace}, configMap)
	if err != nil && k8s_errors.IsNotFound(err) {
		r.GetLogger().Error(err, configMapName+" ConfigMap not found!", "Instance.Namespace", namespace, "ConfigMap.Name", configMapName)
		return configMap, "", err
	}
	configMapHash, err := ObjectHash(configMap)
	if err != nil {
		return configMap, "", fmt.Errorf("error calculating configuration hash: %v", err)
	}
	return configMap, configMapHash, nil
}

//
// GetConfigMap - Get config map
//
// if the config map is not found, requeue after requeueTimeout in seconds
func GetConfigMap(
	ctx context.Context,
	r ReconcilerCommon,
	object client.Object,
	configMapName string,
	requeueTimeout int,
) (*corev1.ConfigMap, ctrl.Result, error) {

	configMap := &corev1.ConfigMap{}
	err := r.GetClient().Get(ctx, types.NamespacedName{Name: configMapName, Namespace: object.GetNamespace()}, configMap)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			msg := fmt.Sprintf("%s config map does not exist: %v", configMapName, err)
			LogForObject(r, msg, object)

			return configMap, ctrl.Result{RequeueAfter: time.Duration(requeueTimeout) * time.Second}, nil
		}
		msg := fmt.Sprintf("Error getting %s config map: %v", configMapName, err)
		err = WrapErrorForObject(msg, object, err)

		return configMap, ctrl.Result{}, err
	}

	return configMap, ctrl.Result{}, nil
}
