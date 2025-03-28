# Deployments

This directory contains Infrastructure as Code (IaC) and deployment configurations for the project. It follows the principle of maintaining infrastructure and deployment specifications as version-controlled code.

## Directory Structure

```
deployments/
├── terraform/    # Infrastructure as Code using Terraform
├── helm/         # Kubernetes deployment configurations using Helm
└── README.md     # This file
```

## Terraform Configuration (/terraform)

Contains Terraform configurations for provisioning and managing cloud infrastructure.

### Usage

```bash
cd terraform
terraform init
terraform plan
terraform apply
```

## Helm Charts (/helm)

Contains Kubernetes deployment configurations using Helm charts.

### Usage

```bash
cd helm
# Install the chart
helm install my-release ./chart-name

# Upgrade existing deployment
helm upgrade my-release ./chart-name

# Uninstall the release
helm uninstall my-release
```

## Best Practices

1. Version all configuration files
2. Use variables and templates for environment-specific values
3. Document all configuration parameters
4. Keep sensitive information in secure storage (not in version control)
5. Follow Infrastructure as Code (IaC) principles

## Prerequisites

- Terraform CLI
- Helm CLI
- Access credentials for your cloud provider
- Kubernetes cluster access (for Helm deployments)
