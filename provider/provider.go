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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/celest-dev/pulumi-turso/provider/internal/tursoclient"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	goGen "github.com/pulumi/pulumi/pkg/v3/codegen/go"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"golang.org/x/oauth2"
)

// Version is initialized by the Go linker to contain the semver of this build.
var Version string

const Name string = "turso"

func Provider() p.Provider {
	prov, err := infer.NewProviderBuilder().
		WithDisplayName("Turso").
		WithDescription("A Pulumi package for creating and managing Turso resources.").
		WithKeywords("pulumi", "turso", "database", "sqlite", "sqlite3", "libsql", "kind/native").
		WithHomepage("https://github.com/celest-dev/pulumi-turso").
		WithPublisher("celest-dev").
		WithLicense("Apache-2.0").
		WithPluginDownloadURL("github://api.github.com/celest-dev/pulumi-turso").
		WithLanguageMap(map[string]any{
			"go": goGen.GoPackageInfo{
				GenerateResourceContainerTypes: true,
				ImportBasePath:                 "github.com/celest-dev/pulumi-turso/sdk/go/turso",
			},
		}).
		WithNamespace("celest-dev").
		WithResources(
			infer.Resource(&Database{}),
			infer.Resource(&DatabaseToken{}),
			infer.Resource(&Group{}),
			infer.Resource(&GroupToken{}),
		).
		WithConfig(infer.Config(&Config{})).
		WithModuleMap(map[tokens.ModuleName]tokens.ModuleName{
			"provider": "index",
		}).
		Build()
	if err != nil {
		panic(fmt.Errorf("unable to build provider: %w", err))
	}
	return prov
}

// Provider-level configuration for the Turso provider.
type Config struct {
	APIToken         *string `pulumi:"apiToken,optional"`
	OrganizationSlug string  `pulumi:"organization,optional"`

	client *tursoclient.Client
}

var _ infer.CustomConfigure = (*Config)(nil)
var _ infer.Annotated = (*Config)(nil)

func (c *Config) Annotate(a infer.Annotator) {
	a.Describe(&c, "Configuration for the Turso provider.")
	a.Describe(&c.APIToken, "The Turso API token. Can also be set via the TURSO_API_TOKEN environment variable. If not provided, the provider will attempt to use the Turso CLI authentication.")
	a.Describe(&c.OrganizationSlug, "The Turso organization slug. Can also be set via the TURSO_ORGANIZATION environment variable.")
}

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
	if config.OrganizationSlug != "" {
		p.GetLogger(ctx).Info("Using organization from configuration")
	} else if organization := os.Getenv("TURSO_ORGANIZATION"); organization != "" {
		p.GetLogger(ctx).Info("Using organization from environment")
		config.OrganizationSlug = organization
	} else {
		return errors.New("organization name is required")
	}

	// Create an HTTP client that logs error responses from the Turso API
	baseTransport := http.DefaultTransport
	loggingTransport := &errorLoggingTransport{base: baseTransport}

	// Wrap with OAuth2 authentication
	authClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiToken}))
	authClient.Transport = &oauth2.Transport{
		Source: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiToken}),
		Base:   loggingTransport,
	}

	client, err := tursoclient.NewClient("https://api.turso.tech", tursoclient.WithClient(authClient))
	if err != nil {
		return fmt.Errorf("failed to create Turso client: %w", err)
	}
	config.client = client
	return nil
}

// errorLoggingTransport is an http.RoundTripper that logs error responses (4xx/5xx)
// from the Turso API with full response body details.
type errorLoggingTransport struct {
	base http.RoundTripper
}

func (t *errorLoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Log error responses (4xx and 5xx status codes)
	if resp.StatusCode >= 400 {
		// Read the response body
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()

		if readErr != nil {
			// If we can't read the body, log what we can
			slog.Error("Turso API error",
				"method", req.Method,
				"path", req.URL.Path,
				"status", resp.StatusCode,
				"error", "failed to read response body",
				"readError", readErr,
			)
		} else {
			// Log the full error details
			slog.Error("Turso API error",
				"method", req.Method,
				"path", req.URL.Path,
				"status", resp.StatusCode,
				"body", string(body),
			)
		}

		// Restore the body so it can be read again by the caller
		resp.Body = io.NopCloser(bytes.NewReader(body))
	}

	return resp, nil
}
