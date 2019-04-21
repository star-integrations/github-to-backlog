.DEFAULT_GOAL := help

deploy: ## Deploy Google Cloud Functions
	gcloud functions deploy Callback --project $(PROJECT_NAME) --runtime go111 --trigger-http

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: help
