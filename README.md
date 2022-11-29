# go-cli-selfupdate

A toy project of a self updatable CLI application written in Golang with Cobra.

And with a K8S twist of reading information (version information and constraints) from a
Kubernetes ConfigMap. Imagine the scenario is that this CLI is a frontend for an API
built as an extension of the K8S API (using operators and CRDs), so the CLI would
already have to talk with a K8S cluster for other things.
