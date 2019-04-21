# GitHub to Backlog

GitHub Webhooks(push event) -> Google Cloud Functions -> Backlog API(comment)

## How To Use

### Deploy GCF

1. cp config.sample.yaml config.yaml
2. edit config.yaml
3. make deploy PROJECT_NAME=YOUR_GCP_PROJECT_NAME

#### How To Get Backlog API Key

- Open https://YOUR_SPACE.backlog.jp/EditApiSettings.action

### Setup Webhook

1. Open https://github.com/YOUR_ORG/YOUR_REPOSITORY/settings/hooks/new

|Settings|Values|
----|----
|Payload URL|https://*.cloudfunctions.net/Callback|
|Content type|**application/json**|
|Secret||
|SSL verification|Enable SSL verification (default)|
|Which events would you like to trigger this webhook?	|Send me everything.|
|Active|On (default)|

2. Add webhook
