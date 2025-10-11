build:
	go build -o kubespiffed

run: build
	./kubespiffed

test:
	go test ./...


docker: build
	docker build -t kubespiffed .

kind:
	kind delete cluster -n kubespiffe
	kind create cluster -n kubespiffe
	kind load docker-image --name kubespiffe "kubespiffed:latest"
	
	# deploy kubespiffed
	kubectl create ns kubespiffe --context kind-kubespiffe
	kubectl apply -f ./deployment/kubespiffed/deployment.yaml --context kind-kubespiffe
