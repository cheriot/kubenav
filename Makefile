
# kind get clusters
# kubectl cluster-info --context <clusterName>
# kind delete cluster --name <clusterName>
cluster-create:
	kind create cluster --name test-cluster --wait 5m
	kubectl create namespace hello

cluster-delete:
	kind delete cluster --name test-cluster

# https://github.com/paulbouwer/hello-kubernetes
blank:
	@kubectl create deployment -n hello --image paulbouwer/hello-kubernetes:1.8 hello -o yaml --dry-run

# kubectl cluster-info dump
# docker exec 14d44c7be6b1 journalctl -xe
events:
	kubectl get events --sort-by=.metadata.creationTimestamp

# See .scalafix.conf and .scalafmt.conf
fix:
	sbt scalafixAll scalafmtAll

run:
	bloop run root -- --log-level debug

sources:
	sbt updateClassifiers

js:
	sbt fastLinkJS