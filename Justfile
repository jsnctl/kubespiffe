build:
	GOOS=linux GOARCH=amd64 go build -o ./kubespiffed cmd/kubespiffe/main.go

gen verb='':
	#!/usr/bin/env bash
	if [ "{{verb}}" = "verify" ]; then
		./hack/verify-codegen.sh
	else
		./hack/update-codegen.sh
	fi

test:
	go test ./...

docker:
	docker build -t kubespiffed .

deploy: docker
	kind load docker-image --name kubespiffe "kubespiffed:latest"
	
	kubectl create ns kubespiffe --context kind-kubespiffe || true
	
	kubectl apply -f ./deployment/kubespiffed/deployment.yaml --context kind-kubespiffe
	kubectl apply -f ./deployment/kubespiffed/service.yaml --context kind-kubespiffe
	kubectl apply -f ./deployment/kubespiffed/rbac.yaml --context kind-kubespiffe
	kubectl apply -f ./deployment/workload-registration/crd.yaml --context kind-kubespiffe
	
	kubectl apply -f ./deployment/workload/deployment.yaml --context kind-kubespiffe
	kubectl apply -f ./deployment/workload/unattested-deployment.yaml --context kind-kubespiffe
	
	kubectl apply -f ./deployment/workload-registration/example.yaml --context kind-kubespiffe

	kubectl rollout restart deployment -n kubespiffe kubespiffed
	kubectl rollout restart deployment workload

kind:
	kind delete cluster -n kubespiffe
	kind create cluster -n kubespiffe
	just deploy
	
