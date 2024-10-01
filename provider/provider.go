// Copyright 2016-2023, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/celest-dev/pulumi-turso/provider/internal/tursoclient"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-go-provider/middleware/schema"
	goGen "github.com/pulumi/pulumi/pkg/v3/codegen/go"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"golang.org/x/oauth2"
)

// Version is initialized by the Go linker to contain the semver of this build.
var Version string

const Name string = "turso"

func Provider() p.Provider {
	return infer.Provider(infer.Options{
		Metadata: schema.Metadata{
			DisplayName: "Turso",
			Description: "A Pulumi package for creating and managing Turso resources.",
			Keywords:    []string{"pulumi", "turso", "database", "sqlite", "sqlite3", "libsql"},
			Repository:  "https://github.com/celest-dev/pulumi-turso",
			Publisher:   "Celest",
			LanguageMap: map[string]any{
				"go": goGen.GoPackageInfo{
					RespectSchemaVersion:           true,
					GenerateResourceContainerTypes: true,
					ImportBasePath:                 "github.com/celest-dev/pulumi-turso/sdk/go/turso",
					ModulePath:                     "github.com/celest-dev/pulumi-turso/sdk",
				},
			},
		},
		Resources: []infer.InferredResource{
			infer.Resource[Database](),
			infer.Resource[Group](),
		},
		Config: infer.Config[*Config](),
		ModuleMap: map[tokens.ModuleName]tokens.ModuleName{
			"provider": "index",
		},
	})
}

// Provider-level configuration for the Turso provider.
type Config struct {
	APIToken         *string `pulumi:"apiToken,optional"`
	OrganizationName string  `pulumi:"organization,optional"`

	client *tursoclient.Client
}

var _ infer.CustomConfigure = (*Config)(nil)

func (config *Config) Configure(ctx context.Context) error {
	p.GetLogger(ctx).Info("Configuring Turso provider")
	var apiToken string
	if token := config.APIToken; token != nil {
		apiToken = *token
		p.GetLogger(ctx).Info("Using API token from configuration")
	} else if token := os.Getenv("TURSO_API_TOKEN"); token != "" {
		apiToken = token
		p.GetLogger(ctx).Info("Using API token from environment")
	} else {
		out, err := exec.Command("turso", "auth", "token").Output()
		if err == nil {
			apiToken = strings.TrimSpace(string(out))
			p.GetLogger(ctx).Info("Using Turso CLI authentication")
		}
	}
	if apiToken == "" {
		return errors.New("API token is required or you must be authenticated with Turso CLI")
	}
	if config.OrganizationName != "" {
		p.GetLogger(ctx).Info("Using organization from configuration")
	} else if organization := os.Getenv("TURSO_ORGANIZATION"); organization != "" {
		p.GetLogger(ctx).Info("Using organization from environment")
		config.OrganizationName = organization
	} else {
		return errors.New("organization name is required")
	}
	authClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiToken}))
	client, err := tursoclient.NewClient("https://api.turso.tech", tursoclient.WithClient(authClient))
	if err != nil {
		return fmt.Errorf("failed to create Turso client: %w", err)
	}
	config.client = client
	return nil
}
