This repo is about Custom Resource Definition<br>
Controller-gen installation:
```
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest

```
Write ``go run main.go`` to start the controller.<br>
Then forward the port to run the Book API server on `localhost:9090` <br>
Port forwarding command.
```
kubectl port-forward svc/my-book-store-service 9090:3200
```

Here my-book-store-service is the service name. Change it if you have different service name.