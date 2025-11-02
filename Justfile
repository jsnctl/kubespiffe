build:
	GOOS=linux GOARCH=amd64 go build -o kubespiffed

gen verb='':
	#!/usr/bin/env bash
	if [ "{{verb}}" = "verify" ]; then
		./hack/verify-codegen.sh
	else
		./hack/update-codegen.sh
	fi

run: build
	./kubespiffed

test:
	go test ./...


docker: build
	docker build -t kubespiffed .

deploy: docker
	kind load docker-image --name kubespiffe "kubespiffed:latest"
	
	kubectl create ns kubespiffe --context kind-kubespiffe || true
	
	kubectl apply -f ./deployment/kubespiffed/deployment.yaml --context kind-kubespiffe
	kubectl apply -f ./deployment/kubespiffed/service.yaml --context kind-kubespiffe
	kubectl apply -f ./deployment/kubespiffed/rbac.yaml --context kind-kubespiffe
	
	kubectl apply -f ./deployment/workload/deployment.yaml --context kind-kubespiffe
	kubectl apply -f ./deployment/workload/unattested-deployment.yaml --context kind-kubespiffe
	
	kubectl rollout restart deployment -n kubespiffe kubespiffed
	kubectl rollout restart deployment workload

kind:
	kind delete cluster -n kubespiffe
	kind create cluster -n kubespiffe
	just deploy
	
