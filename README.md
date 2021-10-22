# Blacklister
This application does the following:
* It responds to the URL like 'http://host/?n=x' and returns `n*n`.
* It responds to the URL 'http://host/blacklisted' with conditions:
  * return error code 444 to the visitor
  * block the IP of the visitor
  * send an email with IP address to "test@domain.com"
  * insert into PostgreSQL table information: path, IP address of the visitor and datetime when he
    got blocked

## Dependencies
* Docker
* Go-migrate
* Golang
* Helm
* Kind
* Kubectl
* Make

You can check dependency versions in `.tool-versions` file.

### Install dependencies
Suggested way to install dependencies is to use '[asdf](https://github.com/asdf-vm/asdf)' tool
(except 'docker' and 'make').
If you have the tool installed, run:
```bash
# make sure all relevant plugins are installed
cat .tool-versions |\
  while read p; do
    asdf plugin add ${p%% *}
  done

# install dependencies
asdf install
```

Check docker manual for docker installation instructions:
[get docker](https://docs.docker.com/get-docker/).

## Running
**IMPORTANT:** Make sure you have all dependencies available (see instructions above).

Use the following command to run the application:
```
make all
```

This will perform the following actions:
- Start a local Kubernetes cluster using Kind
- Start a local Docker registry
- Build Docker images and push them to local registry
- Deploy Helm chart

### Clean-up
You can use this helper to remove Helm release and Kind cluster:
```
make delete-all
```

## Using application
1. Calculate the square of a number:
```bash
curl -XPOST http://localhost:8080/?n=2
```
2. Blacklist your IP:
```bash
curl -XGET http://localhost:8080/blacklisted
```
3. Get logs:
```bash
kubectl logs --context kind-dev -n dev-blacklister -l 'app.kubernetes.io/name=blacklister' -f
```

## Development
You can run the application locally and use the database from the local Kind cluster. To do that you
will need to set `DB_HOST` to `localhost` and get the DB credentials from secret:
```bash
export DB_HOST=localhost
export DB_USER="$(kubectl get secret --context kind-dev -n dev-blacklister "blacklister-blacklister-writer-user.ops-dev-blacklister-db.credentials.postgresql.acid.zalan.do" -o go-template='{{.data.username|base64decode}}')"
export DB_PASSWORD="$(kubectl get secret  --context kind-dev -n dev-blacklister "blacklister-blacklister-writer-user.ops-dev-blacklister-db.credentials.postgresql.acid.zalan.do" -o go-template='{{.data.password|base64decode}}')"

# Run the application
./build/blacklister-darwin-amd64 -log-level debug -dev -listen-address "0.0.0.0:8081"
```

Here is a list of helpers to aid development process:
* `make migrate-up` - Run DB migrations
* `make migrate-down` - Rollback last DB migration
* `make migrate-drop` - Drop all migrations
* `make psql` - Login into the database with `psql` (must have `psql` installed)
* `make build-all` - Build code locally
* `make test` - Run tests
* `make docker` - Build docker images and push to local registry
* `make deploy` - Deploy Helm chart
* `make help` - Display help message

You can also access code reference documentation with: `make godoc` - this will start a local godoc
webserver (see link in the command output).

### Versioning
Please maintain application version in `VERSION` file. Use [semantic
versioning](https://semver.org/).
