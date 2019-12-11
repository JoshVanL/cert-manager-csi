/*
Copyright 2019 The Jetstack cert-manager contributors.

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

package driver

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"

	csiapi "github.com/jetstack/cert-manager-csi/pkg/apis/v1alpha1"
)

func TestValidateNodeServerAttributes(t *testing.T) {
	type vaT struct {
		req      csi.NodePublishVolumeRequest
		expError error
	}

	tests := map[string]vaT{
		"if ephemeral volumes are disabled then error": {
			req: csi.NodePublishVolumeRequest{
				VolumeId:   "target-path",
				TargetPath: "test-namespace",
				VolumeContext: map[string]string{
					csiapi.CSIPodNameKey:           "test-pod",
					csiapi.CSIPodNamespaceKey:      "test-pod",
					"csi.storage.k8s.io/ephemeral": "false",
				},
				VolumeCapability: &csi.VolumeCapability{},
			},
			expError: errors.New("publishing a non-ephemeral volume mount is not supported"),
		},
		"if not volume ID or target path then error": {
			req: csi.NodePublishVolumeRequest{
				VolumeContext: map[string]string{
					csiapi.CSIPodNameKey:      "test-pod",
					csiapi.CSIPodNamespaceKey: "test-namespace",
				},
				VolumeCapability: &csi.VolumeCapability{},
			},
			expError: errors.New("volume ID missing, target path missing"),
		},
		"if no volume capability procided or pod Namespace then error": {
			req: csi.NodePublishVolumeRequest{
				VolumeId:   "volumeID",
				TargetPath: "target-path",
				VolumeContext: map[string]string{
					csiapi.CSIPodNameKey: "test-pod",
				},
				VolumeCapability: nil,
			},
			expError: errors.New(
				"expecting both csi.storage.k8s.io/pod.namespace and csi.storage.k8s.io/pod.name attributes to be set in context, volume capability missing",
			),
		},
		"if block access support added then error": {
			req: csi.NodePublishVolumeRequest{
				VolumeId:   "volumeID",
				TargetPath: "target-path",
				VolumeContext: map[string]string{
					csiapi.CSIPodNameKey:      "test-pod",
					csiapi.CSIPodNamespaceKey: "test-namespace",
				},
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Block{
						Block: &csi.VolumeCapability_BlockVolume{},
					},
				},
			},
			expError: errors.New("block access type not supported"),
		},
		"a request with valid attributes and ephemeral attribute set to 'true' should not error": {
			req: csi.NodePublishVolumeRequest{
				VolumeId:   "volumeID",
				TargetPath: "target-path",
				VolumeContext: map[string]string{
					csiapi.CSIPodNameKey:      "test-pod",
					csiapi.CSIPodNamespaceKey: "test-namespace",

					"csi.storage.k8s.io/ephemeral": "true",
				},
				VolumeCapability: &csi.VolumeCapability{},
			},
			expError: nil,
		},
		"a request with valid attributes and no ephemeral attribute should not error": {
			req: csi.NodePublishVolumeRequest{
				VolumeId:   "volumeID",
				TargetPath: "target-path",
				VolumeContext: map[string]string{
					csiapi.CSIPodNameKey:      "test-pod",
					csiapi.CSIPodNamespaceKey: "test-namespace",
				},
				VolumeCapability: &csi.VolumeCapability{},
			},
			expError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ns := new(NodeServer)
			err := ns.validateVolumeAttributes(&test.req)
			if test.expError == nil {
				if err != nil {
					t.Errorf("unexpected error, got=%s",
						err)
				}

				return
			}

			if err == nil || err.Error() != test.expError.Error() {
				t.Errorf("unexpected error, exp=%s got=%s",
					test.expError, err)
			}
		})
	}
}

func TestCreateDeleteVolume(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(),
		"cert-manager-csi-create-delete-volume")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Error(err)
		}
	}()

	ns := &NodeServer{
		dataRoot: dir,
	}

	id := "test-id"
	targetPath := "test-target-path"
	attr := map[string]string{
		csiapi.CSIPodNameKey:      "test-pod",
		csiapi.CSIPodNamespaceKey: "test-namespace",
	}

	_, err = ns.createVolume(id, targetPath, attr)
	if err != nil {
		t.Error(err)
		return
	}

	path := filepath.Join(dir, "test-id")

	t.Logf("expecting path: %s", path)

	f, err := os.Stat(path)
	if err != nil {
		t.Errorf("expected directory to have been created: %s",
			err)
		return
	}

	if !f.IsDir() {
		t.Errorf("expected volume created to be a directory: %s",
			dir+"/test-id")
		return
	}

}
