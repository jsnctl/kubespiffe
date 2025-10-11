build:
	go build -o kubespiffed

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
	kubectl rollout restart deployment -n kubespiffe kubespiffed

kind:
	kind delete cluster -n kubespiffe
	kind create cluster -n kubespiffe
	just deploy
	
