# Running multiple instances

This project now provides two docker-compose configurations that can be started simultaneously.

- `docker-compose.yaml` – default instance mapping ports `5432`, `7001`, `8001`.
- `docker-compose.second.yaml` – secondary instance mapping ports `5433`, `7002`, `8002`.

Each configuration defines its own containers, network and volume so they do not conflict.

To start both:

```bash
# first instance
docker compose -f docker-compose.yaml up -d

# second instance
docker compose -f docker-compose.second.yaml up -d
```

Each instance has its own PostgreSQL volume (`pg-data` and `pg-data2`), providing isolated databases.
