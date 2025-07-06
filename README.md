# Running multiple instances

This project now provides two docker-compose configurations that can be started simultaneously.

- `docker-compose.yaml` – default instance mapping ports `5432`, `7001`, `8001`.
- `docker-compose.second.yaml` – secondary instance mapping ports `5434`, `7004`, `8002`.

Each configuration defines its own containers, network and volume so they do not conflict.
The second compose file also overrides the application database host so the app
connects to its own `postgres2` service.

Environment variables are loaded from `etc/config.env` by default. You can edit
that file or provide custom values in the compose files. The secondary compose
already sets `PG_HOST` to `postgres2` so it talks to its own database service.

To start both:

```bash
# first instance
docker compose -f docker-compose.yaml up -d

# second instance
docker compose -f docker-compose.second.yaml up -d
```

Each instance has its own PostgreSQL volume (`pg-data` and `pg-data2`), providing isolated databases.
If any of the mapped ports (`5434`, `7004`, `8002`) are in use on your host, edit
`docker-compose.second.yaml` and change them to free ports before starting the
stack.
