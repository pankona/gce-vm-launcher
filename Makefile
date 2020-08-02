
all:
	@find ./cmd -mindepth 1 -type d | xargs -I{} sh -c 'cd {}; echo "building {}"; go build'

lint:
	golangci-lint run

deploy-cloud-functions:
	gcloud functions deploy vm --runtime go113 --entry-point Command --allow-unauthenticated --trigger-http --memory=128MB --region=asia-northeast1 --env-vars-file env.yaml
	gcloud functions deploy store-status-topic --runtime go113 --entry-point StoreStatus --allow-unauthenticated --trigger-topic gce-status --memory=128MB --region=asia-northeast1 --env-vars-file env.yaml
