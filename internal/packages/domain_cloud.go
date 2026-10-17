package packages

import "github.com/KennethanCeyer/coffee-bun/internal/model"

/*
Domain: Cloud & DevOps CLIs
Description:
    Core multi-cloud command-line interfaces for managing modern infrastructure,
    deployments, Kubernetes clusters, and cloud-native services across macOS, Linux, and Windows.

Components:
    1. AWS CLI: Amazon Web Services command-line interface.
    2. Google Cloud SDK: GCP tools including gcloud, gsutil, and bq.
    3. Azure CLI: Microsoft Azure cloud management tool (az).
*/

// CloudGroup defines the UI presentation and grouping metadata for cloud tools
var CloudGroup = model.CustomGroup{
	Category:    model.CategoryCloud,
	Title:       "Cloud & DevOps CLIs (AWS, Google Cloud, Azure)",
	Description: "AWS CLI, Google Cloud SDK (gcloud), and Azure CLI (az)",
	Icon:        "[cloud]",
	PackageIDs:  []string{"aws", "gcloud", "azure"},
}

// CloudPackages defines declarative specifications for major cloud provider CLIs
var CloudPackages = []model.Package{
	{
		ID:          "aws",
		Name:        "AWS CLI",
		Description: "Unified tool to manage your AWS cloud services",
		Category:    model.CategoryCloud,
		Type:        model.TypeBrewFormula,
		BrewName:    "awscli",
		AptName:     "awscli",
		WingetID:    "Amazon.AWSCLI",
		ChocoName:   "awscli",
		BinaryName:  "aws",
	},
	{
		ID:          "gcloud",
		Name:        "Google Cloud SDK",
		Description: "Google Cloud CLI (gcloud) tools and libraries",
		Category:    model.CategoryCloud,
		Type:        model.TypeBrewCask,
		BrewName:    "google-cloud-sdk",
		AptName:     "google-cloud-cli",
		WingetID:    "Google.CloudSDK",
		ChocoName:   "gcloudsdk",
		BinaryName:  "gcloud",
		IsCask:      true,
	},
	{
		ID:          "azure",
		Name:        "Azure CLI",
		Description: "Microsoft Azure command-line tools (az)",
		Category:    model.CategoryCloud,
		Type:        model.TypeBrewFormula,
		BrewName:    "azure-cli",
		AptName:     "azure-cli",
		WingetID:    "Microsoft.AzureCLI",
		ChocoName:   "azure-cli",
		BinaryName:  "az",
	},
}
