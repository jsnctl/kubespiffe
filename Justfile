build:
	GOOS=linux GOARCH=amd64 go build -o kubespiffed

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
	kubectl apply -f ./deployment/workload/deployment.yaml --context kind-kubespiffe
	kubectl rollout restart deployment -n kubespiffe kubespiffed
	kubectl rollout restart deployment example-workload

kind:
	kind delete cluster -n kubespiffe
	kind create cluster -n kubespiffe
	just deploy
	
