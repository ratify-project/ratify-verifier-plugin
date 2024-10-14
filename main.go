package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ratify-project/ratify/pkg/common"
	"github.com/ratify-project/ratify/pkg/ocispecs"
	"github.com/ratify-project/ratify/pkg/referrerstore"
	"github.com/ratify-project/ratify/pkg/verifier"
	"github.com/ratify-project/ratify/pkg/verifier/plugin/skel"

	// Imports are required to utilize built-in referrer stores
	_ "github.com/ratify-project/ratify/pkg/referrerstore/oras"
)

// These values are used in Ratify configuration to identify the plugin
const (
	pluginName    = "sample"
	pluginVersion = "1.0.0"
)

// Configuration specific to your plugin is defined here
// This is what determines the verifier properties available in your Ratify configuration
type PluginConfig struct {
	Name            string   `json:"name"`
	AllowedPrefixes []string `json:"allowedPrefixes"`
}

// Used to unwrap the envelope of data passed to the plugin via STDIN
type PluginInputConfig struct {
	Config PluginConfig `json:"config"`
}

// Use the plugin skeleton provided by Ratify
func main() {
	skel.PluginMain(pluginName, pluginVersion, VerifyReference, []string{pluginVersion})
}

// Given an input subject, determine whether it is considered successfully verified
func VerifyReference(args *skel.CmdArgs, subjectReference common.Reference, referenceDescriptor ocispecs.ReferenceDescriptor, referrerStore referrerstore.ReferrerStore) (*verifier.VerifierResult, error) {

	// Parse the configuration from STDIN
	inputConf := PluginInputConfig{}
	if err := json.Unmarshal(args.StdinData, &inputConf); err != nil {
		return nil, fmt.Errorf("failed to parse stdin for the input: %v", err)
	}
	config := inputConf.Config

	// sample verification: check if the subject reference is allowed against the list of allowed prefixes
	isSuccess := false
	message := fmt.Sprintf("Sample verification failure: subject did not begin with any of the allowed prefixes: %v", config.AllowedPrefixes)
	for _, prefix := range config.AllowedPrefixes {
		if strings.HasPrefix(subjectReference.Original, prefix) {
			isSuccess = true
			message = "Sample verification success"
			break
		}
	}

	// sample usage of referrer store: get the reference manifest
	referenceManifest, err := referrerStore.GetReferenceManifest(context.TODO(), subjectReference, referenceDescriptor)
	if err != nil {
		return nil, err
	}

	// sample usage of referrer store: get the length of the first blob
	blobData, err := referrerStore.GetBlobContent(context.TODO(), subjectReference, referenceManifest.Blobs[0].Digest)
	if err != nil {
		return nil, err
	}

	return &verifier.VerifierResult{
		Name:      config.Name,
		IsSuccess: isSuccess,
		Message:   message,

		// You can optionally include extension data for processing by downstream consumers (ex: Rego policies)
		Extensions: map[string]interface{}{
			"hello":          "world",
			"firstBlobBytes": len(blobData),
		},
	}, nil
}
