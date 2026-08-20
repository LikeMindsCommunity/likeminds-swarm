## Local docker setup

1. Install Docker desktop.

2. If not done already, create a new network named "likeminds-network" using ` docker network create -d bridge likeminds-network`

3. Uncomment the docker envs in `cmd/server/config/.env` file, add more envs if required for running the server locally.

4. `cd` to the root level of this repository.

5. Build the images using `docker compose -f docker-compose-local.yml build --no-cache`

6. Run the containers using `docker compose -f docker-compose-local.yml up`

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

---

## A note on the Jenkins pipelines

The `Jenkinsfile.*` files in this repo are **retained for historical reference and are not
operational**. The Jenkins servers and the AWS and Azure infrastructure they deployed to were
decommissioned in August 2026, so nothing in them runs.

They are kept because they document how this service was built and deployed: the image layout per
environment, the component split, and the deployment topology. Read them as history, not as a build
system you can run.

The GitHub Actions workflows in `.github/workflows`, where present, are the only automation that
still executes.
