## Local docker setup

1. Install Docker desktop.

2. If not done already, create a new network named "likeminds-network" using ` docker network create -d bridge likeminds-network`

3. Uncomment the docker envs in `cmd/server/config/.env` file, add more envs if required for running the server locally.

4. `cd` to the root level of this repository.

5. Build the images using `docker compose -f docker-compose-local.yml build --no-cache`

6. Run the containers using `docker compose -f docker-compose-local.yml up`
