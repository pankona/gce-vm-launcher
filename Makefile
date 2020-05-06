
all:
	@find ./cmd -mindepth 1 -type d | xargs -I{} sh -c 'cd {}; echo "building {}"; go build'

lint:
	golangci-lint run

deploy-cloud-functions:
	gcloud functions deploy vm --runtime go113 --entry-point Command --allow-unauthenticated --trigger-http --memory=128MB --region=asia-northeast1
