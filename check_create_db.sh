#!/bin/bash

DOCKER="/usr/bin/docker"
CONTAINER="postgres"
APP_CONTAINER="app-container"
DB_NAME="db"
DB_USER="postgres"
LOG_FILE="/var/log/check_db.log"
DEBUG=false

# === ПАРСИНГ АРГУМЕНТОВ ===
if [[ "$1" == "--debug" ]]; then
  DEBUG=true
fi

log() {
  local msg="[ $(date +'%Y-%m-%d %H:%M:%S') ] $1"
  echo "$msg"
  echo "$msg" >> "$LOG_FILE"
}

log "=== Запуск скрипта проверки базы данных ==="


EXISTS=$($DOCKER exec -i $CONTAINER psql -U $DB_USER -tAc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME';")


if [[ "$EXISTS" != "1" ]]; then
  log "[`date`] База $DB_NAME не найдена. Создаю..."
  $DOCKER exec -i $CONTAINER psql -U $DB_USER -c "CREATE DATABASE $DB_NAME;"
  
  log "[`date`] Перезапускаю контейнер $APP_CONTAINER..."
  $DOCKER restart $APP_CONTAINER
else
  log "[`date`] База $DB_NAME существует. Ничего не делаю."
fi

if $DEBUG; then
  log
  log "--- DEBUG: текущие базы данных ---"
  $DOCKER exec -i "$CONTAINER" psql -U "$DB_USER" -c "\l"
fi
